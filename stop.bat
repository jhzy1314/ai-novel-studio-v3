@echo off
setlocal
cd /d "%~dp0"
title AI Novel Studio v3.0 - stop

set "SERVER_PID="
if exist "server.pid" (
    set /p SERVER_PID=<"server.pid"
)

if defined SERVER_PID (
    powershell -NoProfile -Command "$p = Get-Process -Id %SERVER_PID% -ErrorAction SilentlyContinue; if ($p -and $p.ProcessName -eq 'server') { Stop-Process -Id %SERVER_PID% -Force; Write-Host '[stop] stopped pid %SERVER_PID%' } else { Write-Host '[info] pid %SERVER_PID% is not a running server' }"
    del /Q "server.pid" >nul 2>&1
) else (
    echo [info] no server.pid found; nothing to stop
)

exit /b 0
