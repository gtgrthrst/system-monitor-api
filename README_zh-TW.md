# 系統監控 API

[English](README.md)

使用 Go 和 gopsutil 建立的輕量級跨平台系統監控 API，具有即時 Web 儀表板。

## 功能特色

- **即時儀表板** - 終端機風格 Web 介面，每 2 秒自動更新
- **趨勢圖表** - CPU 和記憶體使用率歷史視覺化（60 個數據點）
- **溫度監控** - 彩色標示的感測器溫度顯示
- **跨平台支援** - 支援 Linux 和 Windows
- **Windows 服務** - MSI 安裝程式支援開機自動啟動
- **低資源消耗** - 極低的 CPU 和記憶體使用量

## 截圖

### Linux 儀表板
<img width="818" alt="Linux Dashboard" src="https://github.com/user-attachments/assets/b84004a9-ae34-4241-90eb-33b1f963719e" />

### Windows 儀表板
<img width="834" alt="Windows Dashboard" src="https://github.com/user-attachments/assets/8308422d-8cf8-41c9-b1f5-6bfe1f423b97" />

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
| `GET /` | 即時監控 Web 儀表板 |
| `GET /health` | 健康檢查端點 |
| `GET /api/system` | 系統資訊 JSON API |

### API 回應範例

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
