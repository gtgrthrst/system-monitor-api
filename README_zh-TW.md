# 系統監控 API

[English](README.md)

使用 Go 和 gopsutil 建立的輕量級跨平台系統監控 API，具有即時 Web 儀表板。

## 功能特色

- **即時儀表板** - 終端機風格 Web 介面，每 5 秒自動更新
- **程序監控** - 檢視所有執行中的程序，支援分頁瀏覽
- **趨勢圖表** - CPU 和記憶體使用率歷史視覺化（60 個數據點）
- **溫度監控** - 彩色標示的感測器溫度顯示
- **歷史資料 API** - 查詢任意時段的歷史資料，支援 CSV 下載
- **MQTT 整合** - 發布系統指標至 MQTT Broker，支援 Web UI 設定
- **SQLite 持久化** - 歷史資料永久保存於資料庫
- **跨平台支援** - 支援 Linux 和 Windows
- **Windows 服務** - MSI 安裝程式支援開機自動啟動
- **低資源消耗** - 極低的 CPU 和記憶體使用量

## 截圖

### Linux 儀表板
<img width="818" alt="Linux Dashboard" src="https://github.com/user-attachments/assets/b84004a9-ae34-4241-90eb-33b1f963719e" />
<img width="800" height="400" alt="image" src="https://github.com/user-attachments/assets/1a003bd2-ae58-4b29-aee7-41318cacc61a" />

### Windows 儀表板
<img width="834" alt="Windows Dashboard" src="https://github.com/user-attachments/assets/8308422d-8cf8-41c9-b1f5-6bfe1f423b97" />

<img width="800" height="400" alt="image" src="https://github.com/user-attachments/assets/b730840a-6d49-4611-a607-40b3395e1357" />

## 安裝

### Linux（一鍵安裝）

```bash
curl -fsSL https://raw.githubusercontent.com/gtgrthrst/system-monitor-api/main/install.sh | sudo bash
```

安裝腳本會自動：
- 安裝 Go（如未安裝）
- 下載並編譯應用程式
- 建立 systemd 服務
- 設定開機自動啟動

### Windows（MSI 安裝程式）

