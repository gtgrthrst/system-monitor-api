# Dashboard Gauges & Sparklines Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add semi-circular gauges and sparklines for CPU, Memory, and Disk metrics to the system monitor dashboard.

**Architecture:** Pure SVG/CSS implementation embedded in the existing `dashboardHTML` constant in `main.go`. No external dependencies. Uses existing `/api/system` endpoint data and client-side JavaScript for updates.

**Tech Stack:** Go (HTML template string), SVG, CSS, vanilla JavaScript

---

## Task 1: Add CSS Styles for Gauges and Sparklines

**Files:**
- Modify: `/root/system-monitor/main.go:399-449` (inside `<style>` block)

**Step 1: Add gauge and sparkline CSS**

Add these styles before the closing `</style>` tag (after line 448, before line 449):

```css
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
```

**Step 2: Build and verify no syntax errors**

Run:
```bash
cd /root/system-monitor && go build -o sysinfo-api
```
Expected: Build succeeds with no errors

**Step 3: Commit**

```bash
git add main.go && git commit -m "style: add CSS for gauges and sparklines"
```

---

## Task 2: Add Gauges HTML Structure

**Files:**
- Modify: `/root/system-monitor/main.go:452-454` (inside `<body>`, after `<h1>`)

**Step 1: Add gauges container HTML**

Find this line (around line 454):
```html
<div id="content">Loading...</div>
```

Add this **before** that line:
```html
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
```

**Step 2: Build and verify**

Run:
```bash
cd /root/system-monitor && go build -o sysinfo-api
```
Expected: Build succeeds

**Step 3: Commit**

```bash
git add main.go && git commit -m "feat: add gauges and sparklines HTML structure"
```

---

## Task 3: Add JavaScript for Gauge Updates

**Files:**
- Modify: `/root/system-monitor/main.go:471-475` (inside `<script>`, near the top)

**Step 1: Add gauge helper functions and diskHistory**

Find these lines (around line 472-475):
```javascript
const MAX_POINTS = 60;
let cpuHistory = [];
let memHistory = [];
```

Replace with:
```javascript
const MAX_POINTS = 60;
const SPARK_POINTS = 20;
let cpuHistory = [];
let memHistory = [];
let diskHistory = [];

function updateGauge(id, percent) {
  const svg = document.getElementById(id);
  if (!svg) return;
  const arc = svg.querySelector('.gauge-value');
  const text = svg.querySelector('.gauge-percent');
  const arcLength = 157; // π * 50 (radius)
  const offset = arcLength - (percent / 100) * arcLength;
  arc.style.strokeDashoffset = offset;
  text.textContent = percent.toFixed(1) + '%';
}

function updateSparkline(id, data, maxPoints) {
  const svg = document.getElementById(id);
  if (!svg || data.length < 2) return;
  const w = 100, h = 25;
  const points = data.slice(-maxPoints).map((v, i, arr) => {
    const x = (i / (maxPoints - 1)) * w;
    const y = h - (v / 100) * (h - 2);
    return x + ',' + y;
  });
  const linePoints = points.join(' ');
  const areaPoints = '0,' + h + ' ' + linePoints + ' ' + w + ',' + h;
  svg.querySelector('.spark-line').setAttribute('points', linePoints);
  svg.querySelector('.spark-area').setAttribute('points', areaPoints);
  const lastData = data[data.length - 1];
  const lastX = w;
  const lastY = h - (lastData / 100) * (h - 2);
  const dot = svg.querySelector('.spark-dot');
  dot.setAttribute('cx', lastX);
  dot.setAttribute('cy', lastY);
}
```

**Step 2: Build and verify**

Run:
```bash
cd /root/system-monitor && go build -o sysinfo-api
```
Expected: Build succeeds

**Step 3: Commit**

```bash
git add main.go && git commit -m "feat: add gauge and sparkline update functions"
```

---

## Task 4: Integrate Gauge Updates in Main Loop

**Files:**
- Modify: `/root/system-monitor/main.go` (inside `update()` function)

**Step 1: Update the update() function**

Find the `update()` function. Inside, after updating cpuHistory and memHistory (around lines 540-543):
```javascript
cpuHistory.push(cpuAvg);
memHistory.push(d.memory.used_percent);
if (cpuHistory.length > MAX_POINTS) cpuHistory.shift();
if (memHistory.length > MAX_POINTS) memHistory.shift();
```

Add right after:
```javascript
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
```

**Step 2: Build and verify**

Run:
```bash
cd /root/system-monitor && go build -o sysinfo-api
```
Expected: Build succeeds

**Step 3: Commit**

```bash
git add main.go && git commit -m "feat: integrate gauge and sparkline updates in main loop"
```

---

## Task 5: Manual Testing

**Step 1: Start the server**

Run:
```bash
cd /root/system-monitor && ./sysinfo-api &
```

**Step 2: Test the dashboard**

Run:
```bash
curl -s http://localhost:8088/ | head -100
```
Expected: HTML output containing `gauges-container`, `cpu-gauge`, `mem-gauge`, `disk-gauge`

**Step 3: Test API still works**

Run:
```bash
curl -s http://localhost:8088/api/system | head -5
```
Expected: JSON response with cpu, memory, disk data

**Step 4: Stop the server**

Run:
```bash
pkill -f sysinfo-api || true
```

**Step 5: Final commit (if any cleanup needed)**

```bash
git status
```
If clean, proceed to next task.

---

## Task 6: Update Design Document Status

**Files:**
- Modify: `/root/system-monitor/docs/plans/2026-01-21-dashboard-gauges-sparklines-design.md`

**Step 1: Update status**

Change line 4 from:
```
**狀態：** 已核准
```
To:
```
**狀態：** 已實作
```

**Step 2: Commit**

```bash
git add docs/plans/2026-01-21-dashboard-gauges-sparklines-design.md && git commit -m "docs: mark gauges design as implemented"
```

---

## Summary

| Task | Description | Est. Time |
|------|-------------|-----------|
| 1 | Add CSS styles | 2 min |
| 2 | Add HTML structure | 3 min |
| 3 | Add JS helper functions | 3 min |
| 4 | Integrate in update loop | 2 min |
| 5 | Manual testing | 3 min |
| 6 | Update docs | 1 min |

**Total: ~14 minutes**
