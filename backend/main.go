package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

type app struct {
	db       *sql.DB
	sessions *sessionStore
}

type sessionStore struct {
	mu       sync.RWMutex
	sessions map[string]int64
}

type user struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

type weightRecord struct {
	ID        int64   `json:"id"`
	UserID    int64   `json:"userId"`
	Username  string  `json:"username"`
	Date      string  `json:"date"`
	Weight    float64 `json:"weight"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt string  `json:"updatedAt"`
}

type authPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type saveWeightPayload struct {
	Date   string  `json:"date"`
	Weight float64 `json:"weight"`
}

type changePasswordPayload struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

type errorResponse struct {
	Message string `json:"message"`
}

type authResponse struct {
	User user `json:"user"`
}

type weightsResponse struct {
	Records []weightRecord `json:"records"`
}

func main() {
	dbPath := filepath.Clean(filepath.Join("..", "data", "daily_weight.db"))
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := migrate(db); err != nil {
		log.Fatal(err)
	}

	application := &app{
		db:       db,
		sessions: newSessionStore(),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", application.handleHealth)
	mux.HandleFunc("/api/register", application.handleRegister)
	mux.HandleFunc("/api/login", application.handleLogin)
	mux.HandleFunc("/api/logout", application.handleLogout)
	mux.HandleFunc("/api/change-password", application.handleChangePassword)
	mux.HandleFunc("/api/me", application.handleMe)
	mux.HandleFunc("/api/weights", application.handleWeights)

	server := &http.Server{
		Addr:    ":8086",
		Handler: withCORS(mux),
	}

	log.Println("backend listening on http://localhost:8086")
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func migrate(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS weight_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			record_date TEXT NOT NULL,
			weight REAL NOT NULL,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, record_date),
			FOREIGN KEY(user_id) REFERENCES users(id)
		);`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

func newSessionStore() *sessionStore {
	return &sessionStore{sessions: make(map[string]int64)}
}

func (s *sessionStore) create(userID int64) (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	token := hex.EncodeToString(buf)
	s.mu.Lock()
	s.sessions[token] = userID
	s.mu.Unlock()
	return token, nil
}

func (s *sessionStore) get(token string) (int64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	userID, ok := s.sessions[token]
	return userID, ok
}

func (s *sessionStore) delete(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, token)
}

func (a *app) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *app) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	payload, err := decodeAuthPayload(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "密码处理失败")
		return
	}

	result, err := a.db.Exec(
		`INSERT INTO users (username, password_hash) VALUES (?, ?)`,
		payload.Username,
		string(hash),
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			writeError(w, http.StatusConflict, "用户名已存在")
			return
		}
		writeError(w, http.StatusInternalServerError, "注册失败")
		return
	}

	userID, err := result.LastInsertId()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "注册失败")
		return
	}

	currentUser := user{ID: userID, Username: payload.Username}
	if err := a.startSession(w, currentUser.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "创建会话失败")
		return
	}

	writeJSON(w, http.StatusCreated, authResponse{User: currentUser})
}

func (a *app) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	payload, err := decodeAuthPayload(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var currentUser user
	var passwordHash string
	err = a.db.QueryRow(
		`SELECT id, username, password_hash FROM users WHERE username = ?`,
		payload.Username,
	).Scan(&currentUser.ID, &currentUser.Username, &passwordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusUnauthorized, "用户名或密码错误")
			return
		}
		writeError(w, http.StatusInternalServerError, "登录失败")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(payload.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "用户名或密码错误")
		return
	}

	if err := a.startSession(w, currentUser.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "创建会话失败")
		return
	}

	writeJSON(w, http.StatusOK, authResponse{User: currentUser})
}

func (a *app) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	currentUser, err := a.requireUser(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "未登录")
		return
	}

	var payload changePasswordPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "请求数据格式错误")
		return
	}

	payload.OldPassword = strings.TrimSpace(payload.OldPassword)
	payload.NewPassword = strings.TrimSpace(payload.NewPassword)
	if payload.OldPassword == "" || payload.NewPassword == "" {
		writeError(w, http.StatusBadRequest, "旧密码和新密码不能为空")
		return
	}

	var passwordHash string
	err = a.db.QueryRow(`SELECT password_hash FROM users WHERE id = ?`, currentUser.ID).Scan(&passwordHash)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "查询用户失败")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(payload.OldPassword)); err != nil {
		writeError(w, http.StatusUnauthorized, "旧密码错误")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(payload.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "密码处理失败")
		return
	}

	if _, err := a.db.Exec(`UPDATE users SET password_hash = ? WHERE id = ?`, string(hash), currentUser.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "更新密码失败")
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (a *app) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}

	cookie, err := r.Cookie("daily_weight_session")
	if err == nil && cookie.Value != "" {
		a.sessions.delete(cookie.Value)
		http.SetCookie(w, &http.Cookie{
			Name:     "daily_weight_session",
			Value:    "",
			Path:     "/",
			Expires:  time.Unix(0, 0),
			MaxAge:   -1,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
	}

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (a *app) handleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}

	currentUser, err := a.requireUser(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "未登录")
		return
	}

	writeJSON(w, http.StatusOK, authResponse{User: currentUser})
}

