# Katsh Installer for Windows (PowerShell)
# This script builds and installs Katsh to your PATH

param(
    [string]$InstallDir = "$env:LOCALAPPDATA\Katsh\bin"
)

$ErrorActionPreference = "Stop"

Write-Host "═══════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "  Katsh Installer for Windows" -ForegroundColor Cyan
Write-Host "═══════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host ""

# Get script directory
$ScriptDir = $PSScriptRoot
if (-not $ScriptDir) {
    $ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
}

# Find project root (parent of install folder)
$ProjectRoot = Split-Path -Parent $ScriptDir

Write-Host "[1/4] Checking Go installation..." -ForegroundColor Yellow
$GoVersion = & go version 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Go is not installed or not in PATH" -ForegroundColor Red
    Write-Host "Please install Go from: https://go.dev/dl/" -ForegroundColor Red
    exit 1
}
Write-Host "  Found: $GoVersion" -ForegroundColor Green

Write-Host "[2/4] Building Katsh..." -ForegroundColor Yellow
Set-Location $ProjectRoot

# Build for Windows
& go build -o katsh.exe .
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Build failed!" -ForegroundColor Red
    exit 1
}
Write-Host "  Build successful!" -ForegroundColor Green

Write-Host "[3/4] Installing to $InstallDir..." -ForegroundColor Yellow

# Create installation directory
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

# Copy binary
Copy-Item -Path "$ProjectRoot\katsh.exe" -Destination $InstallDir -Force

# Clean up build artifact
Remove-Item -Path "$ProjectRoot\katsh.exe" -Force

Write-Host "  Installed to: $InstallDir" -ForegroundColor Green

Write-Host "[4/4] Adding to PATH..." -ForegroundColor Yellow

# Add to PATH for current session
$CurrentPath = $env:PATH
if ($CurrentPath -notlike "*$InstallDir*") {
    $env:PATH = "$InstallDir;$CurrentPath"
    
    # Persist to user PATH (optional - ask or detect admin)
    $UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($UserPath -notlike "*$InstallDir*") {
        $NewUserPath = "$InstallDir;$UserPath"
        [Environment]::SetEnvironmentVariable("Path", $NewUserPath, "User")
        Write-Host "  Added to user PATH (persistent)" -ForegroundColor Green
    }
} else {
    Write-Host "  Already in PATH" -ForegroundColor Green
}

Write-Host ""
Write-Host "═══════════════════════════════════════════════════════════" -ForegroundColor Green
Write-Host "  Installation complete!" -ForegroundColor Green
Write-Host "═══════════════════════════════════════════════════════════" -ForegroundColor Green
Write-Host ""
Write-Host "Run 'katsh' to start the shell." -ForegroundColor Cyan
Write-Host ""
Write-Host "Note: You may need to restart your terminal for PATH changes to take effect." -ForegroundColor Gray
