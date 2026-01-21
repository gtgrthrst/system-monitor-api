# Dashboard Card Layout Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Refactor dashboard from section-based layout to card-based layout with host subtitle, metric cards, and collapsible MQTT.

**Architecture:** Modify embedded HTML/CSS/JS in `main.go`. Remove canvas charts, add card containers with flexbox/grid layout. Use color thresholds for CPU core cards.

**Tech Stack:** Go (embedded HTML template), CSS3 Flexbox/Grid, Vanilla JavaScript

---

### Task 1: Add Host Subtitle CSS

**Files:**
- Modify: `main.go:403` (after h1 CSS)

**Step 1: Add CSS for host subtitle**

Find line 403:
```css
h1 { color: #00ff00; border-bottom: 1px solid #333; padding-bottom: 10px; margin-bottom: 20px; font-size: 18px; }
```

Add after it:
```css
.host-subtitle { color: #888; font-size: 12px; text-align: center; margin: -15px 0 20px 0; }
```

**Step 2: Verify change**

Run: `grep -n "host-subtitle" main.go`
Expected: One line with `.host-subtitle` CSS

**Step 3: Commit**

```bash
git add main.go
git commit -m "style: add host subtitle CSS"
```

---

### Task 2: Add Metric Cards CSS

**Files:**
- Modify: `main.go:404` (after host-subtitle CSS)

**Step 1: Add metric cards container and card CSS**

Add after `.host-subtitle`:
```css
.metric-cards { display: flex; justify-content: center; flex-wrap: wrap; gap: 15px; margin-bottom: 15px; }
.metric-card { background: #111; border: 1px solid #333; border-radius: 4px; padding: 12px 15px; min-width: 140px; text-align: center; flex: 1; max-width: 180px; }
.metric-card-title { color: #fff; font-size: 12px; font-weight: bold; margin-bottom: 8px; }
.metric-card-bar { background: #222; height: 8px; border-radius: 4px; overflow: hidden; margin-bottom: 6px; }
.metric-card-bar-fill { height: 100%; border-radius: 4px; transition: width 0.3s; }
.metric-card-percent { font-size: 16px; font-weight: bold; margin-bottom: 4px; }
.metric-card-detail { color: #666; font-size: 11px; }
```

**Step 2: Verify change**

Run: `grep -n "metric-card" main.go | head -5`
Expected: Lines with `.metric-cards` and `.metric-card` CSS

**Step 3: Commit**

```bash
git add main.go
git commit -m "style: add metric cards CSS"
```

---

### Task 3: Add CPU Core Cards CSS

**Files:**
- Modify: `main.go` (after metric cards CSS)

**Step 1: Add core cards CSS**

Add after `.metric-card-detail`:
```css
.core-cards { display: flex; justify-content: center; flex-wrap: wrap; gap: 10px; margin-bottom: 15px; }
.core-card { background: #111; border: 1px solid #333; border-radius: 4px; padding: 8px 12px; min-width: 80px; text-align: center; }
.core-card-title { color: #888; font-size: 10px; margin-bottom: 6px; }
.core-card-bar { background: #222; height: 6px; border-radius: 3px; overflow: hidden; margin-bottom: 4px; }
.core-card-bar-fill { height: 100%; border-radius: 3px; transition: width 0.3s, background-color 0.3s; }
.core-card-percent { font-size: 12px; font-weight: bold; }
```

**Step 2: Verify change**

Run: `grep -n "core-card" main.go | head -5`
Expected: Lines with `.core-cards` and `.core-card` CSS

**Step 3: Commit**

```bash
git add main.go
git commit -m "style: add CPU core cards CSS"
```

---

### Task 4: Add Collapsible MQTT CSS

**Files:**
- Modify: `main.go` (after core cards CSS)

**Step 1: Add collapsible section CSS**

Add after `.core-card-percent`:
```css
.mqtt-collapsible { background: #111; border: 1px solid #333; border-radius: 4px; margin-bottom: 15px; }
.mqtt-header { display: flex; justify-content: space-between; align-items: center; padding: 12px 15px; cursor: pointer; user-select: none; }
.mqtt-header:hover { background: #1a1a1a; }
.mqtt-header-title { color: #888; font-size: 12px; display: flex; align-items: center; gap: 8px; }
.mqtt-header-title .arrow { transition: transform 0.2s; }
.mqtt-header-title .arrow.expanded { transform: rotate(90deg); }
.mqtt-body { display: none; padding: 0 15px 15px 15px; border-top: 1px solid #333; }
.mqtt-body.expanded { display: block; }
```

**Step 2: Verify change**