func (a *app) handleWeights(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.handleListWeights(w, r)
	case http.MethodPost:
		currentUser, err := a.requireUser(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "未登录")
			return
		}
		a.handleSaveWeight(w, r, currentUser)
	default:
		methodNotAllowed(w)
	}
}

func (a *app) handleListWeights(w http.ResponseWriter, r *http.Request) {
	month := strings.TrimSpace(r.URL.Query().Get("month"))
	if _, err := time.Parse("2006-01", month); err != nil {
		writeError(w, http.StatusBadRequest, "月份格式应为 YYYY-MM")
		return
	}

	start, _ := time.Parse("2006-01", month)
	end := start.AddDate(0, 1, 0)

	rows, err := a.db.Query(
		`SELECT wr.id, wr.user_id, u.username, wr.record_date, wr.weight, wr.created_at, wr.updated_at
		 FROM weight_records wr
		 JOIN users u ON u.id = wr.user_id
		 WHERE wr.record_date >= ? AND wr.record_date < ?
		 ORDER BY wr.record_date ASC, u.username ASC`,
		start.Format("2006-01-02"),
		end.Format("2006-01-02"),
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "获取体重记录失败")
		return
	}
	defer rows.Close()

	records := make([]weightRecord, 0)
	for rows.Next() {
		var record weightRecord
		if err := rows.Scan(&record.ID, &record.UserID, &record.Username, &record.Date, &record.Weight, &record.CreatedAt, &record.UpdatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "获取体重记录失败")
			return
		}
		records = append(records, record)
	}

	writeJSON(w, http.StatusOK, weightsResponse{Records: records})
}

func (a *app) handleSaveWeight(w http.ResponseWriter, r *http.Request, currentUser user) {
	var payload saveWeightPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "请求数据格式错误")
		return
	}

	payload.Date = strings.TrimSpace(payload.Date)
	if _, err := time.Parse("2006-01-02", payload.Date); err != nil {
		writeError(w, http.StatusBadRequest, "日期格式应为 YYYY-MM-DD")
		return
	}

	if payload.Weight <= 0 {
		writeError(w, http.StatusBadRequest, "体重必须大于 0")
		return
	}

	_, err := a.db.Exec(
		`INSERT INTO weight_records (user_id, record_date, weight, updated_at)
		 VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		 ON CONFLICT(user_id, record_date)
		 DO UPDATE SET weight = excluded.weight, updated_at = CURRENT_TIMESTAMP`,
		currentUser.ID,
		payload.Date,
		payload.Weight,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "保存体重失败")
		return
	}

	var record weightRecord
	err = a.db.QueryRow(
		`SELECT wr.id, wr.user_id, u.username, wr.record_date, wr.weight, wr.created_at, wr.updated_at
		 FROM weight_records wr
		 JOIN users u ON u.id = wr.user_id
		 WHERE wr.user_id = ? AND wr.record_date = ?`,
		currentUser.ID,
		payload.Date,
	).Scan(&record.ID, &record.UserID, &record.Username, &record.Date, &record.Weight, &record.CreatedAt, &record.UpdatedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读取体重失败")
		return
	}

	writeJSON(w, http.StatusOK, record)
}

func (a *app) requireUser(r *http.Request) (user, error) {
	cookie, err := r.Cookie("daily_weight_session")
	if err != nil || cookie.Value == "" {
		return user{}, errors.New("missing session")
	}

	userID, ok := a.sessions.get(cookie.Value)
	if !ok {
		return user{}, errors.New("invalid session")
	}

	var currentUser user
	err = a.db.QueryRow(`SELECT id, username FROM users WHERE id = ?`, userID).Scan(&currentUser.ID, &currentUser.Username)
	if err != nil {
		return user{}, err
	}

	return currentUser, nil
}

func (a *app) startSession(w http.ResponseWriter, userID int64) error {
	token, err := a.sessions.create(userID)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "daily_weight_session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   60 * 60 * 24 * 7,
	})
	return nil
}

func decodeAuthPayload(r *http.Request) (authPayload, error) {
	var payload authPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return authPayload{}, errors.New("请求数据格式错误")
	}

	payload.Username = strings.TrimSpace(payload.Username)
	payload.Password = strings.TrimSpace(payload.Password)
	if payload.Username == "" || payload.Password == "" {
		return authPayload{}, errors.New("用户名和密码不能为空")
	}

	if len([]rune(payload.Username)) > 32 {
		return authPayload{}, errors.New("用户名长度不能超过 32 个字符")
	}

	return payload, nil
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowed := false
		if strings.HasPrefix(origin, "http://127.0.0.1:970") ||
			strings.HasPrefix(origin, "http://localhost:970") ||
			strings.HasPrefix(origin, "http://192.168.") && strings.HasSuffix(origin, ":970") ||
			strings.HasPrefix(origin, "http://10.") && strings.HasSuffix(origin, ":970") ||
			strings.HasPrefix(origin, "http://172.16.") && strings.HasSuffix(origin, ":970") {
			allowed = true
		}
		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode response: %v", err), http.StatusInternalServerError)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Message: message})
}

func methodNotAllowed(w http.ResponseWriter) {
	writeError(w, http.StatusMethodNotAllowed, "不支持的请求方法")
}
