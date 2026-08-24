<#
=============================================================================
Cybermes Windows Automated Setup & Installer
Prepares local Windows environment, Python venv, dependencies & MCP tools
=============================================================================
#>

[CmdletBinding()]
param()

$ErrorActionPreference = "Stop"

$CYBERMES_DIR = $PSScriptRoot
Set-Location -Path $CYBERMES_DIR

Write-Host "========================================================" -ForegroundColor Cyan
Write-Host "  🛡️  Cybermes Windows Host Installation & Setup" -ForegroundColor Cyan
Write-Host "========================================================" -ForegroundColor Cyan
Write-Host "Directory: $CYBERMES_DIR" -ForegroundColor Gray
Write-Host ""

# 1. Check Python 3.10+
$pythonExe = $null
$pythonArgs = @()
if (Get-Command python -ErrorAction SilentlyContinue) {
    $pythonExe = "python"
} elseif (Get-Command py -ErrorAction SilentlyContinue) {
    $pythonExe = "py"
    $pythonArgs = @("-3")
}

if (-not $pythonExe) {
    Write-Host "❌ Error: Python is not installed or not in PATH." -ForegroundColor Red
    Write-Host "Please install Python 3.11+ from https://www.python.org/ and check 'Add Python to PATH'." -ForegroundColor Yellow
    exit 1
}

$pythonCommandDisplay = (@($pythonExe) + $pythonArgs) -join " "
$pyVer = & $pythonExe @pythonArgs -c "import sys; print(f'{sys.version_info.major}.{sys.version_info.minor}')"
Write-Host "✓ Found Python $pyVer ($pythonCommandDisplay)" -ForegroundColor Green

# 2. Check Node.js / npm (Optional for MCP servers)
if ((Get-Command node -ErrorAction SilentlyContinue) -and (Get-Command npm -ErrorAction SilentlyContinue)) {
    $nodeVer = & node --version
    $npmVer = & npm --version
    Write-Host "✓ Found Node.js $nodeVer & npm $npmVer" -ForegroundColor Green
    Write-Host "  Installing MCP servers (Puppeteer & Filesystem)..." -ForegroundColor Gray
    try {
        & npm install --prefix "$CYBERMES_DIR" @modelcontextprotocol/server-puppeteer @modelcontextprotocol/server-filesystem 2>$null
    } catch {
        Write-Host "  [!] Could not install MCP packages locally via npm; continuing..." -ForegroundColor DarkYellow
    }
} else {
    Write-Host "⚠️  Node.js/npm not found. Browser MCP automation will be disabled unless Node is installed." -ForegroundColor Yellow
}

# 3. Create Python Virtual Environment
Write-Host ""
Write-Host "📦 Setting up Python virtual environment (venv)..." -ForegroundColor Cyan
if (-not (Test-Path "$CYBERMES_DIR\venv")) {
    & $pythonExe @pythonArgs -m venv "$CYBERMES_DIR\venv"
}

$venvPython = "$CYBERMES_DIR\venv\Scripts\python.exe"
$venvPip = "$CYBERMES_DIR\venv\Scripts\pip.exe"

if (-not (Test-Path $venvPython)) {
    Write-Host "❌ Failed to create virtual environment." -ForegroundColor Red
    exit 1
}

Write-Host "✓ Virtual environment created" -ForegroundColor Green
Write-Host "📥 Upgrading pip..." -ForegroundColor Gray
& $venvPython -m pip install --upgrade pip --quiet

Write-Host "📥 Installing Python dependencies from requirements.txt..." -ForegroundColor Cyan
& $venvPip install -r "$CYBERMES_DIR\requirements.txt" --quiet

# 4. Create standard workspace directories
Write-Host ""
Write-Host "📁 Initializing workspace directory structure..." -ForegroundColor Cyan
$dirs = @("reports", "recon", "output", "logs", "targets", "tools\bin", ".hermes\skills", "bin")
foreach ($d in $dirs) {
    $targetPath = Join-Path $CYBERMES_DIR $d
    if (-not (Test-Path $targetPath)) {
        New-Item -ItemType Directory -Path $targetPath -Force | Out-Null
    }
}
Write-Host "✓ Workspace directories ready" -ForegroundColor Green

# 5. Copy skills to Hermes home if available
if (Test-Path "$CYBERMES_DIR\skills") {
    Copy-Item -Path "$CYBERMES_DIR\skills\*" -Destination "$CYBERMES_DIR\.hermes\skills\" -Recurse -Force -ErrorAction SilentlyContinue
    Write-Host "✓ Skills synchronized to .hermes/skills/" -ForegroundColor Green
}

# 6. Setup .env and .hermes/config.yaml from examples
if ((-not (Test-Path "$CYBERMES_DIR\.env")) -and (Test-Path "$CYBERMES_DIR\.env.example")) {
    Copy-Item -Path "$CYBERMES_DIR\.env.example" -Destination "$CYBERMES_DIR\.env"
    Write-Host "✓ Generated default .env file from .env.example" -ForegroundColor Green
}

if ((-not (Test-Path "$CYBERMES_DIR\.hermes\config.yaml")) -and (Test-Path "$CYBERMES_DIR\.hermes\config.yaml.example")) {
    Copy-Item -Path "$CYBERMES_DIR\.hermes\config.yaml.example" -Destination "$CYBERMES_DIR\.hermes\config.yaml"
    Write-Host "✓ Initialized .hermes/config.yaml from .hermes/config.yaml.example" -ForegroundColor Green
}

