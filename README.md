# System Monitor API


[繁體中文](README_zh-TW.md)

A simple JSON API for system monitoring using Go and gopsutil.

## Installation

### Linux

```bash
curl -fsSL https://raw.githubusercontent.com/gtgrthrst/system-monitor-api/main/install.sh | sudo bash
```
<img width="818" height="866" alt="Snipaste_2026-01-17_21-47-35" src="https://github.com/user-attachments/assets/4fdb7c02-8818-495c-83a2-499c10b339bd" />

### Windows (MSI Installer)

Download from [Releases](https://github.com/gtgrthrst/system-monitor-api/releases/latest):
- [sysinfo-api.msi](https://github.com/gtgrthrst/system-monitor-api/releases/latest/download/sysinfo-api.msi) - Auto-installs as Windows Service
- [sysinfo-api-windows-amd64.exe](https://github.com/gtgrthrst/system-monitor-api/releases/latest/download/sysinfo-api-windows-amd64.exe) - Standalone executable
<img width="834" height="731" alt="Snipaste_2026-01-17_21-47-14" src="https://github.com/user-attachments/assets/8308422d-8cf8-41c9-b1f5-6bfe1f423b97" />
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
