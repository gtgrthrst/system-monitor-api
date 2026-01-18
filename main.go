package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/kardianos/service"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

// History configuration - optimized for low resource usage
const (
	historyInterval = 10 * time.Second // Collect every 10 seconds
	historyMaxSize  = 360              // 1 hour of data (360 * 10s = 3600s)
)

// HistoryPoint stores minimal data for each time point
type HistoryPoint struct {
	Timestamp  int64   `json:"ts"`          // Unix timestamp
	CPUPercent float64 `json:"cpu"`         // CPU average %
	MemPercent float64 `json:"mem"`         // Memory %
	DiskPercent float64 `json:"disk"`       // Disk %
}

// RingBuffer is a fixed-size circular buffer for history data
type RingBuffer struct {
	data  []HistoryPoint
	head  int
	count int
	mu    sync.RWMutex
}

func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		data: make([]HistoryPoint, size),
	}
}

func (rb *RingBuffer) Push(p HistoryPoint) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	rb.data[rb.head] = p
	rb.head = (rb.head + 1) % len(rb.data)
	if rb.count < len(rb.data) {
		rb.count++
	}
}

func (rb *RingBuffer) GetAll() []HistoryPoint {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	result := make([]HistoryPoint, rb.count)
	for i := 0; i < rb.count; i++ {
		idx := (rb.head - rb.count + i + len(rb.data)) % len(rb.data)
		result[i] = rb.data[idx]
	}
	return result
}

func (rb *RingBuffer) GetSince(since int64) []HistoryPoint {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	var result []HistoryPoint
	for i := 0; i < rb.count; i++ {
		idx := (rb.head - rb.count + i + len(rb.data)) % len(rb.data)
		if rb.data[idx].Timestamp >= since {
			result = append(result, rb.data[idx])
		}
	}
	return result
}