$authJsonPath = "$CYBERMES_DIR\.hermes\auth.json"
if (-not (Test-Path $authJsonPath)) {
    "{}" | Out-File -FilePath $authJsonPath -Encoding utf8
    Write-Host "✓ Initialized empty .hermes/auth.json" -ForegroundColor Green
}

# 7. Compile Go Core Tools (if Go is available on Windows)
if (Get-Command go -ErrorAction SilentlyContinue) {
    Write-Host "⚡ Compiling Cybermes High-Performance Go Tools into tools\bin\..." -ForegroundColor Cyan
    try {
        & go build -ldflags="-s -w" -o "$CYBERMES_DIR\tools\bin\smart_pipe.exe" "$CYBERMES_DIR\cmd\smart_pipe"
        & go build -ldflags="-s -w" -o "$CYBERMES_DIR\tools\bin\secret_scan.exe" "$CYBERMES_DIR\cmd\secret_scan"
        & go build -ldflags="-s -w" -o "$CYBERMES_DIR\tools\bin\search_knowledge.exe" "$CYBERMES_DIR\cmd\search_knowledge"
        & go build -ldflags="-s -w" -o "$CYBERMES_DIR\tools\bin\aggregate_reports.exe" "$CYBERMES_DIR\cmd\aggregate_reports"
        Write-Host "✓ Built smart_pipe.exe, secret_scan.exe, search_knowledge.exe, aggregate_reports.exe" -ForegroundColor Green
    } catch {
        Write-Host "  [!] Note: Could not build Go binaries automatically on Windows; continuing..." -ForegroundColor DarkYellow
    }
}

# 8. Configure dynamically in .hermes/config.yaml & sync .env using Python
& $venvPython -c @"
import os
import re
from pathlib import Path

cybermes_dir = Path(r'''$CYBERMES_DIR''').as_posix()
config_path = Path(r'''$CYBERMES_DIR''') / '.hermes/config.yaml'
env_path = Path(r'''$CYBERMES_DIR''') / '.env'

env_vars = {}
if env_path.exists():
    for line in env_path.read_text(encoding='utf-8', errors='ignore').splitlines():
        line = line.strip()
        if line and not line.startswith('#') and '=' in line:
            k, v = line.split('=', 1)
            env_vars[k.strip()] = v.strip().strip('\'"')

api_key = env_vars.get('ROUTER_API_KEY') or env_vars.get('OPENROUTER_API_KEY') or 'your_api_key_here'
base_url = env_vars.get('ROUTER_BASE_URL') or env_vars.get('OPENROUTER_BASE_URL') or 'http://localhost:20128/v1'
model_name = env_vars.get('HERMES_DEFAULT_MODEL') or 'hermes'

if config_path.exists():
    content = config_path.read_text(encoding='utf-8', errors='ignore')
    content = re.sub(r'root:\s*.*', f'root: {cybermes_dir}', content)
    content = re.sub(r'targets:\s*.*', f'targets: {cybermes_dir}/targets', content)
    content = re.sub(r'recon:\s*.*', f'recon: {cybermes_dir}/recon', content)
    content = re.sub(r'output:\s*.*', f'output: {cybermes_dir}/output', content)
    content = re.sub(r'reports:\s*.*', f'reports: {cybermes_dir}/reports', content)
    content = re.sub(r'logs:\s*.*', f'logs: {cybermes_dir}/logs', content)
    content = re.sub(r'knowledge:\s*.*', f'knowledge: {cybermes_dir}/knowledge', content)
    content = re.sub(r'wordlists:\s*.*', f'wordlists: {cybermes_dir}/tools/wordlists', content)
    content = re.sub(r'directory:\s*.*skills', f'directory: {cybermes_dir}/skills', content)

    content = re.sub(r'default:\s*.*', f'default: {model_name}', content)
    if api_key and api_key != 'your_api_key_here':
        content = re.sub(r'api_key:\s*.*', f'api_key: {api_key}', content)
    if base_url:
        content = re.sub(r'base_url:\s*.*', f'base_url: {base_url}', content)

    config_path.write_text(content, encoding='utf-8')

hermes_env = Path(r'''$CYBERMES_DIR''') / '.hermes/.env'
if env_path.exists():
    hermes_env.write_text(env_path.read_text(encoding='utf-8', errors='ignore'), encoding='utf-8')
"@

Write-Host ""
Write-Host "========================================================" -ForegroundColor Green
Write-Host "  ✅ Cybermes Windows Setup Complete!" -ForegroundColor Green
Write-Host "========================================================" -ForegroundColor Green
Write-Host ""
Write-Host "To start using Cybermes on Windows:" -ForegroundColor Cyan
Write-Host ""
Write-Host "  1. Activate environment in current PowerShell session:"
Write-Host "     . .\env.ps1" -ForegroundColor Yellow
Write-Host ""
Write-Host "  2. Edit your .env file with your API keys:"
Write-Host "     notepad .env" -ForegroundColor Yellow
Write-Host ""
Write-Host "  3. Run Cybermes CLI or run system check:"
Write-Host "     .\hermes.bat `"Assess http://127.0.0.1:8888`"" -ForegroundColor Yellow
Write-Host "     python tools\windows_compat_check.py" -ForegroundColor Yellow
Write-Host ""
