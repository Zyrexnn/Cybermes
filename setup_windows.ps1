[CmdletBinding()]
param()

$ErrorActionPreference = "Stop"
$CYBERMES_DIR = $PSScriptRoot
Set-Location -Path $CYBERMES_DIR

Write-Host "========================================================" -ForegroundColor Cyan
Write-Host "  [+] Cybermes Windows Automated Setup & Installer" -ForegroundColor Cyan
Write-Host "========================================================" -ForegroundColor Cyan
Write-Host "Directory: $CYBERMES_DIR`n" -ForegroundColor Gray

# 1. Python check
$pythonExe = $null
$pythonArgs = @()
if (Get-Command python -ErrorAction SilentlyContinue) {
    $pythonExe = "python"
} elseif (Get-Command py -ErrorAction SilentlyContinue) {
    $pythonExe = "py"
    $pythonArgs = @("-3")
}

if (-not $pythonExe) {
    Write-Host "[-] Error: Python is not installed or not in PATH." -ForegroundColor Red
    Write-Host "Please install Python 3.11+ from https://www.python.org/ and check 'Add Python to PATH'." -ForegroundColor Yellow
    exit 1
}

$pyVer = & $pythonExe @pythonArgs -c "import sys; print(f'{sys.version_info.major}.{sys.version_info.minor}')"
Write-Host "[+] Found Python $pyVer" -ForegroundColor Green

# 2. Workspace directories
$dirs = @("reports", "recon", "output", "logs", "targets", "tools\bin", ".hermes\skills", "bin")
foreach ($d in $dirs) {
    $targetPath = Join-Path $CYBERMES_DIR $d
    if (-not (Test-Path $targetPath)) {
        New-Item -ItemType Directory -Path $targetPath -Force | Out-Null
    }
}

# 3. Setup Python Virtual Environment
Write-Host "`n[*] Setting up Python virtual environment (venv)..." -ForegroundColor Cyan
if (-not (Test-Path "$CYBERMES_DIR\venv")) {
    & $pythonExe @pythonArgs -m venv "$CYBERMES_DIR\venv"
}

$venvPython = "$CYBERMES_DIR\venv\Scripts\python.exe"
$venvPip = "$CYBERMES_DIR\venv\Scripts\pip.exe"

if (-not (Test-Path $venvPython)) {
    Write-Host "[-] Failed to create virtual environment." -ForegroundColor Red
    exit 1
}

Write-Host "[*] Upgrading pip and installing dependencies..." -ForegroundColor Gray
& $venvPython -m pip install --upgrade pip --quiet
& $venvPip install -r "$CYBERMES_DIR\requirements.txt" --quiet

# 4. MCP servers setup (optional)
if (Get-Command npm -ErrorAction SilentlyContinue) {
    Write-Host "[*] Setting up MCP servers..." -ForegroundColor Gray
    try {
        & npm install --prefix "$CYBERMES_DIR" @modelcontextprotocol/server-puppeteer @modelcontextprotocol/server-filesystem 2>$null
    } catch {}
}

# 5. Synchronize skills
if (Test-Path "$CYBERMES_DIR\skills") {
    Copy-Item -Path "$CYBERMES_DIR\skills\*" -Destination "$CYBERMES_DIR\.hermes\skills\" -Recurse -Force -ErrorAction SilentlyContinue
}

# 6. Environment files initialization
if ((-not (Test-Path "$CYBERMES_DIR\.env")) -and (Test-Path "$CYBERMES_DIR\.env.example")) {
    Copy-Item -Path "$CYBERMES_DIR\.env.example" -Destination "$CYBERMES_DIR\.env"
    Write-Host "[+] Generated default .env file" -ForegroundColor Green
}

if ((Test-Path "$CYBERMES_DIR\.env") -and (-not (Test-Path "$CYBERMES_DIR\.hermes\.env"))) {
    Copy-Item -Path "$CYBERMES_DIR\.env" -Destination "$CYBERMES_DIR\.hermes\.env" -ErrorAction SilentlyContinue
}

