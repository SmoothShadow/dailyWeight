#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FRONTEND_DIR="$ROOT_DIR/frontend"
BACKEND_DIR="$ROOT_DIR/backend"
BACKEND_LOG="$ROOT_DIR/.dailyweight-backend.log"

# 检测操作系统
if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "win32" || "$OSTYPE" == "cygwin" ]]; then
    IS_WINDOWS=1
    BACKEND_BIN="$ROOT_DIR/.dailyweight-backend.exe"
else
    IS_WINDOWS=0
    BACKEND_BIN="$ROOT_DIR/.dailyweight-backend"
fi

GO_BIN="${GO_BIN:-}"
if [ -z "$GO_BIN" ]; then
    if command -v go >/dev/null 2>&1; then
        GO_BIN="$(command -v go)"
    elif [ -x "/usr/local/go/bin/go" ]; then
        GO_BIN="/usr/local/go/bin/go"
    else
        echo "未找到 go，请先安装 Go，或通过 GO_BIN 指定可执行文件路径。" >&2
        exit 1
    fi
fi

# 跨平台端口检测函数
is_port_listening() {
    local port="$1"
    if [ "$IS_WINDOWS" -eq 1 ]; then
        netstat -ano 2>/dev/null | grep -E ":${port} .*LISTENING" >/dev/null 2>&1
    else
        lsof -nP -iTCP:"$port" -sTCP:LISTEN >/dev/null 2>&1
    fi
}

require_port_free() {
    local port="$1"
    local name="$2"

    if is_port_listening "$port"; then
        echo "$name 端口 $port 已被占用，请先停止已运行的进程后再执行 start.sh。" >&2
        if [ "$IS_WINDOWS" -eq 1 ]; then
            netstat -ano | grep -E ":${port} .*LISTENING" >&2 || true
        else
            lsof -nP -iTCP:"$port" -sTCP:LISTEN >&2 || true
        fi
        exit 1
    fi
}

# 跨平台进程存活检测
is_pid_alive() {
    local pid="$1"
    if [ "$IS_WINDOWS" -eq 1 ]; then
        tasklist /fi "pid eq $pid" /nh 2>/dev/null | grep -i "$pid" >/dev/null 2>&1
    else
        kill -0 "$pid" >/dev/null 2>&1
    fi
}

cleanup() {
    trap - EXIT INT TERM

    if [ -n "${BACKEND_PID:-}" ] && is_pid_alive "$BACKEND_PID"; then
        if [ "$IS_WINDOWS" -eq 1 ]; then
            taskkill //pid "$BACKEND_PID" //f //t >/dev/null 2>&1 || true
        else
            kill "$BACKEND_PID" >/dev/null 2>&1 || true
        fi
    fi

    if [ -n "${FRONTEND_PID:-}" ] && is_pid_alive "$FRONTEND_PID"; then
        if [ "$IS_WINDOWS" -eq 1 ]; then
            taskkill //pid "$FRONTEND_PID" //f //t >/dev/null 2>&1 || true
        else
            kill "$FRONTEND_PID" >/dev/null 2>&1 || true
        fi
    fi

    if [ -f "$BACKEND_BIN" ]; then
        rm -f "$BACKEND_BIN"
    fi
}
trap cleanup EXIT INT TERM

require_port_free 8086 "后端"
require_port_free 970 "前端"

# 编译并启动后端
(
    cd "$BACKEND_DIR"
    "$GO_BIN" build -buildvcs=false -o "$BACKEND_BIN" .
    exec "$BACKEND_BIN"
) >"$BACKEND_LOG" 2>&1 &
BACKEND_PID=$!

# 等待后端就绪（最多 60 秒）
for _ in {1..60}; do
    if curl -fsS "http://127.0.0.1:8086/api/health" >/dev/null 2>&1; then
        break
    fi

    if ! is_pid_alive "$BACKEND_PID"; then
        wait "$BACKEND_PID" || true
        echo "后端启动失败，请检查日志：$BACKEND_LOG" >&2
        cat "$BACKEND_LOG" >&2 || true
        exit 1
    fi

    sleep 1
done

if ! curl -fsS "http://127.0.0.1:8086/api/health" >/dev/null 2>&1; then
    echo "后端启动超时，请检查日志：$BACKEND_LOG" >&2
    cat "$BACKEND_LOG" >&2 || true
    exit 1
fi

# 启动前端
(
    cd "$FRONTEND_DIR"
    exec pnpm dev --host 0.0.0.0 --port 970 --strictPort
) &
FRONTEND_PID=$!

# 等待前端就绪（最多 30 秒）
for _ in {1..30}; do
    if is_port_listening 970; then
        break
    fi

    if ! is_pid_alive "$FRONTEND_PID"; then
        wait "$FRONTEND_PID" || true
        echo "前端启动失败。" >&2
        exit 1
    fi

    sleep 1
done

if ! is_port_listening 970; then
    echo "前端启动超时，请检查 970 端口或前端日志。" >&2
    exit 1
fi

echo "frontend: http://localhost:970"
echo "backend: http://localhost:8086"
echo "按 Ctrl+C 可同时停止前后端。"

while is_pid_alive "$BACKEND_PID" && is_pid_alive "$FRONTEND_PID"; do
    sleep 1
done

exit 1
