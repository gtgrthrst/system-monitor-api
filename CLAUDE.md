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
- `GET /processes` - Process monitor page with pagination
- `GET /health` - Health check, returns `{"status": "ok"}`
- `GET /api/system` - Returns current system info (host, CPU, memory, disk, temperature)
- `GET /api/processes` - Process list with pagination (?page=1&limit=50)
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
- Memory ring buffer: 120 points (1 hour at 30s interval) for fast recent queries
- SQLite database: `sysinfo_history.db` for persistent long-term storage

**CPU Optimization (Low Resource Mode):**
- Background CPU collector: Runs every 2s with blocking measurement for accuracy
- System info cache: 3s TTL to reduce repeated gopsutil calls
- Host info cache: 5min TTL (host info rarely changes)
- Process list cache: 15s TTL
- History collection: 30s interval (configurable)
- Temperature monitoring: Can be disabled via `enableTemperature` flag

**Data flow:** HTTP handler → `getCachedSystemInfo()` → cache hit or `getSystemInfo()` → gopsutil calls → JSON response / SQLite storage

**Key dependencies:**
- `github.com/shirou/gopsutil/v3` - Cross-platform system metrics
- `github.com/mattn/go-sqlite3` - SQLite driver for persistent history
- `github.com/eclipse/paho.mqtt.golang` - MQTT client for metrics publishing

**MQTT Feature:**
- Config stored in `mqtt_config.json` (auto-created on first run)
- Publishes to topic `{topic_prefix}/{client_id}` every 30 seconds
- Payload: `{"hostname":"...","cpu":45.2,"mem":60.5,"disk":29.5,"boot_time":1737250800}`
- Web UI settings in dashboard at bottom of page
- Optional username/password authentication
