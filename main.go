package main

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/kardianos/service"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
	_ "github.com/mattn/go-sqlite3"
)

// History configuration - optimized for low resource usage
const (
	historyInterval = 30 * time.Second // Collect every 30 seconds (reduced from 10s)
	historyMaxSize  = 120              // 1 hour of data (120 * 30s = 3600s)
)

// Cache configuration for reducing CPU usage
const (
	sysInfoCacheTTL  = 3 * time.Second   // System info cache TTL
	hostInfoCacheTTL = 5 * time.Minute   // Host info rarely changes
	cpuCollectInterval = 2 * time.Second // Background CPU collection interval
)

// Feature flags
var enableTemperature = true // Set to false to disable temperature monitoring

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

// Database for persistent history storage
var db *sql.DB
var dbMutex sync.Mutex

// MQTT configuration and client
type MQTTConfig struct {
	Enabled     bool   `json:"enabled"`
	Broker      string `json:"broker"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	TopicPrefix string `json:"topic_prefix"`
	ClientID    string `json:"client_id"`
}

var (
	mqttConfig  MQTTConfig
	mqttClient  mqtt.Client
	mqttMutex   sync.RWMutex
	mqttConnected bool
)

// getMQTTConfigPath returns the path to the MQTT config file
func getMQTTConfigPath() string {
	return filepath.Join(getDataDir(), "mqtt_config.json")
}

// loadMQTTConfig loads MQTT configuration from file
func loadMQTTConfig() error {
	mqttMutex.Lock()
	defer mqttMutex.Unlock()

	configPath := getMQTTConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create default config
			mqttConfig = MQTTConfig{
				Enabled:     false,
				Broker:      "tcp://localhost:1883",
				TopicPrefix: "sysinfo",
				ClientID:    "",
			}
			return saveMQTTConfigLocked()
		}
		return err
	}

	return json.Unmarshal(data, &mqttConfig)
}

// saveMQTTConfig saves MQTT configuration to file
func saveMQTTConfig() error {
	mqttMutex.Lock()
	defer mqttMutex.Unlock()
	return saveMQTTConfigLocked()
}

// saveMQTTConfigLocked saves config (must hold mqttMutex)
func saveMQTTConfigLocked() error {
	configPath := getMQTTConfigPath()
	data, err := json.MarshalIndent(mqttConfig, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0600)
}

// getEffectiveClientID returns the client ID to use (custom or hostname)
func getEffectiveClientID() string {
	mqttMutex.RLock()
	defer mqttMutex.RUnlock()

	if mqttConfig.ClientID != "" {
		return mqttConfig.ClientID
	}
	hostInfo, err := host.Info()
	if err != nil {
		return "unknown"
	}
	return hostInfo.Hostname
}

// connectMQTT establishes connection to MQTT broker
func connectMQTT() {
	mqttMutex.Lock()
	defer mqttMutex.Unlock()

	if !mqttConfig.Enabled {
		mqttConnected = false
		return
	}

	// Disconnect existing client if any
	if mqttClient != nil && mqttClient.IsConnected() {
		mqttClient.Disconnect(250)
	}

	clientID := mqttConfig.ClientID
	if clientID == "" {
		hostInfo, _ := host.Info()
		if hostInfo != nil {
			clientID = hostInfo.Hostname
		} else {
			clientID = "sysinfo-api"
		}
	}

	opts := mqtt.NewClientOptions().
		AddBroker(mqttConfig.Broker).
		SetClientID(clientID + "-publisher").
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(10 * time.Second).
		SetOnConnectHandler(func(c mqtt.Client) {
			log.Printf("MQTT connected to %s\n", mqttConfig.Broker)
			mqttMutex.Lock()
			mqttConnected = true
			mqttMutex.Unlock()
		}).
		SetConnectionLostHandler(func(c mqtt.Client, err error) {
			log.Printf("MQTT connection lost: %v\n", err)
			mqttMutex.Lock()
			mqttConnected = false
			mqttMutex.Unlock()
		})

	if mqttConfig.Username != "" {
		opts.SetUsername(mqttConfig.Username)
		opts.SetPassword(mqttConfig.Password)
	}

	mqttClient = mqtt.NewClient(opts)
	token := mqttClient.Connect()
	go func() {
		if token.Wait() && token.Error() != nil {
			log.Printf("MQTT connection error: %v\n", token.Error())
		}
	}()
}

// disconnectMQTT closes the MQTT connection
func disconnectMQTT() {
	mqttMutex.Lock()
	defer mqttMutex.Unlock()

	if mqttClient != nil && mqttClient.IsConnected() {
		mqttClient.Disconnect(250)
	}
	mqttConnected = false
}

// publishMetrics publishes current metrics to MQTT
func publishMetrics(point HistoryPoint) {
	mqttMutex.RLock()
	enabled := mqttConfig.Enabled
	topicPrefix := mqttConfig.TopicPrefix
	mqttMutex.RUnlock()

	if !enabled || mqttClient == nil || !mqttClient.IsConnected() {
		return
	}

	clientID := getEffectiveClientID()
	topic := fmt.Sprintf("%s/%s", topicPrefix, clientID)

	// Get system uptime
	var uptime uint64
	if hostInfo, err := getCachedHostInfo(); err == nil {
		uptime = hostInfo.Uptime
	}

	payload := map[string]interface{}{
		"hostname": clientID,
		"cpu":      point.CPUPercent,
		"mem":      point.MemPercent,
		"disk":     point.DiskPercent,
		"uptime":   uptime,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("MQTT marshal error: %v\n", err)
		return
	}

	token := mqttClient.Publish(topic, 0, false, data)
	go func() {
		if token.Wait() && token.Error() != nil {
			log.Printf("MQTT publish error: %v\n", token.Error())
		}
	}()
}

// getDataDir returns the directory for storing data files
func getDataDir() string {
	// Try to use the directory where the executable is located
	exe, err := os.Executable()
	if err == nil {
		return filepath.Dir(exe)
	}
	// Fallback to current directory
	return "."
}

// initDB initializes the SQLite database
func initDB() error {
	dbPath := filepath.Join(getDataDir(), "sysinfo_history.db")
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Create table if not exists
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp INTEGER NOT NULL,
		cpu_percent REAL NOT NULL,
		mem_percent REAL NOT NULL,
		disk_percent REAL NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_history_timestamp ON history(timestamp);
	`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	log.Printf("Database initialized: %s\n", dbPath)
	return nil
}

