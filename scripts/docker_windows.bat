@echo off
setlocal
set "SCRIPT_DIR=%~dp0..\"
cd /d "%SCRIPT_DIR%"

if "%1"=="" goto help
if "%1"=="up" goto up
if "%1"=="down" goto down
if "%1"=="logs" goto logs
if "%1"=="shell" goto shell

:up
docker compose up -d
goto end

:down
docker compose down
goto end

:logs
docker compose logs -f
goto end

:shell
docker compose exec cybermes bash
goto end

:help
echo Cybermes Docker Helper
echo Usage: %0 [up^|down^|logs^|shell]
goto end

:end
