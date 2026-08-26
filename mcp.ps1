param(
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$ArgsList
)

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
$pythonExe = $null

if (Test-Path "$ScriptDir\venv\Scripts\python.exe") {
    $pythonExe = "$ScriptDir\venv\Scripts\python.exe"
} elseif (Get-Command python -ErrorAction SilentlyContinue) {
    $pythonExe = "python"
} elseif (Get-Command py -ErrorAction SilentlyContinue) {
    $pythonExe = "py"
}

if (-not $pythonExe) {
    Write-Error "Python 3 is required to run the MCP Manager. Please install Python."
    exit 1
}

$mcpScript = Join-Path $ScriptDir "scripts\mcp.py"
& $pythonExe $mcpScript @ArgsList
