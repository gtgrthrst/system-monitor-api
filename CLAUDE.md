# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

```bash
go build -o sysinfo-api    # Build the binary
go mod tidy                 # Update dependencies
./sysinfo-api               # Run the server (listens on :8088)
```

## Architecture

Single-file Go HTTP API that exposes system metrics using gopsutil v3.

**Endpoints:**
- `GET /health` - Health check, returns `{"status": "ok"}`
- `GET /api/system` - Returns system info (host, CPU, memory, disk)

**Data flow:** HTTP handler → `getSystemInfo()` → gopsutil calls → JSON response

**Key dependency:** `github.com/shirou/gopsutil/v3` for cross-platform system metrics (cpu, disk, host, mem packages)
