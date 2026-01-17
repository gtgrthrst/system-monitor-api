# 系統監控 API

使用 Go 和 gopsutil 建立的簡易系統監控 JSON API。

## 一鍵安裝

```bash
curl -fsSL https://raw.githubusercontent.com/gtgrthrst/system-monitor-api/main/install.sh | sudo bash
```

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

```bash
go build -o sysinfo-api
./sysinfo-api
```

伺服器監聽 **8088** 連接埠。

## 服務管理

安裝後可使用 systemd 管理服務：

```bash
sudo systemctl status sysinfo-api   # 查看狀態
sudo systemctl restart sysinfo-api  # 重啟服務
sudo systemctl stop sysinfo-api     # 停止服務
sudo systemctl start sysinfo-api    # 啟動服務
```
