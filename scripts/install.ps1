$ErrorActionPreference = 'Stop'

$Repo = "0DayMonxrch/vaultify"
$InstallDir = "$env:USERPROFILE\.vaultify\bin"
$ZipName = "vaultify_Windows_x86_64.zip"

Write-Host "Fetching latest release of Vaultify..." -ForegroundColor Cyan

# Get latest release tag from GitHub API
try {
    $ReleaseInfo = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
    $Version = $ReleaseInfo.tag_name
} catch {
    Write-Host "Failed to fetch latest release. Check your internet connection." -ForegroundColor Red
    exit 1
}

$DownloadUrl = "https://github.com/$Repo/releases/download/$Version/$ZipName"
$ZipPath = "$env:TEMP\$ZipName"

Write-Host "Downloading $DownloadUrl..."
Invoke-WebRequest -Uri $DownloadUrl -OutFile $ZipPath

if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
}

Write-Host "Extracting binary..."
Expand-Archive -Path $ZipPath -DestinationPath $InstallDir -Force
Remove-Item -Path $ZipPath -Force

$UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($UserPath -notmatch [regex]::Escape($InstallDir)) {
    Write-Host "Adding $InstallDir to your User PATH variable..."
    $NewPath = "$InstallDir;$UserPath"
    [Environment]::SetEnvironmentVariable("PATH", $NewPath, "User")
    $env:PATH = "$InstallDir;$env:PATH" 
}

Write-Host "`n✓ Vaultify successfully installed!" -ForegroundColor Green
Write-Host "You can now execute it globally. Try running: vaultify" -ForegroundColor Cyan
