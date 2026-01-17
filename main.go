package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

const dashboardHTML = `<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>System Monitor</title>
<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body { background: #0a0a0a; color: #00ff00; font-family: 'Courier New', monospace; font-size: 14px; padding: 20px; }
.container { max-width: 800px; margin: 0 auto; }
h1 { color: #00ff00; border-bottom: 1px solid #333; padding-bottom: 10px; margin-bottom: 20px; font-size: 18px; }
.section { background: #111; border: 1px solid #333; margin-bottom: 15px; padding: 15px; border-radius: 4px; }
.section-title { color: #0af; font-weight: bold; margin-bottom: 10px; }
.row { display: flex; justify-content: space-between; padding: 3px 0; }
.label { color: #888; }
.value { color: #0f0; }
.bar-container { background: #222; height: 20px; border-radius: 3px; overflow: hidden; margin: 5px 0; }
.bar { height: 100%; transition: width 0.3s; }
.bar-cpu { background: linear-gradient(90deg, #0a0, #0f0); }
.bar-mem { background: linear-gradient(90deg, #a0a, #f0f); }
.bar-disk { background: linear-gradient(90deg, #aa0, #ff0); }
.bar-text { margin-top: 2px; font-size: 12px; color: #666; }
.cpu-cores { display: grid; grid-template-columns: repeat(auto-fill, minmax(180px, 1fr)); gap: 8px; }
.core { background: #1a1a1a; padding: 8px; border-radius: 3px; }
.update-time { color: #444; font-size: 11px; text-align: right; margin-top: 10px; }
</style>
</head>
<body>
<div class="container">
<h1>[ System Monitor ]</h1>
<div id="content">Loading...</div>
<div class="update-time">Refresh: 2s</div>
</div>
<script>
function formatBytes(b) {
  const u = ['B', 'KB', 'MB', 'GB', 'TB'];
  let i = 0;
  while (b >= 1024 && i < u.length - 1) { b /= 1024; i++; }
  return b.toFixed(1) + ' ' + u[i];
}
function formatUptime(s) {
  const d = Math.floor(s / 86400);
  const h = Math.floor((s % 86400) / 3600);
  const m = Math.floor((s % 3600) / 60);
  return d + 'd ' + h + 'h ' + m + 'm';
}
function update() {
  fetch('/api/system').then(r => r.json()).then(d => {
    let cpuAvg = d.cpu.usage_percent.reduce((a,b) => a+b, 0) / d.cpu.usage_percent.length;
    let cores = d.cpu.usage_percent.map((p, i) =>
      '<div class="core"><div class="row"><span class="label">Core ' + i + '</span><span class="value">' + p.toFixed(1) + '%</span></div>' +
      '<div class="bar-container"><div class="bar bar-cpu" style="width:' + p + '%"></div></div></div>'
    ).join('');
    document.getElementById('content').innerHTML =
      '<div class="section"><div class="section-title">HOST</div>' +
      '<div class="row"><span class="label">Hostname</span><span class="value">' + d.host.hostname + '</span></div>' +
      '<div class="row"><span class="label">OS</span><span class="value">' + d.host.platform + ' (' + d.host.os + ')</span></div>' +
      '<div class="row"><span class="label">Uptime</span><span class="value">' + formatUptime(d.host.uptime_seconds) + '</span></div></div>' +
      '<div class="section"><div class="section-title">CPU - ' + d.cpu.model_name + '</div>' +
      '<div class="row"><span class="label">Average</span><span class="value">' + cpuAvg.toFixed(1) + '%</span></div>' +
      '<div class="bar-container"><div class="bar bar-cpu" style="width:' + cpuAvg + '%"></div></div>' +
      '<div class="cpu-cores">' + cores + '</div></div>' +
      '<div class="section"><div class="section-title">MEMORY</div>' +
      '<div class="row"><span class="label">Used</span><span class="value">' + formatBytes(d.memory.used_bytes) + ' / ' + formatBytes(d.memory.total_bytes) + '</span></div>' +
      '<div class="bar-container"><div class="bar bar-mem" style="width:' + d.memory.used_percent + '%"></div></div>' +
      '<div class="bar-text">' + d.memory.used_percent.toFixed(1) + '% used</div></div>' +
      '<div class="section"><div class="section-title">DISK /</div>' +
      '<div class="row"><span class="label">Used</span><span class="value">' + formatBytes(d.disk.used_bytes) + ' / ' + formatBytes(d.disk.total_bytes) + '</span></div>' +
      '<div class="bar-container"><div class="bar bar-disk" style="width:' + d.disk.used_percent + '%"></div></div>' +
      '<div class="bar-text">' + d.disk.used_percent.toFixed(1) + '% used</div></div>';
  }).catch(e => { document.getElementById('content').innerHTML = '<div style="color:red">Error: ' + e + '</div>'; });
}
update();
setInterval(update, 2000);
</script>
</body>
</html>`

