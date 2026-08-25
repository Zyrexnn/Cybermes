# Cybermes Environment Loader for Windows PowerShell
# Usage: . .\env.ps1

$ScriptDir = $PSScriptRoot
if (-not $ScriptDir) {
    $ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
}
if (-not $ScriptDir) {
    $ScriptDir = Get-Location | Select-Object -ExpandProperty Path
}

$env:CYBERMES_DIR = $ScriptDir
$env:HERMES_HOME = Join-Path $ScriptDir ".hermes"
$env:PATH = "$ScriptDir\tools\bin;$ScriptDir\bin;$ScriptDir\venv\Scripts;" + $env:PATH

# Load .env file variables if present
$envFile = Join-Path $ScriptDir ".env"
if (Test-Path $envFile) {
    Get-Content $envFile | ForEach-Object {
        $line = $_.Trim()
        if ($line -and -not $line.StartsWith("#") -and ($line -match '^([^=]+)=(.*)$')) {
            $name = $matches[1].Trim()
            $val = $matches[2].Trim().Trim('"').Trim("'")
            [System.Environment]::SetEnvironmentVariable($name, $val, "Process")
        }
    }
}

# Activate Python Virtual Environment if present
$activateScript = Join-Path $ScriptDir "venv\Scripts\Activate.ps1"
if (Test-Path $activateScript) {
    & $activateScript
}

Write-Host "[+] Cybermes Environment Activated: $ScriptDir" -ForegroundColor Green
Write-Host "  Tools added to PATH: tools\bin, bin, venv\Scripts" -ForegroundColor Gray
Write-Host "  You can now run 'hermes' directly in this session." -ForegroundColor Cyan
