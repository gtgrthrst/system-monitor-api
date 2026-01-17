# System Monitor API - Windows Installer
# Run as Administrator: powershell -ExecutionPolicy Bypass -File install.ps1

$ErrorActionPreference = "Stop"
$DOWNLOAD_URL = "https://raw.githubusercontent.com/gtgrthrst/system-monitor-api/main/sysinfo-api-windows-amd64.exe"
$INSTALL_DIR = "$env:ProgramFiles\sysinfo-api"
$SERVICE_NAME = "SysinfoAPI"

Write-Host "=== System Monitor API Installer ===" -ForegroundColor Green

# Check if running as Administrator
if (-NOT ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")) {
    Write-Host "Please run as Administrator" -ForegroundColor Red
    exit 1
}

# Create install directory
if (-NOT (Test-Path $INSTALL_DIR)) {
    New-Item -ItemType Directory -Path $INSTALL_DIR -Force | Out-Null
}

# Download pre-built binary
Write-Host "Downloading sysinfo-api..."
Invoke-WebRequest -Uri $DOWNLOAD_URL -OutFile "$INSTALL_DIR\sysinfo-api.exe"
Write-Host "Download complete" -ForegroundColor Green

# Create Windows Service using sc.exe
Write-Host "Creating Windows service..."
$binPath = "$INSTALL_DIR\sysinfo-api.exe"

# Remove existing service if exists
sc.exe stop $SERVICE_NAME 2>$null
sc.exe delete $SERVICE_NAME 2>$null
Start-Sleep -Seconds 1

# Create new service
sc.exe create $SERVICE_NAME binPath= $binPath start= auto
sc.exe description $SERVICE_NAME "System Monitor API Service"
sc.exe start $SERVICE_NAME

Write-Host ""
Write-Host "=== Installation Complete ===" -ForegroundColor Green
Write-Host "Service status: $((Get-Service $SERVICE_NAME).Status)"
Write-Host "API endpoints:"
Write-Host "  http://localhost:8088/"
Write-Host "  http://localhost:8088/health"
Write-Host "  http://localhost:8088/api/system"
