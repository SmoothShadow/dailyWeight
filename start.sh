#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FRONTEND_DIR="$ROOT_DIR/frontend"
BACKEND_DIR="$ROOT_DIR/backend"
BACKEND_LOG="$ROOT_DIR/.dailyweight-backend.log"
BACKEND_BIN="$ROOT_DIR/.dailyweight-backend"

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

require_port_free() {
  local port="$1"
  local name="$2"

  if lsof -nP -iTCP:"$port" -sTCP:LISTEN >/dev/null 2>&1; then
    echo "$name 端口 $port 已被占用，请先停止已运行的进程后再执行 start.sh。" >&2
    lsof -nP -iTCP:"$port" -sTCP:LISTEN >&2 || true
    exit 1
  fi
}

cleanup() {
  trap - EXIT INT TERM

  if [ -n "${BACKEND_PID:-}" ] && kill -0 "$BACKEND_PID" >/dev/null 2>&1; then
    kill "$BACKEND_PID" >/dev/null 2>&1 || true
  fi

  if [ -n "${FRONTEND_PID:-}" ] && kill -0 "$FRONTEND_PID" >/dev/null 2>&1; then
    kill "$FRONTEND_PID" >/dev/null 2>&1 || true
  fi

  if [ -f "$BACKEND_BIN" ]; then
    rm -f "$BACKEND_BIN"
  fi
}
trap cleanup EXIT INT TERM

require_port_free 8086 "后端"
require_port_free 970 "前端"

(
  cd "$BACKEND_DIR"
  "$GO_BIN" build -buildvcs=false -o "$BACKEND_BIN" .
  exec "$BACKEND_BIN"
) >"$BACKEND_LOG" 2>&1 &
BACKEND_PID=$!

for _ in {1..60}; do
  if curl -4 -fsS "http://127.0.0.1:8086/api/health" >/dev/null 2>&1; then
    break
  fi

  if ! kill -0 "$BACKEND_PID" >/dev/null 2>&1; then
    wait "$BACKEND_PID" || true
    echo "后端启动失败，请检查日志：$BACKEND_LOG" >&2
    cat "$BACKEND_LOG" >&2 || true
    exit 1
  fi

  sleep 1
done

if ! curl -4 -fsS "http://127.0.0.1:8086/api/health" >/dev/null 2>&1; then
  echo "后端启动超时，请检查日志：$BACKEND_LOG" >&2
  cat "$BACKEND_LOG" >&2 || true
  exit 1
fi

(
  cd "$FRONTEND_DIR"
  exec pnpm dev --host 0.0.0.0 --port 970 --strictPort
) &
FRONTEND_PID=$!

for _ in {1..30}; do
  if lsof -nP -iTCP:970 -sTCP:LISTEN >/dev/null 2>&1; then
    break
  fi

  if ! kill -0 "$FRONTEND_PID" >/dev/null 2>&1; then
    wait "$FRONTEND_PID"
    exit 1
  fi

  sleep 1
done

if ! lsof -nP -iTCP:970 -sTCP:LISTEN >/dev/null 2>&1; then
  echo "前端启动超时，请检查 970 端口或前端日志。" >&2
  exit 1
fi

echo "frontend: http://localhost:970"
echo "backend:  http://localhost:8086"
echo "按 Ctrl+C 可同时停止前后端。"

while kill -0 "$BACKEND_PID" >/dev/null 2>&1 && kill -0 "$FRONTEND_PID" >/dev/null 2>&1; do
  sleep 1
done

exit 1
