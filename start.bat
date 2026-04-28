@echo off
chcp 65001 >nul 2>&1
setlocal enabledelayedexpansion

set "ROOT_DIR=%~dp0"
set "ROOT_DIR=%ROOT_DIR:~0,-1%"
set "FRONTEND_DIR=%ROOT_DIR%\frontend"
set "BACKEND_DIR=%ROOT_DIR%\backend"
set "BACKEND_LOG=%ROOT_DIR%\.dailyweight-backend.log"
set "BACKEND_BIN=%ROOT_DIR%\.dailyweight-backend.exe"
set "BACKEND_PID="
set "FRONTEND_PID="

:: 检查 go
where go >nul 2>&1
if errorlevel 1 (
    echo 未找到 go，请先安装 Go，或将 Go 添加到 PATH。
    exit /b 1
)

:: 检查端口占用
call :check_port 8086 "后端"
call :check_port 970 "前端"

:: 检查 pnpm
where pnpm >nul 2>&1
if errorlevel 1 (
    echo 未找到 pnpm，请先安装 pnpm，或将 pnpm 添加到 PATH。
    call :cleanup
    exit /b 1
)

:: 检查 curl
where curl >nul 2>&1
if errorlevel 1 (
    echo 未找到 curl，请先安装 curl，或将 curl 添加到 PATH。
    call :cleanup
    exit /b 1
)

:: 清理例程
goto :main

:check_port
set "port=%~1"
set "name=%~2"
netstat -ano | findstr /r ":%port% .*LISTENING" >nul 2>&1
if not errorlevel 1 (
    echo %name% 端口 %port% 已被占用，请先停止已运行的进程后再执行 start.bat。
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr /r ":%port% .*LISTENING"') do (
        echo   PID: %%a
    )
    exit /b 1
)
exit /b 0

:cleanup
echo.
echo 正在停止前后端...
if defined BACKEND_PID (
    taskkill /pid !BACKEND_PID! /f /t >nul 2>&1
)
if defined FRONTEND_PID (
    taskkill /pid !FRONTEND_PID! /f /t >nul 2>&1
)
if exist "%BACKEND_BIN%" (
    del /f "%BACKEND_BIN%" >nul 2>&1
)
echo 已停止。
exit /b 0

:main

:: 编译并启动后端
echo 正在编译后端...
cd /d "%BACKEND_DIR%"
go build -buildvcs=false -o "%BACKEND_BIN%" . >"%BACKEND_LOG%" 2>&1
if errorlevel 1 (
    echo 后端编译失败，请检查日志：%BACKEND_LOG%
    type "%BACKEND_LOG%"
    exit /b 1
)

echo 正在启动后端...
start /b "" "%BACKEND_BIN%" >>"%BACKEND_LOG%" 2>&1
:: 获取刚启动进程的 PID（通过端口反查有延迟，先用 timeout 等一小段再查）
timeout /t 2 /nobreak >nul

:: 等待后端就绪（最多 60 秒）
set "ready=0"
for /l %%i in (1,1,60) do (
    curl -fsS http://127.0.0.1:8086/api/health >nul 2>&1
    if not errorlevel 1 (
        set "ready=1"
        goto :backend_ready
    )
    :: 检查后端进程是否还活着（通过端口是否被占用来判断）
    netstat -ano | findstr /r ":8086 .*LISTENING" >nul 2>&1
    if errorlevel 1 (
        if %%i gtr 3 (
            echo 后端启动失败，请检查日志：%BACKEND_LOG%
            type "%BACKEND_LOG%"
            call :cleanup
            exit /b 1
        )
    )
    timeout /t 1 /nobreak >nul
)

:backend_ready
if "!ready!"=="0" (
    echo 后端启动超时，请检查日志：%BACKEND_LOG%
    type "%BACKEND_LOG%"
    call :cleanup
    exit /b 1
)

:: 获取后端 PID
for /f "tokens=5" %%a in ('netstat -ano ^| findstr /r ":8086 .*LISTENING"') do (
    set "BACKEND_PID=%%a"
    goto :got_backend_pid
)
:got_backend_pid

:: 启动前端
echo 正在启动前端...
cd /d "%FRONTEND_DIR%"
start /b "" cmd /c "pnpm dev --host 0.0.0.0 --port 970 --strictPort" >nul 2>&1

:: 等待前端就绪（最多 30 秒）
set "frontend_ready=0"
for /l %%i in (1,1,30) do (
    netstat -ano | findstr /r ":970 .*LISTENING" >nul 2>&1
    if not errorlevel 1 (
        set "frontend_ready=1"
        goto :frontend_ready
    )
    timeout /t 1 /nobreak >nul
)

:frontend_ready
if "!frontend_ready!"=="0" (
    echo 前端启动超时，请检查 970 端口或前端日志。
    call :cleanup
    exit /b 1
)

:: 获取前端 PID
for /f "tokens=5" %%a in ('netstat -ano ^| findstr /r ":970 .*LISTENING"') do (
    set "FRONTEND_PID=%%a"
    goto :got_frontend_pid
)
:got_frontend_pid

echo.
echo frontend: http://localhost:970
echo backend:  http://localhost:8086
echo 按 Ctrl+C 可同时停止前后端。
echo.

:: 保持脚本运行，监控进程存活
:watchdog
timeout /t 2 /nobreak >nul
:: 检查后端是否还活着
if defined BACKEND_PID (
    tasklist /fi "pid eq !BACKEND_PID!" /nh 2>nul | findstr /i "!BACKEND_PID!" >nul 2>&1
    if errorlevel 1 (
        echo 后端进程已退出。
        call :cleanup
        exit /b 1
    )
)
:: 检查前端是否还活着（pnpm 会启动子进程，用端口判断更可靠）
netstat -ano | findstr /r ":970 .*LISTENING" >nul 2>&1
if errorlevel 1 (
    echo 前端进程已退出。
    call :cleanup
    exit /b 1
)
goto watchdog