Run: `grep -n "mqtt-collapsible\|mqtt-header\|mqtt-body" main.go | head -5`
Expected: Lines with collapsible MQTT CSS

**Step 3: Commit**

```bash
git add main.go
git commit -m "style: add collapsible MQTT section CSS"
```

---

### Task 5: Remove Old CSS (chart, section styles)

**Files:**
- Modify: `main.go:404-427` area

**Step 1: Remove .section CSS**

Find and remove:
```css
.section { background: #111; border: 1px solid #333; margin-bottom: 15px; padding: 15px; border-radius: 4px; }
.section-title { color: #0af; font-weight: bold; margin-bottom: 10px; }
```

**Step 2: Remove .chart CSS**

Find and remove:
```css
.chart { background: #0a0a0a; border: 1px solid #333; border-radius: 3px; margin: 10px 0; }
.chart-label { display: flex; justify-content: space-between; font-size: 11px; color: #555; padding: 0 5px; }
```

**Step 3: Verify removal**

Run: `grep -n "\.section\|\.chart" main.go | grep -v "mqtt-section"`
Expected: No matches (only mqtt-section should remain as ID reference)

**Step 4: Commit**

```bash
git add main.go
git commit -m "style: remove old section and chart CSS"
```

---

### Task 6: Add Host Subtitle HTML

**Files:**
- Modify: `main.go:468` (after h1 tag)

**Step 1: Add subtitle div**

Find:
```html
<h1>[ System Monitor ]</h1>
```

Change to:
```html
<h1>[ System Monitor ]</h1>
<div id="host-subtitle" class="host-subtitle">Loading...</div>
```

**Step 2: Verify change**

Run: `grep -n "host-subtitle" main.go`
Expected: CSS line and HTML line with host-subtitle

**Step 3: Commit**

```bash
git add main.go
git commit -m "feat: add host subtitle HTML element"
```

---

### Task 7: Add Cards Container HTML

**Files:**
- Modify: `main.go:510` (content div area)

**Step 1: Replace content div with cards containers**

Find:
```html
<div id="content">Loading...</div>
```

Change to:
```html
<div id="metric-cards" class="metric-cards">Loading...</div>
<div id="core-cards" class="core-cards"></div>
```

**Step 2: Verify change**

Run: `grep -n "metric-cards\|core-cards" main.go`
Expected: HTML lines with metric-cards and core-cards divs

**Step 3: Commit**

```bash
git add main.go
git commit -m "feat: add metric and core cards container HTML"
```

---

### Task 8: Refactor MQTT Section to Collapsible

**Files:**
- Modify: `main.go:511-524` (mqtt-section area)

**Step 1: Replace MQTT section HTML**

Find the mqtt-section div (lines 511-524) and replace with:
```html
<div id="mqtt-section" class="mqtt-collapsible">
  <div class="mqtt-header" onclick="toggleMQTT()">
    <span class="mqtt-header-title"><span class="arrow" id="mqtt-arrow">▶</span> MQTT Settings</span>
    <span id="mqtt-status-indicator" class="mqtt-status"><span class="dot disabled"></span><span id="mqtt-status-text">Disabled</span></span>
  </div>
  <div class="mqtt-body" id="mqtt-body">
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
</div>
```

**Step 2: Verify structure**

Run: `grep -n "mqtt-collapsible\|mqtt-header\|mqtt-body" main.go`
Expected: HTML structure with collapsible elements

**Step 3: Commit**

```bash
git add main.go
git commit -m "feat: refactor MQTT section to collapsible"
```

---

### Task 9: Add JavaScript Helper Functions

**Files:**
- Modify: `main.go:566` (after formatUptime function)

**Step 1: Add getColorByPercent function**

Find line after `formatUptime` function (around line 577), add before `drawChart`:
```javascript
function getColorByPercent(p) {
  if (p <= 50) return '#0f0';
  if (p <= 80) return '#ff0';
  return '#f44';
}
```

**Step 2: Add toggleMQTT function**

Add after `getColorByPercent`:
```javascript
function toggleMQTT() {
  const body = document.getElementById('mqtt-body');
  const arrow = document.getElementById('mqtt-arrow');
  const expanded = body.classList.toggle('expanded');
  arrow.classList.toggle('expanded', expanded);
  arrow.textContent = expanded ? '▼' : '▶';
  localStorage.setItem('mqttExpanded', expanded);
}
function initMQTTState() {
  if (localStorage.getItem('mqttExpanded') === 'true') {
    document.getElementById('mqtt-body').classList.add('expanded');
    document.getElementById('mqtt-arrow').classList.add('expanded');
    document.getElementById('mqtt-arrow').textContent = '▼';
  }
}
```

