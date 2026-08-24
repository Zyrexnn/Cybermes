param(
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$ArgsList
)
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
$env:CYBERMES_DIR = $ScriptDir
$env:HERMES_HOME = Join-Path $ScriptDir ".hermes"
$env:PATH = "$ScriptDir\tools\bin;$ScriptDir\bin;$ScriptDir\venv\Scripts;" + $env:PATH

$hermesExe = Join-Path $ScriptDir "venv\Scripts\hermes.exe"
if (Test-Path $hermesExe) {
    & $hermesExe @ArgsList
} else {
    Write-Error "hermes.exe not found. Please run .\setup_windows.ps1 first."
}