// saveHistoryToDB saves a history point to the database
func saveHistoryToDB(p HistoryPoint) error {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	_, err := db.Exec(
		"INSERT INTO history (timestamp, cpu_percent, mem_percent, disk_percent) VALUES (?, ?, ?, ?)",
		p.Timestamp, p.CPUPercent, p.MemPercent, p.DiskPercent,
	)
	return err
}

// queryHistoryFromDB queries history from database with time range
func queryHistoryFromDB(startTime, endTime int64) ([]HistoryPoint, error) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	rows, err := db.Query(
		"SELECT timestamp, cpu_percent, mem_percent, disk_percent FROM history WHERE timestamp >= ? AND timestamp <= ? ORDER BY timestamp ASC",
		startTime, endTime,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []HistoryPoint
	for rows.Next() {
		var p HistoryPoint
		if err := rows.Scan(&p.Timestamp, &p.CPUPercent, &p.MemPercent, &p.DiskPercent); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, nil
}

// getHistoryStats returns statistics about stored history
func getHistoryStats() (minTime, maxTime int64, count int64, err error) {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	if db == nil {
		return 0, 0, 0, fmt.Errorf("database not initialized")
	}

	err = db.QueryRow("SELECT COALESCE(MIN(timestamp), 0), COALESCE(MAX(timestamp), 0), COUNT(*) FROM history").Scan(&minTime, &maxTime, &count)
	return
}

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
.mqtt-form { display: grid; gap: 10px; }
.mqtt-row { display: flex; align-items: center; gap: 10px; }
.mqtt-row label { width: 100px; color: #888; }
.mqtt-row input { flex: 1; background: #222; border: 1px solid #444; color: #0f0; padding: 8px; border-radius: 3px; font-family: inherit; }
.mqtt-row input:focus { outline: none; border-color: #0af; }
.mqtt-topic { color: #666; font-size: 12px; margin-top: 5px; }
.mqtt-actions { display: flex; gap: 10px; align-items: center; margin-top: 10px; }
.mqtt-toggle { display: flex; align-items: center; gap: 8px; cursor: pointer; }
.mqtt-toggle input { display: none; }
.mqtt-toggle .slider { width: 40px; height: 20px; background: #333; border-radius: 10px; position: relative; transition: 0.3s; }
.mqtt-toggle .slider:before { content: ''; position: absolute; width: 16px; height: 16px; background: #666; border-radius: 50%; top: 2px; left: 2px; transition: 0.3s; }
.mqtt-toggle input:checked + .slider { background: #0a0; }
.mqtt-toggle input:checked + .slider:before { left: 22px; background: #0f0; }
.mqtt-btn { background: #0af; color: #000; border: none; padding: 8px 16px; border-radius: 3px; cursor: pointer; font-family: inherit; font-weight: bold; }
.mqtt-btn:hover { background: #0cf; }
.mqtt-btn:disabled { background: #444; color: #666; cursor: not-allowed; }
.mqtt-status { display: flex; align-items: center; gap: 8px; margin-left: auto; }
.mqtt-status .dot { width: 10px; height: 10px; border-radius: 50%; }
.mqtt-status .dot.connected { background: #0f0; box-shadow: 0 0 5px #0f0; }
.mqtt-status .dot.disconnected { background: #f00; box-shadow: 0 0 5px #f00; }
.mqtt-status .dot.disabled { background: #666; }
.gauges-container { display: flex; justify-content: space-around; flex-wrap: wrap; gap: 15px; margin-bottom: 15px; }
.gauge-card { background: #111; border: 1px solid #333; border-radius: 4px; padding: 15px 10px 10px; text-align: center; min-width: 160px; flex: 1; max-width: 200px; }
.gauge-svg { display: block; margin: 0 auto; }
.gauge-percent { font-size: 18px; font-weight: bold; font-family: 'Courier New', monospace; }
.gauge-label { font-size: 11px; font-family: 'Courier New', monospace; }
.gauge-track { fill: none; stroke: #222; stroke-width: 10; stroke-linecap: round; }
.gauge-value { fill: none; stroke-width: 10; stroke-linecap: round; transition: stroke-dashoffset 0.5s ease-out; }
.gauge-cpu { stroke: #0f0; }
.gauge-mem { stroke: #f0f; }
.gauge-disk { stroke: #ff0; }
.sparkline { margin-top: 8px; }
.spark-line { fill: none; stroke-width: 1.5; }
.spark-area { opacity: 0.15; }
.spark-dot { transition: cx 0.3s, cy 0.3s; }
@media (max-width: 600px) { .gauges-container { flex-direction: column; align-items: center; } .gauge-card { width: 100%; max-width: 250px; } }
</style>
</head>
<body>
<div class="container">
<h1>[ System Monitor ]</h1>
<div class="gauges-container">
  <div class="gauge-card">
    <svg id="cpu-gauge" class="gauge-svg" viewBox="0 0 120 70" width="120" height="70">
      <path class="gauge-track" d="M 10 60 A 50 50 0 0 1 110 60"/>
      <path class="gauge-value gauge-cpu" d="M 10 60 A 50 50 0 0 1 110 60" stroke-dasharray="157" stroke-dashoffset="157"/>
      <text x="60" y="52" text-anchor="middle" class="gauge-percent" fill="#0f0">--%</text>
      <text x="60" y="67" text-anchor="middle" class="gauge-label" fill="#888">CPU</text>
    </svg>
    <svg id="cpu-spark" class="sparkline" viewBox="0 0 100 25" width="100" height="25">
      <polygon class="spark-area" fill="#0f0" points="0,25 100,25"/>
      <polyline class="spark-line" stroke="#0f0" points="0,25"/>
      <circle class="spark-dot" fill="#0f0" r="2" cx="0" cy="25"/>
    </svg>
  </div>
  <div class="gauge-card">
    <svg id="mem-gauge" class="gauge-svg" viewBox="0 0 120 70" width="120" height="70">
      <path class="gauge-track" d="M 10 60 A 50 50 0 0 1 110 60"/>
      <path class="gauge-value gauge-mem" d="M 10 60 A 50 50 0 0 1 110 60" stroke-dasharray="157" stroke-dashoffset="157"/>
      <text x="60" y="52" text-anchor="middle" class="gauge-percent" fill="#f0f">--%</text>
      <text x="60" y="67" text-anchor="middle" class="gauge-label" fill="#888">MEMORY</text>
    </svg>
    <svg id="mem-spark" class="sparkline" viewBox="0 0 100 25" width="100" height="25">
      <polygon class="spark-area" fill="#f0f" points="0,25 100,25"/>
      <polyline class="spark-line" stroke="#f0f" points="0,25"/>
      <circle class="spark-dot" fill="#f0f" r="2" cx="0" cy="25"/>
    </svg>
  </div>
  <div class="gauge-card">
    <svg id="disk-gauge" class="gauge-svg" viewBox="0 0 120 70" width="120" height="70">
      <path class="gauge-track" d="M 10 60 A 50 50 0 0 1 110 60"/>
      <path class="gauge-value gauge-disk" d="M 10 60 A 50 50 0 0 1 110 60" stroke-dasharray="157" stroke-dashoffset="157"/>
      <text x="60" y="52" text-anchor="middle" class="gauge-percent" fill="#ff0">--%</text>
      <text x="60" y="67" text-anchor="middle" class="gauge-label" fill="#888">DISK</text>
    </svg>
    <svg id="disk-spark" class="sparkline" viewBox="0 0 100 25" width="100" height="25">
      <polygon class="spark-area" fill="#ff0" points="0,25 100,25"/>
      <polyline class="spark-line" stroke="#ff0" points="0,25"/>
      <circle class="spark-dot" fill="#ff0" r="2" cx="0" cy="25"/>
    </svg>
  </div>
</div>
<div id="content">Loading...</div>
<div id="mqtt-section" class="section">
  <div class="section-title">MQTT SETTINGS <span id="mqtt-status-indicator" class="mqtt-status"><span class="dot disabled"></span><span id="mqtt-status-text">Disabled</span></span></div>
  <div class="mqtt-form">
    <div class="mqtt-row"><label>Broker</label><input type="text" id="mqtt-broker" placeholder="tcp://broker.example.com:1883"></div>
    <div class="mqtt-row"><label>Client ID</label><input type="text" id="mqtt-client-id" placeholder="(auto: hostname)"></div>
    <div class="mqtt-row"><label>Username</label><input type="text" id="mqtt-username" placeholder="(optional)"></div>
    <div class="mqtt-row"><label>Password</label><input type="password" id="mqtt-password" placeholder="(optional)"></div>
    <div class="mqtt-topic">Topic: <span id="mqtt-topic-preview">sysinfo/...</span></div>
    <div class="mqtt-actions">
      <label class="mqtt-toggle"><input type="checkbox" id="mqtt-enabled"><span class="slider"></span><span>Enable MQTT</span></label>
      <button class="mqtt-btn" id="mqtt-save">Save Settings</button>
    </div>
  </div>
</div>
<div class="update-time"><a href="/processes" style="color:#0af;text-decoration:none">View Processes →</a> | Refresh: 5s | History: 120 points (30s interval)</div>
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
setInterval(update, 5000);

// MQTT Configuration
let mqttConfig = {};
let systemHostname = '';

function loadMQTTConfig() {
  fetch('/api/mqtt/config').then(r => r.json()).then(cfg => {
    mqttConfig = cfg;
    document.getElementById('mqtt-broker').value = cfg.broker || '';
    document.getElementById('mqtt-client-id').value = cfg.client_id || '';
    document.getElementById('mqtt-username').value = cfg.username || '';
    document.getElementById('mqtt-password').value = cfg.password === '***' ? '***' : '';
    document.getElementById('mqtt-enabled').checked = cfg.enabled;
    updateTopicPreview();
  });
}

function loadMQTTStatus() {
  fetch('/api/mqtt/status').then(r => r.json()).then(st => {
    const dot = document.querySelector('#mqtt-status-indicator .dot');
    const text = document.getElementById('mqtt-status-text');
    dot.className = 'dot ' + st.status;
    if (st.status === 'connected') {
      text.textContent = 'Connected';
    } else if (st.status === 'disconnected') {
      text.textContent = 'Disconnected';
    } else {
      text.textContent = 'Disabled';
    }
    document.getElementById('mqtt-topic-preview').textContent = st.topic;
  });
}

function updateTopicPreview() {
  const clientId = document.getElementById('mqtt-client-id').value || systemHostname || '(hostname)';
  const prefix = 'sysinfo';
  document.getElementById('mqtt-topic-preview').textContent = prefix + '/' + clientId;
}

function saveMQTTConfig() {
  const btn = document.getElementById('mqtt-save');
  btn.disabled = true;
  btn.textContent = 'Saving...';

  const cfg = {
    enabled: document.getElementById('mqtt-enabled').checked,
    broker: document.getElementById('mqtt-broker').value,
    client_id: document.getElementById('mqtt-client-id').value,
    username: document.getElementById('mqtt-username').value,
    password: document.getElementById('mqtt-password').value,
    topic_prefix: 'sysinfo'
  };

  fetch('/api/mqtt/config', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify(cfg)
  }).then(r => r.json()).then(() => {
    btn.textContent = 'Saved!';
    setTimeout(() => { btn.textContent = 'Save Settings'; btn.disabled = false; }, 1500);
    setTimeout(loadMQTTStatus, 1000);
  }).catch(e => {
    btn.textContent = 'Error!';
    setTimeout(() => { btn.textContent = 'Save Settings'; btn.disabled = false; }, 1500);
  });
}

// Get system hostname for topic preview
fetch('/api/system').then(r => r.json()).then(d => {
  systemHostname = d.host.hostname;
  updateTopicPreview();
});

document.getElementById('mqtt-save').addEventListener('click', saveMQTTConfig);
document.getElementById('mqtt-client-id').addEventListener('input', updateTopicPreview);

loadMQTTConfig();
loadMQTTStatus();
setInterval(loadMQTTStatus, 5000);
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

// ProcessInfo represents information about a single process
type ProcessInfo struct {
	PID        int32   `json:"pid"`
	Name       string  `json:"name"`
	CPUPercent float64 `json:"cpu_percent"`
	MemPercent float32 `json:"mem_percent"`
	Status     string  `json:"status"`
	Username   string  `json:"username"`
}

// ProcessListResponse is the API response for process list
type ProcessListResponse struct {
	Total      int           `json:"total"`
	Page       int           `json:"page"`
	Limit      int           `json:"limit"`
	TotalPages int           `json:"total_pages"`
	Timestamp  int64         `json:"timestamp"`
	Processes  []ProcessInfo `json:"processes"`
}

// Process list cache to reduce CPU usage
var (
	processCache      []ProcessInfo
	processCacheTime  time.Time
	processCacheMutex sync.RWMutex
	processCacheTTL   = 15 * time.Second // Increased from 5s to reduce CPU
)

// System info cache to avoid repeated gopsutil calls
var (
	sysInfoCache      *SystemInfo
	sysInfoCacheTime  time.Time
	sysInfoCacheMutex sync.RWMutex
)

// Host info cache (rarely changes)
var (
	hostInfoCache      *host.InfoStat
	hostInfoCacheTime  time.Time
	hostInfoCacheMutex sync.RWMutex
)

// Background CPU collection cache
var (
	cpuPercentCache      []float64
	cpuPercentCacheMutex sync.RWMutex
)

// getProcessList returns a list of processes sorted by CPU usage (with caching)
func getProcessList() ([]ProcessInfo, error) {
	processCacheMutex.RLock()
	if time.Since(processCacheTime) < processCacheTTL && processCache != nil {
		result := make([]ProcessInfo, len(processCache))
		copy(result, processCache)
		processCacheMutex.RUnlock()
		return result, nil
	}
	processCacheMutex.RUnlock()

	// Cache miss - fetch fresh data
	processCacheMutex.Lock()
	defer processCacheMutex.Unlock()

	// Double-check after acquiring write lock
	if time.Since(processCacheTime) < processCacheTTL && processCache != nil {
		result := make([]ProcessInfo, len(processCache))
		copy(result, processCache)
		return result, nil
	}

	processList, err := fetchProcessList()
	if err != nil {
		return nil, err
	}

	processCache = processList
	processCacheTime = time.Now()

	result := make([]ProcessInfo, len(processCache))
	copy(result, processCache)
	return result, nil
}

// fetchProcessList fetches the actual process list from the system
func fetchProcessList() ([]ProcessInfo, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}

	var processList []ProcessInfo
	for _, p := range procs {
		name, _ := p.Name()
		cpuPercent, _ := p.CPUPercent()
		memPercent, _ := p.MemoryPercent()
		status, _ := p.Status()
		username, _ := p.Username()

		// Convert status slice to string
		statusStr := "unknown"
		if len(status) > 0 {
			statusStr = status[0]
		}

		processList = append(processList, ProcessInfo{
			PID:        p.Pid,
			Name:       name,
			CPUPercent: cpuPercent,
			MemPercent: memPercent,
			Status:     statusStr,
			Username:   username,
		})
	}

	// Sort by CPU usage descending
	sort.Slice(processList, func(i, j int) bool {
		return processList[i].CPUPercent > processList[j].CPUPercent
	})

	return processList, nil
}

// getCachedHostInfo returns cached host info to avoid repeated syscalls
func getCachedHostInfo() (*host.InfoStat, error) {
	hostInfoCacheMutex.RLock()
	if hostInfoCache != nil && time.Since(hostInfoCacheTime) < hostInfoCacheTTL {
		info := hostInfoCache
		hostInfoCacheMutex.RUnlock()
		return info, nil
	}
	hostInfoCacheMutex.RUnlock()

	hostInfoCacheMutex.Lock()
	defer hostInfoCacheMutex.Unlock()

	// Double-check after acquiring write lock
	if hostInfoCache != nil && time.Since(hostInfoCacheTime) < hostInfoCacheTTL {
		return hostInfoCache, nil
	}

	info, err := host.Info()
	if err != nil {
		return nil, err
	}

	hostInfoCache = info
	hostInfoCacheTime = time.Now()
	return info, nil
}

// getCachedCPUPercent returns cached CPU percent from background collector
func getCachedCPUPercent() []float64 {
	cpuPercentCacheMutex.RLock()
	defer cpuPercentCacheMutex.RUnlock()

	if cpuPercentCache == nil {
		// Fallback if background collector hasn't run yet
		percent, _ := cpu.Percent(0, true)
		return percent
	}

	result := make([]float64, len(cpuPercentCache))
	copy(result, cpuPercentCache)
	return result
}

// startCPUCollector starts background CPU usage collection
func startCPUCollector() {
	// Initial collection
	percent, _ := cpu.Percent(time.Second, true)
	cpuPercentCacheMutex.Lock()
	cpuPercentCache = percent
	cpuPercentCacheMutex.Unlock()

	go func() {
		ticker := time.NewTicker(cpuCollectInterval)
		defer ticker.Stop()

		for range ticker.C {
			// Use blocking call for accurate measurement
			percent, err := cpu.Percent(time.Second, true)
			if err != nil {
				continue
			}

			cpuPercentCacheMutex.Lock()
			cpuPercentCache = percent
			cpuPercentCacheMutex.Unlock()
		}
	}()
}

func getSystemInfo() (*SystemInfo, error) {
	hostInfo, err := getCachedHostInfo()
	if err != nil {
		return nil, err
	}

	cpuInfo, err := cpu.Info()
	if err != nil {
		return nil, err
	}

	// Use cached CPU percent from background collector
	cpuPercent := getCachedCPUPercent()

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

	// Temperature sensors (optional - can be disabled for CPU savings)
	var temps []TempInfo
	if enableTemperature {
		sensors, _ := host.SensorsTemperatures()
		for _, s := range sensors {
			if s.Temperature > 0 {
				temps = append(temps, TempInfo{
					Name:        s.SensorKey,
					Temperature: s.Temperature,
				})
			}
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

// getCachedSystemInfo returns cached system info to reduce CPU usage
func getCachedSystemInfo() (*SystemInfo, error) {
	sysInfoCacheMutex.RLock()
	if sysInfoCache != nil && time.Since(sysInfoCacheTime) < sysInfoCacheTTL {
		info := sysInfoCache
		sysInfoCacheMutex.RUnlock()
		return info, nil
	}
	sysInfoCacheMutex.RUnlock()

	sysInfoCacheMutex.Lock()
	defer sysInfoCacheMutex.Unlock()

	// Double-check after acquiring write lock
	if sysInfoCache != nil && time.Since(sysInfoCacheTime) < sysInfoCacheTTL {
		return sysInfoCache, nil
	}

	info, err := getSystemInfo()
	if err != nil {
		return nil, err
	}

	sysInfoCache = info
	sysInfoCacheTime = time.Now()
	return info, nil
}

func handleSystemInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	info, err := getCachedSystemInfo()
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
// Query params:
//   - minutes: for recent data (default: 60, uses memory buffer for <=60 min)
//   - start: Unix timestamp for range start
//   - end: Unix timestamp for range end (default: now)
//   - format: "json" (default) or "csv"
func handleHistory(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	format := query.Get("format")
	if format == "" {
		format = "json"
	}

	var data []HistoryPoint
	var startTime, endTime int64
	var useDB bool

	// Check if time range is specified
	if startStr := query.Get("start"); startStr != "" {
		startTime, _ = strconv.ParseInt(startStr, 10, 64)
		endTime = time.Now().Unix()
		if endStr := query.Get("end"); endStr != "" {
			endTime, _ = strconv.ParseInt(endStr, 10, 64)
		}
		useDB = true
	} else {
		// Use minutes parameter (backward compatible)
		minutes := 60
		if m := query.Get("minutes"); m != "" {
			if v, err := strconv.Atoi(m); err == nil && v > 0 {
				minutes = v
			}
		}
		endTime = time.Now().Unix()
		startTime = time.Now().Add(-time.Duration(minutes) * time.Minute).Unix()

		// Use memory buffer for recent data (<=60 min), DB for longer periods
		if minutes <= 60 {
			data = historyBuffer.GetSince(startTime)
		} else {
			useDB = true
		}
	}

	// Query from database if needed
	if useDB {
		var err error
		data, err = queryHistoryFromDB(startTime, endTime)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
	}

	// Return CSV format
	if format == "csv" {
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=sysinfo_history_%d_%d.csv", startTime, endTime))

		writer := csv.NewWriter(w)
		// Write header
		writer.Write([]string{"timestamp", "datetime", "cpu_percent", "mem_percent", "disk_percent"})

		// Write data
		for _, p := range data {
			t := time.Unix(p.Timestamp, 0)
			writer.Write([]string{
				strconv.FormatInt(p.Timestamp, 10),
				t.Format("2006-01-02 15:04:05"),
				fmt.Sprintf("%.2f", p.CPUPercent),
				fmt.Sprintf("%.2f", p.MemPercent),
				fmt.Sprintf("%.2f", p.DiskPercent),
			})
		}
		writer.Flush()
		return
	}

	// Return JSON format (default)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"interval_seconds": int(historyInterval.Seconds()),
		"start_time":       startTime,
		"end_time":         endTime,
		"count":            len(data),
		"data":             data,
	})
}

// handleHistoryStats returns statistics about stored history data
func handleHistoryStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	minTime, maxTime, count, err := getHistoryStats()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// Calculate time range
	var durationHours float64
	var minTimeStr, maxTimeStr string
	if count > 0 {
		durationHours = float64(maxTime-minTime) / 3600.0
		minTimeStr = time.Unix(minTime, 0).Format("2006-01-02 15:04:05")
		maxTimeStr = time.Unix(maxTime, 0).Format("2006-01-02 15:04:05")
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_records":   count,
		"min_timestamp":   minTime,
		"max_timestamp":   maxTime,
		"min_datetime":    minTimeStr,
		"max_datetime":    maxTimeStr,
		"duration_hours":  durationHours,
		"interval_seconds": int(historyInterval.Seconds()),
	})
}

// handleMQTTConfig handles GET/POST for MQTT configuration
func handleMQTTConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		mqttMutex.RLock()
		config := MQTTConfig{
			Enabled:     mqttConfig.Enabled,
			Broker:      mqttConfig.Broker,
			Username:    mqttConfig.Username,
			Password:    "",
			TopicPrefix: mqttConfig.TopicPrefix,
			ClientID:    mqttConfig.ClientID,
		}
		// Mask password if set
		if mqttConfig.Password != "" {
			config.Password = "***"
		}
		mqttMutex.RUnlock()
		json.NewEncoder(w).Encode(config)

	case http.MethodPost:
		var newConfig MQTTConfig
		if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		mqttMutex.Lock()
		mqttConfig.Enabled = newConfig.Enabled
		mqttConfig.Broker = newConfig.Broker
		mqttConfig.Username = newConfig.Username
		// Only update password if not masked
		if newConfig.Password != "***" && newConfig.Password != "" {
			mqttConfig.Password = newConfig.Password
		} else if newConfig.Password == "" {
			mqttConfig.Password = ""
		}
		mqttConfig.TopicPrefix = newConfig.TopicPrefix
		mqttConfig.ClientID = newConfig.ClientID
		mqttMutex.Unlock()

		if err := saveMQTTConfig(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		// Reconnect with new settings
		go connectMQTT()

		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
	}
}

// handleMQTTStatus returns current MQTT connection status
func handleMQTTStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	mqttMutex.RLock()
	enabled := mqttConfig.Enabled
	broker := mqttConfig.Broker
	topicPrefix := mqttConfig.TopicPrefix
	connected := mqttConnected
	mqttMutex.RUnlock()

	status := "disabled"
	if enabled {
		if connected {
			status = "connected"
		} else {
			status = "disconnected"
		}
	}

	clientID := getEffectiveClientID()
	topic := fmt.Sprintf("%s/%s", topicPrefix, clientID)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"enabled":   enabled,
		"connected": connected,
		"status":    status,
		"broker":    broker,
		"topic":     topic,
	})
}

// handleProcessesAPI returns the list of processes as JSON
func handleProcessesAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse pagination parameters
	query := r.URL.Query()
	page := 1
	limit := 50

	if p := query.Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if l := query.Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 200 {
			limit = v
		}
	}

	processList, err := getProcessList()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	total := len(processList)
	totalPages := (total + limit - 1) / limit

	// Calculate offset
	offset := (page - 1) * limit
	end := offset + limit
	if offset > total {
		offset = total
	}
	if end > total {
		end = total
	}

	response := ProcessListResponse{
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
		Timestamp:  time.Now().Unix(),
		Processes:  processList[offset:end],
	}

	json.NewEncoder(w).Encode(response)
}

const processesPageHTML = `<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Process Monitor</title>
<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body { background: #0a0a0a; color: #00ff00; font-family: 'Courier New', monospace; font-size: 14px; padding: 20px; }
.container { max-width: 1000px; margin: 0 auto; }
h1 { color: #00ff00; border-bottom: 1px solid #333; padding-bottom: 10px; margin-bottom: 20px; font-size: 18px; }
.section { background: #111; border: 1px solid #333; margin-bottom: 15px; padding: 15px; border-radius: 4px; }
.section-title { color: #0af; font-weight: bold; margin-bottom: 10px; display: flex; justify-content: space-between; align-items: center; }
.stats { color: #666; font-size: 12px; }
table { width: 100%; border-collapse: collapse; }
th, td { padding: 8px; text-align: left; border-bottom: 1px solid #222; }
th { color: #0af; font-weight: bold; background: #1a1a1a; }
td { color: #0f0; }
tr:hover { background: #1a1a1a; }
.pid { color: #888; }
.name { color: #0f0; max-width: 200px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.cpu { color: #ff0; }
.mem { color: #f0f; }
.status { color: #0af; }
.user { color: #888; }
.high-cpu { color: #f00; }
.pagination { display: flex; justify-content: center; align-items: center; gap: 15px; margin-top: 15px; }
.page-btn { background: #222; color: #0af; border: 1px solid #444; padding: 8px 16px; border-radius: 3px; cursor: pointer; font-family: inherit; }
.page-btn:hover:not(:disabled) { background: #333; border-color: #0af; }
.page-btn:disabled { color: #444; cursor: not-allowed; }
.page-info { color: #666; }
.back-link { color: #0af; text-decoration: none; }
.back-link:hover { text-decoration: underline; }
.footer { display: flex; justify-content: space-between; align-items: center; margin-top: 15px; color: #444; font-size: 11px; }
</style>
</head>
<body>
<div class="container">
<h1>[ Process Monitor ]</h1>
<div class="section">
  <div class="section-title">
    <span>PROCESSES</span>
    <span class="stats">Total: <span id="total">-</span> processes</span>
  </div>
  <table>
    <thead>
      <tr>
        <th>PID</th>
        <th>NAME</th>
        <th>CPU%</th>
        <th>MEM%</th>
        <th>STATUS</th>
        <th>USER</th>
      </tr>
    </thead>
    <tbody id="process-list">
      <tr><td colspan="6" style="text-align:center;color:#666">Loading...</td></tr>
    </tbody>
  </table>
  <div class="pagination">
    <button class="page-btn" id="prev-btn" onclick="prevPage()">← Prev</button>
    <span class="page-info">Page <span id="current-page">1</span> / <span id="total-pages">1</span></span>
    <button class="page-btn" id="next-btn" onclick="nextPage()">Next →</button>
    <button class="page-btn" id="refresh-btn" onclick="loadProcesses()" style="margin-left:20px">↻ Refresh</button>
  </div>
</div>
<div class="footer">
  <a href="/" class="back-link">← Back to Dashboard</a>
  <span>Manual refresh | Cache: 5s</span>
</div>
</div>
<script>
let currentPage = 1;
let totalPages = 1;
const limit = 50;

function loadProcesses() {
  fetch('/api/processes?page=' + currentPage + '&limit=' + limit)
    .then(r => r.json())
    .then(data => {
      totalPages = data.total_pages;
      document.getElementById('total').textContent = data.total;
      document.getElementById('current-page').textContent = data.page;
      document.getElementById('total-pages').textContent = data.total_pages;
      document.getElementById('prev-btn').disabled = currentPage <= 1;
      document.getElementById('next-btn').disabled = currentPage >= totalPages;

      const tbody = document.getElementById('process-list');
      if (data.processes.length === 0) {
        tbody.innerHTML = '<tr><td colspan="6" style="text-align:center;color:#666">No processes</td></tr>';
        return;
      }

      tbody.innerHTML = data.processes.map(p => {
        const cpuClass = p.cpu_percent > 50 ? 'high-cpu' : 'cpu';
        return '<tr>' +
          '<td class="pid">' + p.pid + '</td>' +
          '<td class="name" title="' + p.name + '">' + p.name + '</td>' +
          '<td class="' + cpuClass + '">' + p.cpu_percent.toFixed(1) + '%</td>' +
          '<td class="mem">' + p.mem_percent.toFixed(1) + '%</td>' +
          '<td class="status">' + p.status + '</td>' +
          '<td class="user">' + (p.username || '-') + '</td>' +
          '</tr>';
      }).join('');
    })
    .catch(e => {
      document.getElementById('process-list').innerHTML = '<tr><td colspan="6" style="color:red">Error: ' + e + '</td></tr>';
    });
}

function prevPage() {
  if (currentPage > 1) {
    currentPage--;
    loadProcesses();
  }
}

function nextPage() {
  if (currentPage < totalPages) {
    currentPage++;
    loadProcesses();
  }
}

loadProcesses();
</script>
</body>
</html>`

// handleProcessesPage serves the processes HTML page
func handleProcessesPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(processesPageHTML))
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

		point := HistoryPoint{
			Timestamp:   time.Now().Unix(),
			CPUPercent:  cpuAvg,
			MemPercent:  info.Memory.UsedPercent,
			DiskPercent: info.Disk.UsedPercent,
		}

		// Save to memory buffer (for fast recent queries)
		historyBuffer.Push(point)

		// Save to database (for persistent long-term storage)
		if err := saveHistoryToDB(point); err != nil {
			log.Printf("Failed to save history to DB: %v\n", err)
		}

		// Publish to MQTT if enabled
		publishMetrics(point)
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
	// Initialize database for persistent history storage
	if err := initDB(); err != nil {
		log.Printf("Warning: Failed to initialize database: %v\n", err)
		log.Println("History will only be stored in memory (max 1 hour)")
	}

	// Load MQTT configuration and connect if enabled
	if err := loadMQTTConfig(); err != nil {
		log.Printf("Warning: Failed to load MQTT config: %v\n", err)
	} else {
		connectMQTT()
	}

	// Start background CPU collector for accurate measurements
	startCPUCollector()
	log.Printf("CPU collector started (interval: %v)\n", cpuCollectInterval)

	// Start history collector in background
	go collectHistory()

	http.HandleFunc("/", handleDashboard)
	http.HandleFunc("/api/system", handleSystemInfo)
	http.HandleFunc("/api/history", handleHistory)
	http.HandleFunc("/api/history/stats", handleHistoryStats)
	http.HandleFunc("/api/mqtt/config", handleMQTTConfig)
	http.HandleFunc("/api/mqtt/status", handleMQTTStatus)
	http.HandleFunc("/processes", handleProcessesPage)
	http.HandleFunc("/api/processes", handleProcessesAPI)
	http.HandleFunc("/health", handleHealth)
	log.Println("Server starting on :8088...")
	log.Printf("History: collecting every %v, memory buffer %d points, persistent storage enabled\n", historyInterval, historyMaxSize)
	http.ListenAndServe(":8088", nil)
}

func (p *program) Stop(s service.Service) error {
	// Disconnect MQTT
	disconnectMQTT()

	// Close database connection
	if db != nil {
		db.Close()
	}
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