**Step 3: Verify functions added**

Run: `grep -n "getColorByPercent\|toggleMQTT" main.go`
Expected: Both function definitions found

**Step 4: Commit**

```bash
git add main.go
git commit -m "feat: add color threshold and MQTT toggle functions"
```

---

### Task 10: Remove drawChart Function

**Files:**
- Modify: `main.go:578-623` (drawChart function)

**Step 1: Remove entire drawChart function**

Find and remove the entire `drawChart` function (lines 578-623):
```javascript
function drawChart(canvasId, data, color, fillColor) {
  // ... entire function body ...
  ctx.stroke();
}
```

**Step 2: Verify removal**

Run: `grep -n "drawChart" main.go`
Expected: No matches

**Step 3: Commit**

```bash
git add main.go
git commit -m "refactor: remove drawChart function"
```

---

### Task 11: Rewrite update() Function - Render Cards

**Files:**
- Modify: `main.go:625-691` (update function)

**Step 1: Replace update function content**

Replace the entire content rendering part of `update()` (from line 647 to 687) with:
```javascript
function update() {
  fetch('/api/system').then(r => r.json()).then(d => {
    let cpuAvg = d.cpu.usage_percent.reduce((a,b) => a+b, 0) / d.cpu.usage_percent.length;

    // Update history
    cpuHistory.push(cpuAvg);
    memHistory.push(d.memory.used_percent);
    if (cpuHistory.length > MAX_POINTS) cpuHistory.shift();
    if (memHistory.length > MAX_POINTS) memHistory.shift();
    diskHistory.push(d.disk.used_percent);
    if (diskHistory.length > MAX_POINTS) diskHistory.shift();

    // Update gauges
    updateGauge('cpu-gauge', cpuAvg);
    updateGauge('mem-gauge', d.memory.used_percent);
    updateGauge('disk-gauge', d.disk.used_percent);

    // Update sparklines
    updateSparkline('cpu-spark', cpuHistory, SPARK_POINTS);
    updateSparkline('mem-spark', memHistory, SPARK_POINTS);
    updateSparkline('disk-spark', diskHistory, SPARK_POINTS);

    // Update host subtitle
    document.getElementById('host-subtitle').textContent =
      d.host.hostname + ' • ' + d.host.platform + ' • Uptime: ' + formatUptime(d.host.uptime_seconds);

    // Render metric cards
    let metricCards =
      '<div class="metric-card">' +
        '<div class="metric-card-title">CPU</div>' +
        '<div class="metric-card-bar"><div class="metric-card-bar-fill" style="width:' + cpuAvg + '%;background:#0f0"></div></div>' +
        '<div class="metric-card-percent" style="color:#0f0">' + cpuAvg.toFixed(1) + '%</div>' +
        '<div class="metric-card-detail">' + d.cpu.model_name.split(' ').slice(0,3).join(' ') + '</div>' +
      '</div>' +
      '<div class="metric-card">' +
        '<div class="metric-card-title">MEMORY</div>' +
        '<div class="metric-card-bar"><div class="metric-card-bar-fill" style="width:' + d.memory.used_percent + '%;background:#f0f"></div></div>' +
        '<div class="metric-card-percent" style="color:#f0f">' + d.memory.used_percent.toFixed(1) + '%</div>' +
        '<div class="metric-card-detail">' + formatBytes(d.memory.used_bytes) + ' / ' + formatBytes(d.memory.total_bytes) + '</div>' +
      '</div>' +
      '<div class="metric-card">' +
        '<div class="metric-card-title">DISK</div>' +
        '<div class="metric-card-bar"><div class="metric-card-bar-fill" style="width:' + d.disk.used_percent + '%;background:#ff0"></div></div>' +
        '<div class="metric-card-percent" style="color:#ff0">' + d.disk.used_percent.toFixed(1) + '%</div>' +
        '<div class="metric-card-detail">' + formatBytes(d.disk.used_bytes) + ' / ' + formatBytes(d.disk.total_bytes) + '</div>' +
      '</div>';

    // Temperature card (conditional)
    if (d.temperature && d.temperature.length > 0) {
      let maxTemp = Math.max(...d.temperature.map(t => t.temperature));
      let tempColor = maxTemp < 60 ? '#0f0' : (maxTemp < 80 ? '#ff0' : '#f44');
      metricCards +=
        '<div class="metric-card">' +
          '<div class="metric-card-title">TEMP</div>' +
          '<div class="metric-card-bar"><div class="metric-card-bar-fill" style="width:' + Math.min(maxTemp, 100) + '%;background:' + tempColor + '"></div></div>' +
          '<div class="metric-card-percent" style="color:' + tempColor + '">' + maxTemp.toFixed(1) + '°C</div>' +
          '<div class="metric-card-detail">' + d.temperature[0].name + '</div>' +
        '</div>';
    }
    document.getElementById('metric-cards').innerHTML = metricCards;

    // Render core cards
    let coreCards = d.cpu.usage_percent.map((p, i) => {
      let color = getColorByPercent(p);
      return '<div class="core-card">' +
        '<div class="core-card-title">Core ' + i + '</div>' +
        '<div class="core-card-bar"><div class="core-card-bar-fill" style="width:' + p + '%;background:' + color + '"></div></div>' +
        '<div class="core-card-percent" style="color:' + color + '">' + p.toFixed(1) + '%</div>' +
      '</div>';
    }).join('');
    document.getElementById('core-cards').innerHTML = coreCards;

  }).catch(e => { document.getElementById('metric-cards').innerHTML = '<div style="color:red">Error: ' + e + '</div>'; });
}
```