type SystemInfo struct {
	Host   HostInfo   `json:"host"`
	CPU    CPUInfo    `json:"cpu"`
	Memory MemoryInfo `json:"memory"`
	Disk   DiskInfo   `json:"disk"`
}

type HostInfo struct {
	Hostname string `json:"hostname"`
	OS       string `json:"os"`
	Platform string `json:"platform"`
	Uptime   uint64 `json:"uptime_seconds"`
}

type CPUInfo struct {
	Cores       int       `json:"cores"`
	ModelName   string    `json:"model_name"`
	UsagePercent []float64 `json:"usage_percent"`
}

type MemoryInfo struct {
	Total       uint64  `json:"total_bytes"`
	Used        uint64  `json:"used_bytes"`
	Free        uint64  `json:"free_bytes"`
	UsedPercent float64 `json:"used_percent"`
}

type DiskInfo struct {
	Total       uint64  `json:"total_bytes"`
	Used        uint64  `json:"used_bytes"`
	Free        uint64  `json:"free_bytes"`
	UsedPercent float64 `json:"used_percent"`
}

func getSystemInfo() (*SystemInfo, error) {
	// Host info
	hostInfo, err := host.Info()
	if err != nil {
		return nil, err
	}

	// CPU info
	cpuInfo, err := cpu.Info()
	if err != nil {
		return nil, err
	}
	cpuPercent, err := cpu.Percent(0, true)
	if err != nil {
		return nil, err
	}

	modelName := ""
	if len(cpuInfo) > 0 {
		modelName = cpuInfo[0].ModelName
	}

	// Memory info
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	// Disk info
	diskInfo, err := disk.Usage("/")
	if err != nil {
		return nil, err
	}

	return &SystemInfo{
		Host: HostInfo{
			Hostname: hostInfo.Hostname,
			OS:       hostInfo.OS,
			Platform: hostInfo.Platform,
			Uptime:   hostInfo.Uptime,
		},
		CPU: CPUInfo{
			Cores:        len(cpuInfo),
			ModelName:    modelName,
			UsagePercent: cpuPercent,
		},
		Memory: MemoryInfo{
			Total:       memInfo.Total,
			Used:        memInfo.Used,
			Free:        memInfo.Free,
			UsedPercent: memInfo.UsedPercent,
		},
		Disk: DiskInfo{
			Total:       diskInfo.Total,
			Used:        diskInfo.Used,
			Free:        diskInfo.Free,
			UsedPercent: diskInfo.UsedPercent,
		},
	}, nil
}

func handleSystemInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	info, err := getSystemInfo()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(info)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(dashboardHTML))
}

func main() {
	http.HandleFunc("/", handleDashboard)
	http.HandleFunc("/api/system", handleSystemInfo)
	http.HandleFunc("/health", handleHealth)

	log.Println("Server starting on :8088...")
	log.Fatal(http.ListenAndServe(":8088", nil))
}
