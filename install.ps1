# System Monitor API - Windows Installer
# Run as Administrator: powershell -ExecutionPolicy Bypass -File install.ps1

$ErrorActionPreference = "Stop"
$REPO = "https://github.com/gtgrthrst/system-monitor-api.git"
$INSTALL_DIR = "$env:ProgramFiles\sysinfo-api"
$SERVICE_NAME = "SysinfoAPI"

Write-Host "=== System Monitor API Installer ===" -ForegroundColor Green

# Check if running as Administrator
if (-NOT ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")) {
    Write-Host "Please run as Administrator" -ForegroundColor Red
    exit 1
}

# Check if Go is installed, if not install it
if (-NOT (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "Installing Go..." -ForegroundColor Yellow
    $goVersion = "1.22.0"
    $goInstaller = "$env:TEMP\go$goVersion.windows-amd64.msi"

    Write-Host "Downloading Go $goVersion..."
    Invoke-WebRequest -Uri "https://go.dev/dl/go$goVersion.windows-amd64.msi" -OutFile $goInstaller

    Write-Host "Installing Go..."
    Start-Process msiexec.exe -ArgumentList "/i", $goInstaller, "/quiet", "/norestart" -Wait

    # Add Go to PATH for current session
    $env:Path = "$env:Path;C:\Program Files\Go\bin"

    Remove-Item $goInstaller -Force
    Write-Host "Go installed successfully" -ForegroundColor Green
}

Write-Host "Go version: $(go version)"

# Clone or update repo
if (Test-Path $INSTALL_DIR) {
    Write-Host "Updating existing installation..."
    Set-Location $INSTALL_DIR
    git pull
} else {
    Write-Host "Cloning repository..."
    git clone $REPO $INSTALL_DIR
    Set-Location $INSTALL_DIR
}

# Build
Write-Host "Building..."
go build -o sysinfo-api.exe

# Create Windows Service using sc.exe
Write-Host "Creating Windows service..."
$binPath = "$INSTALL_DIR\sysinfo-api.exe"

# Remove existing service if exists
sc.exe stop $SERVICE_NAME 2>$null
sc.exe delete $SERVICE_NAME 2>$null

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
