# System Monitor API

[繁體中文](README_zh-TW.md)

A simple JSON API for system monitoring using Go and gopsutil.

## Installation

### Linux

```bash
curl -fsSL https://raw.githubusercontent.com/gtgrthrst/system-monitor-api/main/install.sh | sudo bash
```

### Windows

Run PowerShell as Administrator:

```powershell
irm https://raw.githubusercontent.com/gtgrthrst/system-monitor-api/main/install.ps1 | iex
```

Or download and run manually:

```powershell
powershell -ExecutionPolicy Bypass -File install.ps1
```

## Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /` | Web dashboard (Glances-like UI) |
| `GET /health` | Health check |
| `GET /api/system` | System information (CPU, memory, disk, host) |

## Manual Build

### Linux / macOS

```bash
go build -o sysinfo-api
./sysinfo-api
```

### Windows

```powershell
go build -o sysinfo-api.exe
.\sysinfo-api.exe
```

Server runs on port **8088**.
