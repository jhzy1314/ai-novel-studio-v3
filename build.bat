@echo off
setlocal
cd /d "%~dp0"
title AI Novel Studio v3.0 - build

where go >nul 2>&1
if errorlevel 1 (
    echo [error] Go is not installed or not on PATH.
    pause
    exit /b 1
)

echo [build] go build -trimpath -ldflags="-s -w" -o server.exe ./cmd/server
go build -trimpath -ldflags="-s -w" -o server.exe ./cmd/server
if errorlevel 1 (
    echo [error] Build failed.
    pause
    exit /b 1
)
echo [ok] server.exe updated.
pause