# Sanitize directory traps
if ((Test-Path "$CYBERMES_DIR\.hermes\config.yaml" -PathType Container)) {
    Remove-Item -Path "$CYBERMES_DIR\.hermes\config.yaml" -Recurse -Force -ErrorAction SilentlyContinue
}
if ((Test-Path "$CYBERMES_DIR\.hermes\auth.json" -PathType Container)) {
    Remove-Item -Path "$CYBERMES_DIR\.hermes\auth.json" -Recurse -Force -ErrorAction SilentlyContinue
}

if ((-not (Test-Path "$CYBERMES_DIR\.hermes\config.yaml")) -and (Test-Path "$CYBERMES_DIR\.hermes\config.yaml.example")) {
    Copy-Item -Path "$CYBERMES_DIR\.hermes\config.yaml.example" -Destination "$CYBERMES_DIR\.hermes\config.yaml"
    Write-Host "[+] Initialized .hermes/config.yaml" -ForegroundColor Green
}

$authJsonPath = "$CYBERMES_DIR\.hermes\auth.json"
if (-not (Test-Path $authJsonPath)) {
    "{}" | Out-File -FilePath $authJsonPath -Encoding utf8
}

# 7. Compile Go Core Tools (if Go is available)
if (Get-Command go -ErrorAction SilentlyContinue) {
    Write-Host "[*] Compiling Cybermes Go tools into tools\bin\..." -ForegroundColor Cyan
    try {
        & go build -ldflags="-s -w" -o "$CYBERMES_DIR\tools\bin\smart_pipe.exe" "$CYBERMES_DIR\cmd\smart_pipe"
        & go build -ldflags="-s -w" -o "$CYBERMES_DIR\tools\bin\secret_scan.exe" "$CYBERMES_DIR\cmd\secret_scan"
        & go build -ldflags="-s -w" -o "$CYBERMES_DIR\tools\bin\search_knowledge.exe" "$CYBERMES_DIR\cmd\search_knowledge"
        & go build -ldflags="-s -w" -o "$CYBERMES_DIR\tools\bin\aggregate_reports.exe" "$CYBERMES_DIR\cmd\aggregate_reports"
        & go build -ldflags="-s -w" -o "$CYBERMES_DIR\tools\bin\cybermes-mcp.exe" "$CYBERMES_DIR\cmd\cybermes-mcp"
        Write-Host "[+] Built Go tools (including cybermes-mcp.exe) in tools\bin\" -ForegroundColor Green
    } catch {
        Write-Host "[!] Could not compile Go binaries locally; skipping" -ForegroundColor DarkYellow
    }
}

# 8. Download ProjectDiscovery Security Toolchain
$updaterPs1 = Join-Path $CYBERMES_DIR "tools\update_tools.ps1"
if (Test-Path $updaterPs1) {
    try {
        & powershell -ExecutionPolicy Bypass -File $updaterPs1
    } catch {
        Write-Host "[!] Toolchain updater notice: $_" -ForegroundColor DarkYellow
    }
}

Write-Host "`n========================================================" -ForegroundColor Green
Write-Host "  [+] Cybermes Windows Setup Complete!" -ForegroundColor Green
Write-Host "========================================================" -ForegroundColor Green
Write-Host "`nQuick Start:" -ForegroundColor Cyan
Write-Host "  1. Edit your API keys:        notepad .env" -ForegroundColor Yellow
Write-Host "  2. 1-Click MCP AI Setup:      python scripts\setup_mcp.py --local" -ForegroundColor Yellow
Write-Host "  3. Run assessment:            .\cybermes.bat `"Assess http://target.com`"" -ForegroundColor Yellow
Write-Host "  4. Run doctor diagnostics:    python tools\doctor.py" -ForegroundColor Yellow
