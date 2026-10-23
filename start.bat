@echo off
setlocal
cd /d "%~dp0"
title AI Novel Studio v3.0

if not exist "configs\config.yaml" (
    if exist "configs\config.yaml.example" (
        copy /Y "configs\config.yaml.example" "configs\config.yaml" >nul
        echo [init] Created configs\config.yaml from template.
    )
)

if not exist "server.exe" (
    echo [build] server.exe not found, trying to build from source...
    where go >nul 2>&1
    if errorlevel 1 (
        echo [error] server.exe is missing and Go is not installed.
        echo         Please use a release package that includes server.exe.
        pause
        exit /b 1
    )
    go build -trimpath -ldflags="-s -w" -o server.exe ./cmd/server
    if errorlevel 1 (
        echo [error] Build failed.
        pause
        exit /b 1
    )
)

rem If 8082 is already listening, verify it is this v3 instance.
netstat -ano | findstr ":8082" | findstr "LISTENING" >nul 2>&1
if not errorlevel 1 (
    where curl >nul 2>&1
    if errorlevel 1 (
        echo [info] Port 8082 is in use, opening the running instance...
        if not "%NO_BROWSER%"=="1" start "" "http://127.0.0.1:8082/"
        exit /b 0
    )
    curl -s --max-time 3 "http://127.0.0.1:8082/api/version" | findstr /C:"3.0.0" >nul
    if not errorlevel 1 (
        echo [info] AI Novel Studio v3.0 is already running.
        if not "%NO_BROWSER%"=="1" start "" "http://127.0.0.1:8082/"
        exit /b 0
    )
    echo [warn] Port 8082 is occupied by an old/unknown instance.
    echo [warn] Stopping it and starting AI Novel Studio v3.0...
    powershell -NoProfile -Command "$c = Get-NetTCPConnection -LocalPort 8082 -State Listen -ErrorAction SilentlyContinue | Select-Object -First 1; if ($c) { $p = Get-Process -Id $c.OwningProcess -ErrorAction SilentlyContinue; if ($p -and $p.ProcessName -eq 'server') { Stop-Process -Id $p.Id -Force; Write-Host ('[stop] pid ' + $p.Id) } else { Write-Host '[warn] port 8082 is held by a non-server process'; exit 2 } }"
    if errorlevel 2 (
        echo [error] Port 8082 is held by another program. Close it or change server.port.
        pause
        exit /b 1
    )
    timeout /t 2 /nobreak >nul
)

echo [start] Launching AI Novel Studio...
for /f "usebackq delims=" %%i in (`powershell -NoProfile -Command "(Start-Process -FilePath '%~dp0server.exe' -WorkingDirectory '%~dp0' -WindowStyle Hidden -PassThru).Id"`) do set "SERVER_PID=%%i"
if defined SERVER_PID echo %SERVER_PID%> "%~dp0server.pid"

set /a tries=0
:waitloop
set /a tries+=1
if %tries% gtr 30 goto ready_anyway
curl -s -o nul "http://127.0.0.1:8082/api/auth/check" >nul 2>&1
if not errorlevel 1 goto ready
timeout /t 1 /nobreak >nul
goto waitloop

:ready
echo [ready] Server is up: http://127.0.0.1:8082/
if not "%NO_BROWSER%"=="1" start "" "http://127.0.0.1:8082/"
goto end

:ready_anyway
echo [warn] Server did not answer in 30s, opening the page anyway.
if not "%NO_BROWSER%"=="1" start "" "http://127.0.0.1:8082/"

:end
echo.
echo Tip: run stop.bat to stop this instance.
exit /b 0
