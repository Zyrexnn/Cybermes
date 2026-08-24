[CmdletBinding()]
param(
    [string]$BinDir = (Join-Path $PSScriptRoot "bin")
)

$ErrorActionPreference = "Stop"

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

Write-Host "🔄 [Cybermes Windows Updater] Architecture: windows/$arch" -ForegroundColor Cyan

function Install-PDTool {
    param(
        [string]$Repo,
        [string]$ToolName
    )

    Write-Host "📦 Checking $ToolName ($Repo)..." -ForegroundColor Gray
    $apiUrl = "https://api.github.com/repos/projectdiscovery/$Repo/releases/latest"

    try {
        $headers = @{ "User-Agent" = "Cybermes-Updater" }
        $release = Invoke-RestMethod -Uri $apiUrl -Headers $headers -TimeoutSec 15
        
        $asset = $release.assets | Where-Object { 
            $_.name -like "*${ToolName}*_windows_${arch}.zip" 
        } | Select-Object -First 1

        if (-not $asset) {
            Write-Host "   ℹ️ Release asset not found for $ToolName (windows/$arch)" -ForegroundColor DarkYellow
            return
        }

        $tempZip = Join-Path $env:TEMP "$ToolName-$arch.zip"
        $tempExtract = Join-Path $env:TEMP "$ToolName-extract"

        Write-Host "   ⬇️ Downloading $($asset.name)..." -ForegroundColor Gray
        Invoke-WebRequest -Uri $asset.browser_download_url -OutFile $tempZip -UseBasicParsing

        if (Test-Path $tempExtract) {
            Remove-Item -Path $tempExtract -Recurse -Force
        }
        Expand-Archive -Path $tempZip -DestinationPath $tempExtract -Force

        $exeFile = Get-ChildItem -Path $tempExtract -Filter "$ToolName.exe" -Recurse | Select-Object -First 1
        if ($exeFile) {
            $destFile = Join-Path $BinDir "$ToolName.exe"
            Copy-Item -Path $exeFile.FullName -Destination $destFile -Force
            Write-Host "   ✅ Installed: $destFile" -ForegroundColor Green
        }

        Remove-Item -Path $tempZip -Force -ErrorAction SilentlyContinue
        Remove-Item -Path $tempExtract -Recurse -Force -ErrorAction SilentlyContinue
    } catch {
        Write-Host "   [!] Failed to update $ToolName : $_" -ForegroundColor DarkYellow
    }
}

$tools = @(
    @{ Repo = "subfinder"; Tool = "subfinder" },
    @{ Repo = "httpx"; Tool = "httpx" },
    @{ Repo = "katana"; Tool = "katana" },
    @{ Repo = "nuclei"; Tool = "nuclei" }
)

foreach ($t in $tools) {
    Install-PDTool -Repo $t.Repo -ToolName $t.Tool
}

$nucleiExe = Join-Path $BinDir "nuclei.exe"
if (Test-Path $nucleiExe) {
    Write-Host "📜 Updating Nuclei community templates..." -ForegroundColor Cyan
    try {
        & $nucleiExe -update-templates -silent
        Write-Host "✅ Nuclei templates updated." -ForegroundColor Green
    } catch {
        Write-Host "   [!] Template update skipped or failed." -ForegroundColor DarkYellow
    }
}

Write-Host "✨ [Cybermes Windows Updater] Toolchain update completed." -ForegroundColor Green
