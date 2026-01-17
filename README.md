# System Monitor API

[繁體中文](README_zh-TW.md)

A simple JSON API for system monitoring using Go and gopsutil.

## One-Line Install

```bash
curl -fsSL https://raw.githubusercontent.com/gtgrthrst/system-monitor-api/main/install.sh | sudo bash
```

## API Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /health` | Health check |
| `GET /api/system` | System information (CPU, memory, disk, host) |

## Manual Build

```bash
go build -o sysinfo-api
./sysinfo-api
```

Server runs on port **8088**.
