# 系統監控 API

使用 Go 和 gopsutil 建立的簡易系統監控 JSON API。


## 安裝

### Linux

```bash
curl -fsSL https://raw.githubusercontent.com/gtgrthrst/system-monitor-api/main/install.sh | sudo bash
```
<img width="818" height="866" alt="Snipaste_2026-01-17_21-47-35" src="https://github.com/user-attachments/assets/b84004a9-ae34-4241-90eb-33b1f963719e" />

### Windows (MSI 安裝程式)

從 [Releases](https://github.com/gtgrthrst/system-monitor-api/releases/latest) 下載：
- [sysinfo-api.msi](https://github.com/gtgrthrst/system-monitor-api/releases/latest/download/sysinfo-api.msi) - 自動安裝為 Windows 服務
- [sysinfo-api-windows-amd64.exe](https://github.com/gtgrthrst/system-monitor-api/releases/latest/download/sysinfo-api-windows-amd64.exe) - 獨立執行檔
<img width="834" height="731" alt="Snipaste_2026-01-17_21-47-14" src="https://github.com/user-attachments/assets/8308422d-8cf8-41c9-b1f5-6bfe1f423b97" />
## 端點

| 端點 | 說明 |
|------|------|
| `GET /` | Web 監控儀表板（類似 Glances） |
| `GET /health` | 健康檢查 |
| `GET /api/system` | 系統資訊（CPU、記憶體、磁碟、主機） |

## 回應範例

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
  }
}
```

## 手動編譯

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

伺服器監聽 **8088** 連接埠。

## 服務管理

### Linux (systemd)

```bash
sudo systemctl status sysinfo-api   # 查看狀態
sudo systemctl restart sysinfo-api  # 重啟服務
sudo systemctl stop sysinfo-api     # 停止服務
sudo systemctl start sysinfo-api    # 啟動服務
```

### Windows

```powershell
Get-Service SysinfoAPI              # 查看狀態
Restart-Service SysinfoAPI          # 重啟服務
Stop-Service SysinfoAPI             # 停止服務
Start-Service SysinfoAPI            # 啟動服務
```
