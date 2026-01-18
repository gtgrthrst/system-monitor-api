# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

```bash
go build -o sysinfo-api    # Build the binary
go mod tidy                 # Update dependencies
./sysinfo-api               # Run the server (listens on :8088)
```

## Architecture

Single-file Go HTTP API that exposes system metrics using gopsutil v3 with SQLite persistence.

**Endpoints:**
- `GET /` - Real-time web dashboard with MQTT settings
- `GET /health` - Health check, returns `{"status": "ok"}`
- `GET /api/system` - Returns current system info (host, CPU, memory, disk, temperature)
- `GET /api/history` - Query historical data with time range and CSV export support
- `GET /api/history/stats` - Historical data statistics
- `GET /api/mqtt/config` - Get MQTT configuration (password masked)
- `POST /api/mqtt/config` - Save MQTT configuration and reconnect
- `GET /api/mqtt/status` - Get MQTT connection status

**History API Parameters:**
- `?minutes=N` - Query last N minutes (uses memory buffer for ≤60 min, DB for longer)
- `?start=<ts>&end=<ts>` - Query specific Unix timestamp range
- `?format=csv` - Download as CSV file

**Data Storage:**
- Memory ring buffer: 360 points (1 hour) for fast recent queries
- SQLite database: `sysinfo_history.db` for persistent long-term storage

**Data flow:** HTTP handler → `getSystemInfo()` → gopsutil calls → JSON response / SQLite storage

**Key dependencies:**
- `github.com/shirou/gopsutil/v3` - Cross-platform system metrics
- `github.com/mattn/go-sqlite3` - SQLite driver for persistent history
- `github.com/eclipse/paho.mqtt.golang` - MQTT client for metrics publishing

**MQTT Feature:**
- Config stored in `mqtt_config.json` (auto-created on first run)
- Publishes to topic `{topic_prefix}/{client_id}` every 10 seconds
- Payload: `{"hostname":"...","cpu":45.2,"mem":60.5,"disk":29.5,"timestamp":...}`
- Web UI settings in dashboard at bottom of page
- Optional username/password authentication