**Step 2: Verify update function**

Run: `grep -n "metric-cards\|core-cards\|host-subtitle" main.go | grep "getElementById"`
Expected: Lines updating metric-cards, core-cards, and host-subtitle

**Step 3: Commit**

```bash
git add main.go
git commit -m "feat: rewrite update function for card-based layout"
```

---

### Task 12: Add initMQTTState Call

**Files:**
- Modify: `main.go:768` (after loadMQTTConfig)

**Step 1: Add init call**

Find:
```javascript
loadMQTTConfig();
loadMQTTStatus();
```

Change to:
```javascript
loadMQTTConfig();
loadMQTTStatus();
initMQTTState();
```

**Step 2: Verify**

Run: `grep -n "initMQTTState" main.go`
Expected: Function definition and call

**Step 3: Commit**

```bash
git add main.go
git commit -m "feat: initialize MQTT collapsed state from localStorage"
```

---

### Task 13: Update RWD Media Query

**Files:**
- Modify: `main.go:463` (media query area)

**Step 1: Update media query for cards**

Find:
```css
@media (max-width: 600px) { .gauges-container { flex-direction: column; align-items: center; } .gauge-card { width: 100%; max-width: 250px; } }
```

Change to:
```css
@media (max-width: 768px) { .gauges-container { flex-direction: column; align-items: center; } .gauge-card { width: 100%; max-width: 250px; } .metric-cards { flex-direction: row; } .metric-card { min-width: 45%; } .core-cards { justify-content: center; } .core-card { min-width: 60px; } }
```

**Step 2: Verify**

Run: `grep -n "@media" main.go`
Expected: Updated media query with metric-cards and core-cards

**Step 3: Commit**

```bash
git add main.go
git commit -m "style: update responsive design for card layout"
```

---

### Task 14: Build and Test

**Files:**
- Build: `sysinfo-api` binary

**Step 1: Build the project**

Run: `go build -o sysinfo-api`
Expected: No errors, binary created

**Step 2: Run the server**

Run: `./sysinfo-api &`
Expected: Server starts on :8088

**Step 3: Test health endpoint**

Run: `curl -s http://localhost:8088/health`
Expected: `{"status":"ok"}`

**Step 4: Test dashboard elements**

Run: `curl -s http://localhost:8088/ | grep -o "metric-card\|core-card\|host-subtitle\|mqtt-collapsible" | sort -u`
Expected: All four element types found

**Step 5: Stop test server**

Run: `pkill -f sysinfo-api`

**Step 6: Commit build artifact**

```bash
git add sysinfo-api
git commit -m "build: update binary with card layout"
```

---

### Task 15: Final Integration Test

**Step 1: Start server and verify visually**

Run: `./sysinfo-api &`

**Step 2: Verify API response**

Run: `curl -s http://localhost:8088/api/system | head -c 200`
Expected: JSON with host, cpu, memory, disk fields

**Step 3: Verify HTML structure**

Run: `curl -s http://localhost:8088/ | grep -E "(host-subtitle|metric-cards|core-cards|mqtt-collapsible)" | wc -l`
Expected: At least 4 lines

**Step 4: Cleanup**

Run: `pkill -f sysinfo-api`

**Step 5: Final commit if needed**

If any fixes were made:
```bash
git add -A
git commit -m "fix: final adjustments for card layout"
```

---

## Summary

Total tasks: 15
Estimated commits: 13-15

Key changes:
1. CSS: Add card styles, remove section/chart styles
2. HTML: Add host subtitle, card containers, collapsible MQTT
3. JS: Add helper functions, rewrite update(), remove drawChart()
4. RWD: Update breakpoints for card layout
