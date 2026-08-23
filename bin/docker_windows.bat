@echo off
setlocal
set "SCRIPT_DIR=%~dp0"
set "CYBERMES_DIR=%SCRIPT_DIR%..\"
cd /d "%CYBERMES_DIR%"

if "%1"=="" goto help
if "%1"=="up" goto up
if "%1"=="down" goto down
if "%1"=="logs" goto logs
if "%1"=="exec" goto exec
if "%1"=="build" goto build

:help
echo Cybermes Docker Windows Helper
echo.
echo Usage:
echo   bin\docker_windows.bat up        - Start containers in background
echo   bin\docker_windows.bat down      - Stop containers
echo   bin\docker_windows.bat logs      - View logs
echo   bin\docker_windows.bat exec ...  - Run command in container
echo   bin\docker_windows.bat build     - Rebuild images
exit /b 0

:up
docker compose up -d
exit /b %ERRORLEVEL%

:down
docker compose down
exit /b %ERRORLEVEL%

:logs
docker compose logs -f
exit /b %ERRORLEVEL%

:exec
shift
docker compose exec hermes-cybermes %*
exit /b %ERRORLEVEL%

:build
docker compose build --no-cache
exit /b %ERRORLEVEL%