從 [Releases](https://github.com/gtgrthrst/system-monitor-api/releases/latest) 下載：

| 檔案 | 說明 |
|------|------|
| [sysinfo-api.msi](https://github.com/gtgrthrst/system-monitor-api/releases/latest/download/sysinfo-api.msi) | MSI 安裝程式（含 Windows 服務） |
| [sysinfo-api-windows-amd64.exe](https://github.com/gtgrthrst/system-monitor-api/releases/latest/download/sysinfo-api-windows-amd64.exe) | 獨立執行檔 |

**MSI 安裝特色：**
- 一鍵安裝
- 自動註冊為 Windows 服務
- 開機自動啟動
- 安裝路徑：`C:\Program Files\SysinfoAPI\`

## 使用方式

安裝完成後，開啟瀏覽器：

```
http://localhost:8088/
```

## API 端點

| 端點 | 說明 |
|------|------|
| `GET /` | 即時監控 Web 儀表板（含 MQTT 設定） |
| `GET /processes` | 程序監控頁面（支援分頁） |
| `GET /health` | 健康檢查端點 |
| `GET /api/system` | 系統資訊 JSON API |
| `GET /api/processes` | 程序列表 API（支援分頁） |
| `GET /api/history` | 歷史資料查詢（支援任意時段） |
| `GET /api/history/stats` | 歷史資料統計資訊 |
| `GET /api/mqtt/config` | 取得 MQTT 設定 |
| `POST /api/mqtt/config` | 儲存 MQTT 設定 |
| `GET /api/mqtt/status` | 取得 MQTT 連線狀態 |

### 歷史資料 API

```
GET /api/history?minutes=N
GET /api/history?start=<unix_timestamp>&end=<unix_timestamp>
GET /api/history?start=<unix_timestamp>&end=<unix_timestamp>&format=csv
```

| 參數 | 預設值 | 說明 |
|------|--------|------|
| `minutes` | 60 | 查詢最近 N 分鐘的資料（無上限） |
| `start` | - | 起始時間 Unix 時間戳 |
| `end` | 現在 | 結束時間 Unix 時間戳 |
| `format` | json | 輸出格式：`json` 或 `csv` |

**回應範例（JSON）：**
```json
{
  "interval_seconds": 30,
  "start_time": 1768706921,
  "end_time": 1768708721,
  "count": 180,
  "data": [
    {"ts": 1768708721, "cpu": 45.2, "mem": 60.5, "disk": 29.5},
    ...
  ]
}
```

**回應範例（CSV）：**
```csv
timestamp,datetime,cpu_percent,mem_percent,disk_percent
1768708721,2026-01-18 10:30:21,45.20,60.50,29.50
1768708731,2026-01-18 10:30:31,42.10,60.80,29.50
...
```

### 歷史統計 API

```
GET /api/history/stats
```

**回應範例：**
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

### 使用範例

```bash
# 查詢最近 30 分鐘資料
curl "http://localhost:8088/api/history?minutes=30"

# 查詢最近 24 小時資料
curl "http://localhost:8088/api/history?minutes=1440"

# 查詢指定時間範圍
curl "http://localhost:8088/api/history?start=1768622321&end=1768708721"

# 下載 CSV 檔案
curl -o history.csv "http://localhost:8088/api/history?minutes=60&format=csv"

# 查看歷史資料統計
curl "http://localhost:8088/api/history/stats"
```

### 系統資訊 API 回應範例

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

## 儀表板功能

| 區塊 | 說明 |
|------|------|
| **HOST** | 主機名稱、作業系統、平台、運行時間 |
| **CPU** | 處理器型號、各核心使用率進度條、趨勢圖 |
| **MEMORY** | 總計/已用/可用記憶體、使用率、趨勢圖 |
| **DISK** | 總計/已用/可用磁碟空間、使用率 |
| **TEMPERATURE** | 感測器溫度（彩色標示） |

### 溫度顏色標示

| 顏色 | 溫度範圍 | 狀態 |
|------|----------|------|
| 藍色 | < 30°C | 低溫 |
| 綠色 | 30-50°C | 正常 |
| 橙色 | 50-70°C | 偏高 |
| 紅色 | > 70°C | 過熱 |

## 程序監控

獨立頁面可檢視所有執行中的程序，路徑為 `/processes`。

### 功能

- 依 CPU 使用率排序顯示所有系統程序
- 支援分頁瀏覽（每頁 50 筆）
- 手動更新（點擊 Refresh 按鈕）
- 顯示欄位：PID、名稱、CPU%、記憶體%、狀態、使用者

### 程序 API

```
GET /api/processes?page=1&limit=50
```

| 參數 | 預設值 | 說明 |
|------|--------|------|
| `page` | 1 | 頁碼 |
| `limit` | 50 | 每頁筆數（最大 200） |

**回應範例：**
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

## MQTT 整合

儀表板內建 MQTT 設定介面，可將系統指標發布至 MQTT Broker。

### 設定欄位

透過 Web 儀表板或 API 設定：

| 欄位 | 說明 |
|------|------|
| **Broker** | MQTT Broker 位址（例如 `tcp://broker.example.com:1883`） |
| **Client ID** | 自訂裝置識別名稱（留空則使用主機名稱） |
| **Username** | 認證帳號（選填） |
| **Password** | 認證密碼（選填） |

### MQTT 訊息格式

**Topic：** `sysinfo/{client_id}`

**Payload（每 30 秒發送）：**
```json
{
  "hostname": "my-device",
  "cpu": 45.2,
  "mem": 60.5,
  "disk": 29.5,
  "timestamp": 1737200000
}
```

### MQTT API 範例

```bash
# 取得目前設定
curl http://localhost:8088/api/mqtt/config

# 儲存設定
curl -X POST http://localhost:8088/api/mqtt/config \
  -H "Content-Type: application/json" \
  -d '{"enabled":true,"broker":"tcp://broker:1883","client_id":"my-device","username":"","password":"","topic_prefix":"sysinfo"}'

# 查看連線狀態
curl http://localhost:8088/api/mqtt/status
```

### 設定檔

MQTT 設定儲存於 `mqtt_config.json`：

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

## 手動編譯

### 前置需求

- Go 1.20 或更新版本

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

## 服務管理

### Linux (systemd)

```bash
sudo systemctl status sysinfo-api   # 查看狀態
sudo systemctl restart sysinfo-api  # 重啟服務
sudo systemctl stop sysinfo-api     # 停止服務
sudo systemctl start sysinfo-api    # 啟動服務
sudo systemctl enable sysinfo-api   # 啟用開機自動啟動
sudo systemctl disable sysinfo-api  # 停用開機自動啟動
```

### Windows

```powershell
Get-Service SysinfoAPI              # 查看狀態
Restart-Service SysinfoAPI          # 重啟服務
Stop-Service SysinfoAPI             # 停止服務
Start-Service SysinfoAPI            # 啟動服務
```

## 設定

預設連接埠：**8088**

如需更改連接埠，請修改原始碼後重新編譯。

## 授權

MIT License
