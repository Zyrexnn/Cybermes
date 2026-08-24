@echo off
setlocal
set "CYBERMES_DIR=%~dp0"
set "HERMES_HOME=%CYBERMES_DIR%.hermes"
set "PATH=%CYBERMES_DIR%tools\bin;%CYBERMES_DIR%bin;%CYBERMES_DIR%venv\Scripts;%PATH%"

if exist "%CYBERMES_DIR%venv\Scripts\hermes.exe" (
    "%CYBERMES_DIR%venv\Scripts\hermes.exe" %*
) else (
    echo [!] hermes.exe not found. Please run .\setup_windows.ps1 first.
    exit /b 1
)
