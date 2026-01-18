# System Monitor API

[繁體中文](README_zh-TW.md)

A lightweight, cross-platform system monitoring API with real-time web dashboard built with Go and gopsutil.

## Features

- **Real-time Dashboard** - Terminal-style web UI with live updates every 2 seconds
- **Process Monitor** - View all running processes with CPU/memory usage, pagination support
- **Trend Charts** - CPU and memory usage history visualization (60 data points)
- **Temperature Monitoring** - Color-coded sensor temperature display
- **History API** - Query any time range with CSV export support
- **MQTT Integration** - Publish metrics to MQTT broker with web UI configuration
- **SQLite Persistence** - Historical data permanently stored in database
- **Cross-platform** - Works on Linux and Windows
- **Windows Service** - MSI installer with auto-start service support
- **Low Resource Usage** - Minimal CPU and memory footprint

## Screenshots

### Linux Dashboard
<img width="600" alt="Linux Dashboard" src="https://github.com/user-attachments/assets/4fdb7c02-8818-495c-83a2-499c10b339bd" />
<img width="800" height="400" alt="image" src="https://github.com/user-attachments/assets/1a003bd2-ae58-4b29-aee7-41318cacc61a" />

### Windows Dashboard
<img width="600" alt="Windows Dashboard" src="https://github.com/user-attachments/assets/8308422d-8cf8-41c9-b1f5-6bfe1f423b97" />
<img width="800" height="400" alt="image" src="https://github.com/user-attachments/assets/b730840a-6d49-4611-a607-40b3395e1357" />
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
| `GET /` | Web dashboard with real-time monitoring and MQTT settings |
| `GET /processes` | Process monitor page with pagination |
| `GET /health` | Health check endpoint |
| `GET /api/system` | JSON API for system information |
| `GET /api/processes` | Process list API with pagination |
| `GET /api/history` | Historical data query (supports any time range) |
| `GET /api/history/stats` | Historical data statistics |
| `GET /api/mqtt/config` | Get MQTT configuration |
| `POST /api/mqtt/config` | Save MQTT configuration |
| `GET /api/mqtt/status` | Get MQTT connection status |

### History API

```
GET /api/history?minutes=N
GET /api/history?start=<unix_timestamp>&end=<unix_timestamp>
GET /api/history?start=<unix_timestamp>&end=<unix_timestamp>&format=csv
```

| Parameter | Default | Description |
|-----------|---------|-------------|
| `minutes` | 60 | Query last N minutes of data (no limit) |
| `start` | - | Start time Unix timestamp |
| `end` | now | End time Unix timestamp |
| `format` | json | Output format: `json` or `csv` |

**Response (JSON):**
```json
{
  "interval_seconds": 10,
  "start_time": 1768706921,
  "end_time": 1768708721,
  "count": 180,
  "data": [
    {"ts": 1768708721, "cpu": 45.2, "mem": 60.5, "disk": 29.5},
    ...
  ]
}
```

**Response (CSV):**
```csv
timestamp,datetime,cpu_percent,mem_percent,disk_percent
1768708721,2026-01-18 10:30:21,45.20,60.50,29.50
1768708731,2026-01-18 10:30:31,42.10,60.80,29.50
...
```

### History Stats API

```
GET /api/history/stats
```

**Response:**
```json
{
  "total_records": 8640,
  "min_timestamp": 1768622321,
  "max_timestamp": 1768708721,
  "min_datetime": "2026-01-17 10:30:21",
  "max_datetime": "2026-01-18 10:30:21",
  "duration_hours": 24.0,
  "interval_seconds": 10
}
```

### Usage Examples

```bash
# Query last 30 minutes
curl "http://localhost:8088/api/history?minutes=30"

# Query last 24 hours
curl "http://localhost:8088/api/history?minutes=1440"

# Query specific time range
curl "http://localhost:8088/api/history?start=1768622321&end=1768708721"

# Download CSV file
curl -o history.csv "http://localhost:8088/api/history?minutes=60&format=csv"

# View history statistics
curl "http://localhost:8088/api/history/stats"
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

## Process Monitor

A dedicated page for viewing all running processes, accessible at `/processes`.

### Features

- View all system processes sorted by CPU usage
- Pagination support (50 processes per page)
- Auto-refresh every 2 seconds
- Displays: PID, Name, CPU%, Memory%, Status, User

### Process API

```
GET /api/processes?page=1&limit=50
```

| Parameter | Default | Description |
|-----------|---------|-------------|
| `page` | 1 | Page number |
| `limit` | 50 | Processes per page (max 200) |

**Response:**
```json
{
  "total": 156,
  "page": 1,
  "limit": 50,
  "total_pages": 4,
  "timestamp": 1737200000,
  "processes": [
    {
      "pid": 1234,
      "name": "chrome",
      "cpu_percent": 25.3,
      "mem_percent": 12.5,
      "status": "running",
      "username": "root"
    }
  ]
}
```

## MQTT Integration

The dashboard includes a built-in MQTT settings panel for publishing system metrics to an MQTT broker.

### Configuration

Configure MQTT via the web dashboard or API:

| Field | Description |
|-------|-------------|
| **Broker** | MQTT broker address (e.g., `tcp://broker.example.com:1883`) |
| **Client ID** | Custom device identifier (uses hostname if empty) |
| **Username** | Optional authentication username |
| **Password** | Optional authentication password |

### MQTT Message Format

**Topic:** `sysinfo/{client_id}`

**Payload (every 10 seconds):**
```json
{
  "hostname": "my-device",
  "cpu": 45.2,
  "mem": 60.5,
  "disk": 29.5,
  "timestamp": 1737200000
}
```

### MQTT API

```bash
# Get current configuration
curl http://localhost:8088/api/mqtt/config

# Save configuration
curl -X POST http://localhost:8088/api/mqtt/config \
  -H "Content-Type: application/json" \
  -d '{"enabled":true,"broker":"tcp://broker:1883","client_id":"my-device","username":"","password":"","topic_prefix":"sysinfo"}'

# Check connection status
curl http://localhost:8088/api/mqtt/status
```

### Configuration File

MQTT settings are stored in `mqtt_config.json`:

```json
{
  "enabled": true,
  "broker": "tcp://broker.example.com:1883",
  "username": "",
  "password": "",
  "topic_prefix": "sysinfo",
  "client_id": "my-device"
}
```

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
