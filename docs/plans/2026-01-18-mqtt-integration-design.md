# MQTT Integration Design

## Overview

Add MQTT configuration interface to sysinfo-api that periodically publishes CPU and RAM metrics to an external MQTT broker.

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Web Dashboard                         │
│  ┌──────────────┐  ┌──────────────────────────────────┐ │
│  │ System Monitor│  │ MQTT Settings                    │ │
│  │ (existing)    │  │ ├─ Broker: tcp://broker:1883    │ │
│  │              │  │ ├─ Client ID: my-device          │ │
│  │              │  │ ├─ Username: (optional)          │ │
│  │              │  │ ├─ Password: ********            │ │
│  │              │  │ ├─ [Enable/Disable] toggle       │ │
│  │              │  │ └─ Connection status indicator   │ │
│  └──────────────┘  └──────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────┐
│                    Go Backend                            │
│  collectHistory() ──┬──► Memory Buffer + SQLite (existing)│
│      every 10s      └──► MQTT Publish (new)              │
│                          │                               │
│                          ▼                               │
│              Topic: sysinfo/{client_id}                  │
│              Payload: {"hostname":...,"cpu":...,"mem":...}│
└─────────────────────────────────────────────────────────┘
                            │
                            ▼
                   ┌─────────────────┐
                   │  External MQTT  │
                   │     Broker      │
                   └─────────────────┘
```

## Configuration

**File:** `mqtt_config.json`

```json
{
  "enabled": true,
  "broker": "tcp://broker.example.com:1883",
  "username": "",
  "password": "",
  "topic_prefix": "sysinfo",
  "client_id": "my-custom-device"
}
```

| Field | Type | Description |
|-------|------|-------------|
| `enabled` | bool | Enable/disable MQTT publishing |
| `broker` | string | MQTT broker address (tcp://host:port) |
| `username` | string | Optional authentication username |
| `password` | string | Optional authentication password |
| `topic_prefix` | string | Topic prefix (default: "sysinfo") |
| `client_id` | string | Custom device name; uses system hostname if empty |

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/mqtt/config` | Get current config (password masked as `"***"`) |
| `POST` | `/api/mqtt/config` | Save config and reconnect |
| `GET` | `/api/mqtt/status` | Get connection status |

**Status Response:**
```json
{
  "connected": true,
  "status": "connected",
  "broker": "tcp://broker.example.com:1883",
  "topic": "sysinfo/my-device"
}
```

## MQTT Message Format

**Topic:** `{topic_prefix}/{client_id}`
**Example:** `sysinfo/my-device`

**Payload:**
```json
{
  "hostname": "my-device",
  "cpu": 45.2,
  "mem": 60.5,
  "disk": 29.5,
  "timestamp": 1737200000
}
```

## Implementation

**New dependency:**
```
github.com/eclipse/paho.mqtt.golang v1.4.3
```

**New functions in main.go:**
- `loadMQTTConfig()` - Load config from file on startup
- `saveMQTTConfig()` - Save config to file
- `connectMQTT()` - Establish/reconnect MQTT connection
- `disconnectMQTT()` - Close MQTT connection
- `publishMetrics()` - Publish CPU/RAM data
- `handleMQTTConfig()` - Handle GET/POST /api/mqtt/config
- `handleMQTTStatus()` - Handle GET /api/mqtt/status

**Integration points:**
- Call `publishMetrics()` in `collectHistory()` loop
- Call `loadMQTTConfig()` and `connectMQTT()` in `program.run()`
- Call `disconnectMQTT()` in `program.Stop()`

## Web UI

Add MQTT settings section to existing dashboard with terminal-style design:
- Connection status indicator (green=connected, gray=disabled, red=error)
- Live topic preview based on Client ID
- Auto-reconnect on save
- Password field masked with `type="password"`
