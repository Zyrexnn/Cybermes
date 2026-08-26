[CmdletBinding()]
param(
    [string]$BinDir = "",
    [switch]$IncludeNuclei = $false
)

$ErrorActionPreference = "Stop"

# Ensure TLS 1.2+ for GitHub API on PowerShell 5.1 & 7+
try {
    [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.SecurityProtocolType]::Tls12 -bor [System.Net.SecurityProtocolType]::Tls13
} catch {
    [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.SecurityProtocolType]::Tls12
}

if ([string]::IsNullOrWhiteSpace($BinDir)) {
    $scriptDir = $PSScriptRoot
    if (-not $scriptDir) {
        $scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
    }
    if (-not $scriptDir) {
        $scriptDir = (Get-Location).Path
    }
    $BinDir = Join-Path $scriptDir "bin"
}

if (-not (Test-Path $BinDir)) {
    New-Item -ItemType Directory -Path $BinDir -Force | Out-Null
}

$arch = if ([System.Environment]::Is64BitOperatingSystem) {
    if ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture -eq [System.Runtime.InteropServices.Architecture]::Arm64) {
        "arm64"
    } else {
        "amd64"
    }
} else {
    "386"
}

Write-Host "[*] Downloading ProjectDiscovery tools (windows/$arch)..." -ForegroundColor Cyan

function Install-PDTool {
    param(
        [string]$Repo,
        [string]$ToolName
    )

    $apiUrl = "https://api.github.com/repos/projectdiscovery/$Repo/releases/latest"
    try {
        $headers = @{ "User-Agent" = "Cybermes-Installer" }
        $release = Invoke-RestMethod -Uri $apiUrl -Headers $headers -TimeoutSec 15
        
        $asset = $release.assets | Where-Object { 
            $_.name -like "*${ToolName}*_windows_${arch}.zip" 
        } | Select-Object -First 1

        if (-not $asset) {
            Write-Host "[!] Asset not found for $ToolName (windows/$arch)" -ForegroundColor DarkYellow
            return
        }

        $tempZip = Join-Path $env:TEMP "$ToolName-$arch.zip"
        $tempExtract = Join-Path $env:TEMP "$ToolName-extract"

        Invoke-WebRequest -Uri $asset.browser_download_url -OutFile $tempZip -UseBasicParsing

        if (Test-Path $tempExtract) {
            Remove-Item -Path $tempExtract -Recurse -Force
        }
        Expand-Archive -Path $tempZip -DestinationPath $tempExtract -Force

        $exeFile = Get-ChildItem -Path $tempExtract -Filter "$ToolName.exe" -Recurse | Select-Object -First 1
        if ($exeFile) {
            $destFile = Join-Path $BinDir "$ToolName.exe"
            Copy-Item -Path $exeFile.FullName -Destination $destFile -Force
            Write-Host "[+] Installed $ToolName -> $destFile" -ForegroundColor Green
        }

        Remove-Item -Path $tempZip -Force -ErrorAction SilentlyContinue
        Remove-Item -Path $tempExtract -Recurse -Force -ErrorAction SilentlyContinue
    } catch {
        Write-Host "[!] Failed downloading $ToolName : $_" -ForegroundColor DarkYellow
    }
}

# Core Lightweight Tools
$tools = @(
    @{ Repo = "subfinder"; Tool = "subfinder" },
    @{ Repo = "httpx"; Tool = "httpx" },
    @{ Repo = "katana"; Tool = "katana" }
)

if ($IncludeNuclei) {
    Write-Host "[*] Nuclei included as requested (-IncludeNuclei)..." -ForegroundColor Cyan
    $tools += @{ Repo = "nuclei"; Tool = "nuclei" }
} else {
    Write-Host "[i] Nuclei skipped by default (On-Demand / Optional). Use -IncludeNuclei to install." -ForegroundColor DarkGray
}

foreach ($t in $tools) {
    Install-PDTool -Repo $t.Repo -ToolName $t.Tool
}

$nucleiExe = Join-Path $BinDir "nuclei.exe"
if ($IncludeNuclei -and (Test-Path $nucleiExe)) {
    Write-Host "[*] Updating Nuclei templates..." -ForegroundColor Cyan
    try {
        & $nucleiExe -update-templates -silent
        Write-Host "[+] Nuclei templates updated" -ForegroundColor Green
    } catch {
        Write-Host "[!] Template update skipped" -ForegroundColor DarkYellow
    }
}
