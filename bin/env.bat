@echo off
set "SCRIPT_DIR=%~dp0"
set "CYBERMES_DIR=%SCRIPT_DIR%..\"
set "HERMES_HOME=%CYBERMES_DIR%.hermes"
set "PATH=%CYBERMES_DIR%tools\bin;%CYBERMES_DIR%bin;%CYBERMES_DIR%venv\Scripts;%PATH%"

if "%1"=="status" (
    echo [+] CYBERMES_DIR: %CYBERMES_DIR%
    echo [+] HERMES_HOME:  %HERMES_HOME%
    python "%CYBERMES_DIR%tools\windows_compat_check.py"
    exit /b 0
)

echo [+] Cybermes environment variables set.
if exist "%CYBERMES_DIR%venv\Scripts\activate.bat" (
    call "%CYBERMES_DIR%venv\Scripts\activate.bat"
    echo [+] Python virtualenv activated.
)
