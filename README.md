# System Monitor API

[繁體中文](README_zh-TW.md)

A lightweight, cross-platform system monitoring API with real-time web dashboard built with Go and gopsutil.

## Features

- **Real-time Dashboard** - Terminal-style web UI with live updates every 2 seconds
- **Trend Charts** - CPU and memory usage history visualization (60 data points)
- **Temperature Monitoring** - Color-coded sensor temperature display
- **History API** - Query historical data up to 1 hour
- **Cross-platform** - Works on Linux and Windows
- **Windows Service** - MSI installer with auto-start service support
- **Low Resource Usage** - Minimal CPU and memory footprint

## Screenshots

### Linux Dashboard
<img width="600" alt="Linux Dashboard" src="https://github.com/user-attachments/assets/4fdb7c02-8818-495c-83a2-499c10b339bd" />

### Windows Dashboard
<img width="600" alt="Windows Dashboard" src="https://github.com/user-attachments/assets/8308422d-8cf8-41c9-b1f5-6bfe1f423b97" />

## Installation

### Linux (One-line Install)

```bash
curl -fsSL https://raw.githubusercontent.com/gtgrthrst/system-monitor-api/main/install.sh | sudo bash
```

This will:
- Install Go if not present
- Download and build the application
- Create a systemd service
- Auto-start on boot

### Windows (MSI Installer)

Download from [Releases](https://github.com/gtgrthrst/system-monitor-api/releases/latest):

| File | Description |
|------|-------------|
| [sysinfo-api.msi](https://github.com/gtgrthrst/system-monitor-api/releases/latest/download/sysinfo-api.msi) | MSI installer with Windows Service |
| [sysinfo-api-windows-amd64.exe](https://github.com/gtgrthrst/system-monitor-api/releases/latest/download/sysinfo-api-windows-amd64.exe) | Standalone executable |

**MSI Features:**
- One-click installation
- Auto-registers as Windows Service
- Auto-starts on system boot
- Install location: `C:\Program Files\SysinfoAPI\`

## Usage

After installation, open your browser:

```
http://localhost:8088/
```

## API Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /` | Web dashboard with real-time monitoring |
| `GET /health` | Health check endpoint |
| `GET /api/system` | JSON API for system information |
| `GET /api/history` | Historical data (up to 1 hour) |

### History API

```
GET /api/history?minutes=N
```

| Parameter | Default | Max | Description |
|-----------|---------|-----|-------------|
| `minutes` | 60 | 60 | Query last N minutes of data |

**Response:**
```json
{
  "interval_seconds": 10,
  "max_minutes": 60,
  "requested_minutes": 30,
  "count": 180,
  "data": [
    {"ts": 1768708721, "cpu": 45.2, "mem": 60.5, "disk": 29.5},
    ...
  ]
}
```

### System API Response Example

```json
{
  "host": {
    "hostname": "my-server",
    "os": "linux",
    "platform": "ubuntu",
    "uptime_seconds": 123456
  },
  "cpu": {
    "cores": 4,
    "model_name": "Intel(R) Xeon(R) CPU",
    "usage_percent": [12.5, 8.3, 15.2, 10.1]
  },
  "memory": {
    "total_bytes": 8589934592,
    "used_bytes": 4294967296,
    "free_bytes": 4294967296,
    "used_percent": 50.0
  },
  "disk": {
    "total_bytes": 107374182400,
    "used_bytes": 53687091200,
    "free_bytes": 53687091200,
    "used_percent": 50.0
  },
  "temperature": [
    {"name": "coretemp_core_0", "temperature": 45.0},
    {"name": "coretemp_core_1", "temperature": 47.0}
  ]
}
```

## Dashboard Features

| Section | Description |
|---------|-------------|
| **HOST** | Hostname, OS, platform, uptime |
| **CPU** | Model name, per-core usage with progress bars, trend chart |
| **MEMORY** | Total/used/free memory, usage percentage, trend chart |
| **DISK** | Total/used/free disk space, usage percentage |
| **TEMPERATURE** | Sensor temperatures with color coding |

### Temperature Color Codes

| Color | Temperature | Status |
|-------|-------------|--------|
| Blue | < 30°C | Cold |
| Green | 30-50°C | Normal |
| Orange | 50-70°C | Warm |
| Red | > 70°C | Hot |

## Manual Build

### Prerequisites

- Go 1.20 or later

### Linux / macOS

```bash
git clone https://github.com/gtgrthrst/system-monitor-api.git
cd system-monitor-api
go build -o sysinfo-api
./sysinfo-api
```

### Windows

```powershell
git clone https://github.com/gtgrthrst/system-monitor-api.git
cd system-monitor-api
go build -o sysinfo-api.exe
.\sysinfo-api.exe
```

## Service Management

### Linux (systemd)

```bash
sudo systemctl status sysinfo-api   # Check status
sudo systemctl restart sysinfo-api  # Restart service
sudo systemctl stop sysinfo-api     # Stop service
sudo systemctl start sysinfo-api    # Start service
sudo systemctl enable sysinfo-api   # Enable on boot
sudo systemctl disable sysinfo-api  # Disable on boot
```

### Windows

```powershell
Get-Service SysinfoAPI              # Check status
Restart-Service SysinfoAPI          # Restart service
Stop-Service SysinfoAPI             # Stop service
Start-Service SysinfoAPI            # Start service
```

## Configuration

Default port: **8088**

To change the port, modify the source code and rebuild.

## License

MIT License