var historyBuffer = NewRingBuffer(historyMaxSize)

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
.bar-temp { background: linear-gradient(90deg, #a00, #f00); }
.temp-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 8px; }
.temp-item { background: #1a1a1a; padding: 8px; border-radius: 3px; }
.temp-value { font-size: 18px; font-weight: bold; }
.temp-cold { color: #0af; }
.temp-normal { color: #0f0; }
.temp-warm { color: #fa0; }
.temp-hot { color: #f00; }
.bar-text { margin-top: 2px; font-size: 12px; color: #666; }
.cpu-cores { display: grid; grid-template-columns: repeat(auto-fill, minmax(180px, 1fr)); gap: 8px; }
.core { background: #1a1a1a; padding: 8px; border-radius: 3px; }
.update-time { color: #444; font-size: 11px; text-align: right; margin-top: 10px; }
.chart { background: #0a0a0a; border: 1px solid #333; border-radius: 3px; margin: 10px 0; }
.chart-label { display: flex; justify-content: space-between; font-size: 11px; color: #555; padding: 0 5px; }
</style>
</head>
<body>
<div class="container">
<h1>[ System Monitor ]</h1>
<div id="content">Loading...</div>
<div class="update-time">Refresh: 2s | History: 60 points</div>
</div>
<script>
const MAX_POINTS = 60;
let cpuHistory = [];
let memHistory = [];

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
function drawChart(canvasId, data, color, fillColor) {
  const canvas = document.getElementById(canvasId);
  if (!canvas) return;
  const ctx = canvas.getContext('2d');
  const w = canvas.width, h = canvas.height;
  ctx.clearRect(0, 0, w, h);

  // Grid lines
  ctx.strokeStyle = '#222';
  ctx.lineWidth = 1;
  for (let i = 0; i <= 4; i++) {
    const y = (h / 4) * i;
    ctx.beginPath();
    ctx.moveTo(0, y);
    ctx.lineTo(w, y);
    ctx.stroke();
  }

  if (data.length < 2) return;

  // Fill area
  ctx.beginPath();
  ctx.moveTo(0, h);
  data.forEach((v, i) => {
    const x = (i / (MAX_POINTS - 1)) * w;
    const y = h - (v / 100) * h;
    if (i === 0) ctx.lineTo(x, y);
    else ctx.lineTo(x, y);
  });
  ctx.lineTo(((data.length - 1) / (MAX_POINTS - 1)) * w, h);
  ctx.closePath();
  ctx.fillStyle = fillColor;
  ctx.fill();

  // Line
  ctx.beginPath();
  ctx.strokeStyle = color;
  ctx.lineWidth = 2;
  data.forEach((v, i) => {
    const x = (i / (MAX_POINTS - 1)) * w;
    const y = h - (v / 100) * h;
    if (i === 0) ctx.moveTo(x, y);
    else ctx.lineTo(x, y);
  });
  ctx.stroke();
}

function update() {
  fetch('/api/system').then(r => r.json()).then(d => {
    let cpuAvg = d.cpu.usage_percent.reduce((a,b) => a+b, 0) / d.cpu.usage_percent.length;

    // Update history
    cpuHistory.push(cpuAvg);
    memHistory.push(d.memory.used_percent);
    if (cpuHistory.length > MAX_POINTS) cpuHistory.shift();
    if (memHistory.length > MAX_POINTS) memHistory.shift();

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
      '<canvas id="cpuChart" class="chart" width="740" height="80"></canvas>' +
      '<div class="chart-label"><span>2 min ago</span><span>now</span></div>' +
      '<div class="cpu-cores">' + cores + '</div></div>' +
      '<div class="section"><div class="section-title">MEMORY</div>' +
      '<div class="row"><span class="label">Used</span><span class="value">' + formatBytes(d.memory.used_bytes) + ' / ' + formatBytes(d.memory.total_bytes) + '</span></div>' +
      '<div class="bar-container"><div class="bar bar-mem" style="width:' + d.memory.used_percent + '%"></div></div>' +
      '<canvas id="memChart" class="chart" width="740" height="80"></canvas>' +
      '<div class="chart-label"><span>2 min ago</span><span>now</span></div>' +
      '<div class="bar-text">' + d.memory.used_percent.toFixed(1) + '% used</div></div>' +
      '<div class="section"><div class="section-title">DISK /</div>' +
      '<div class="row"><span class="label">Used</span><span class="value">' + formatBytes(d.disk.used_bytes) + ' / ' + formatBytes(d.disk.total_bytes) + '</span></div>' +
      '<div class="bar-container"><div class="bar bar-disk" style="width:' + d.disk.used_percent + '%"></div></div>' +
      '<div class="bar-text">' + d.disk.used_percent.toFixed(1) + '% used</div></div>' +
      (d.temperature && d.temperature.length > 0 ?
        '<div class="section"><div class="section-title">TEMPERATURE</div><div class="temp-grid">' +
        d.temperature.map(t => {
          let cls = 'temp-normal';
          if (t.temperature < 30) cls = 'temp-cold';
          else if (t.temperature > 70) cls = 'temp-hot';
          else if (t.temperature > 50) cls = 'temp-warm';
          return '<div class="temp-item"><div class="row"><span class="label">' + t.name + '</span></div>' +
            '<div class="temp-value ' + cls + '">' + t.temperature.toFixed(1) + '°C</div></div>';
        }).join('') + '</div></div>' : '');

    // Draw charts after DOM update
    setTimeout(() => {
      drawChart('cpuChart', cpuHistory, '#0f0', 'rgba(0,255,0,0.1)');
      drawChart('memChart', memHistory, '#f0f', 'rgba(255,0,255,0.1)');
    }, 0);
  }).catch(e => { document.getElementById('content').innerHTML = '<div style="color:red">Error: ' + e + '</div>'; });
}
update();
setInterval(update, 2000);
</script>
</body>
</html>`

type SystemInfo struct {
	Host        HostInfo      `json:"host"`
	CPU         CPUInfo       `json:"cpu"`
	Memory      MemoryInfo    `json:"memory"`
	Disk        DiskInfo      `json:"disk"`
	Temperature []TempInfo    `json:"temperature"`
}

type TempInfo struct {
	Name        string  `json:"name"`
	Temperature float64 `json:"temperature"`
}

type HostInfo struct {
	Hostname string `json:"hostname"`
	OS       string `json:"os"`
	Platform string `json:"platform"`
	Uptime   uint64 `json:"uptime_seconds"`
}

type CPUInfo struct {
	Cores        int       `json:"cores"`
	ModelName    string    `json:"model_name"`
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
	hostInfo, err := host.Info()
	if err != nil {
		return nil, err
	}

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

	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	diskPath := "/"
	if hostInfo.OS == "windows" {
		diskPath = "C:"
	}
	diskInfo, err := disk.Usage(diskPath)
	if err != nil {
		return nil, err
	}

	// Temperature sensors
	var temps []TempInfo
	sensors, _ := host.SensorsTemperatures()
	for _, s := range sensors {
		if s.Temperature > 0 {
			temps = append(temps, TempInfo{
				Name:        s.SensorKey,
				Temperature: s.Temperature,
			})
		}
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
		Temperature: temps,
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

// handleHistory returns historical data
// Query params: minutes (default: 60, max: 60)
func handleHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	minutes := 60
	if m := r.URL.Query().Get("minutes"); m != "" {
		if v, err := strconv.Atoi(m); err == nil && v > 0 && v <= 60 {
			minutes = v
		}
	}

	since := time.Now().Add(-time.Duration(minutes) * time.Minute).Unix()
	data := historyBuffer.GetSince(since)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"interval_seconds": int(historyInterval.Seconds()),
		"max_minutes":      60,
		"requested_minutes": minutes,
		"count":            len(data),
		"data":             data,
	})
}

// collectHistory runs in background to collect system metrics
func collectHistory() {
	ticker := time.NewTicker(historyInterval)
	defer ticker.Stop()

	for range ticker.C {
		info, err := getSystemInfo()
		if err != nil {
			continue
		}

		// Calculate CPU average
		var cpuAvg float64
		if len(info.CPU.UsagePercent) > 0 {
			for _, v := range info.CPU.UsagePercent {
				cpuAvg += v
			}
			cpuAvg /= float64(len(info.CPU.UsagePercent))
		}

		historyBuffer.Push(HistoryPoint{
			Timestamp:   time.Now().Unix(),
			CPUPercent:  cpuAvg,
			MemPercent:  info.Memory.UsedPercent,
			DiskPercent: info.Disk.UsedPercent,
		})
	}
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(dashboardHTML))
}

// Service wrapper for Windows service support
type program struct{}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) run() {
	// Start history collector in background
	go collectHistory()

	http.HandleFunc("/", handleDashboard)
	http.HandleFunc("/api/system", handleSystemInfo)
	http.HandleFunc("/api/history", handleHistory)
	http.HandleFunc("/health", handleHealth)
	log.Println("Server starting on :8088...")
	log.Printf("History: collecting every %v, max %d points (1 hour)\n", historyInterval, historyMaxSize)
	http.ListenAndServe(":8088", nil)
}

func (p *program) Stop(s service.Service) error {
	return nil
}

func main() {
	svcConfig := &service.Config{
		Name:        "SysinfoAPI",
		DisplayName: "System Monitor API",
		Description: "System Monitor API Service",
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}

	err = s.Run()
	if err != nil {
		log.Fatal(err)
	}
}
