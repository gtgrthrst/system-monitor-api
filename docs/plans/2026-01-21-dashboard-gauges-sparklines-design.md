# 儀表板 UI 改進設計：圓形儀表盤與迷你走勢圖

**日期：** 2026-01-21
**狀態：** 已核准

## 概述

為 System Monitor 儀表板新增視覺元素，提升資料呈現的直觀性：
1. **圓形儀表盤 (Gauges)** - CPU、記憶體、磁碟使用率
2. **迷你走勢圖 (Sparklines)** - 各指標的歷史趨勢

## 技術選型

**方案：純 CSS/SVG 實作**
- 無外部函式庫依賴
- 保持單一檔案架構（HTML/CSS/JS 內嵌於 main.go）
- 與現有終端機視覺風格一致
- 載入速度快

## 設計細節

### 1. 圓形儀表盤

**視覺設計：**
- 半圓形儀表盤（180 度弧形）
- 三個主要儀表：CPU、記憶體、磁碟
- 尺寸：150x90 像素，並排顯示

**配色：**
| 指標 | 漸層色 |
|------|--------|
| CPU | `#0a0` → `#0f0`（綠色）|
| 記憶體 | `#a0a` → `#f0f`（紫色）|
| 磁碟 | `#aa0` → `#ff0`（黃色）|
| 背景軌道 | `#222`（深灰）|

**結構示意：**
```
      ╭─────────╮
     ╱           ╲      ← 背景軌道（灰色）
    ╱  ████████   ╲     ← 數值弧線（彩色）
   │      45%      │    ← 中間顯示百分比
   └───────────────┘
        CPU             ← 底部標籤
```

**動畫：**
- CSS transition 0.5s，數值變化時平滑過渡

### 2. 迷你走勢圖 (Sparklines)

**放置位置：**
- 每個儀表盤下方
- 尺寸：100x30 像素

**顯示資料：**
- 最近 20 個數據點（約 100 秒歷史）
- 僅顯示趨勢線，無軸線或標籤

**視覺設計：**
- 線條顏色與對應儀表盤相同
- 線寬 1.5px
- 半透明填充區域
- 最新數據點以小圓點標示

### 3. 整體版面配置

```
┌─────────────────────────────────────────┐
│  [ System Monitor ]                      │
├─────────────────────────────────────────┤
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  │
│  │  CPU    │  │ MEMORY  │  │  DISK   │  │
│  │  (儀表) │  │  (儀表) │  │  (儀表) │  │  ← 新增
│  │ ▁▂▃▅▆▇ │  │ ▁▂▃▅▆▇ │  │ ▁▂▃▅▆▇ │  │  ← 新增
│  └─────────┘  └─────────┘  └─────────┘  │
├─────────────────────────────────────────┤
│  HOST（主機資訊）                         │
├─────────────────────────────────────────┤
│  CPU DETAILS（CPU 詳細圖表 + 各核心）     │
├─────────────────────────────────────────┤
│  MEMORY（記憶體詳細資訊）                 │
├─────────────────────────────────────────┤
│  DISK（磁碟詳細資訊）                     │
├─────────────────────────────────────────┤
│  TEMPERATURE（溫度，如有）               │
├─────────────────────────────────────────┤
│  MQTT SETTINGS                          │
└─────────────────────────────────────────┘
```

**響應式設計：**
- 寬螢幕（>600px）：三個儀表盤並排
- 窄螢幕（≤600px）：儀表盤垂直堆疊

## 技術實作

### SVG 儀表盤結構

```html
<svg viewBox="0 0 150 90">
  <!-- 漸層定義 -->
  <defs>
    <linearGradient id="cpuGradient">
      <stop offset="0%" stop-color="#0a0"/>
      <stop offset="100%" stop-color="#0f0"/>
    </linearGradient>
  </defs>

  <!-- 背景軌道 -->
  <path d="M 15 75 A 60 60 0 0 1 135 75"
        stroke="#222" stroke-width="12" fill="none"
        stroke-linecap="round"/>

  <!-- 數值弧線 -->
  <path class="gauge-value"
        stroke="url(#cpuGradient)" stroke-width="12" fill="none"
        stroke-linecap="round"/>

  <!-- 中間數值 -->
  <text x="75" y="60" text-anchor="middle" class="gauge-percent">45%</text>

  <!-- 標籤 -->
  <text x="75" y="85" text-anchor="middle" class="gauge-label">CPU</text>
</svg>
```

### JavaScript 更新邏輯

```javascript
// 新增磁碟歷史陣列
let diskHistory = [];
const SPARKLINE_POINTS = 20;

// 更新函數中加入
function update() {
  fetch('/api/system').then(r => r.json()).then(d => {
    // ... 現有邏輯 ...

    // 更新磁碟歷史
    diskHistory.push(d.disk.used_percent);
    if (diskHistory.length > SPARKLINE_POINTS) diskHistory.shift();

    // 更新儀表盤
    updateGauge('cpu-gauge', cpuAvg);
    updateGauge('mem-gauge', d.memory.used_percent);
    updateGauge('disk-gauge', d.disk.used_percent);

    // 更新走勢圖
    drawSparkline('cpu-spark', cpuHistory.slice(-SPARKLINE_POINTS), '#0f0');
    drawSparkline('mem-spark', memHistory.slice(-SPARKLINE_POINTS), '#f0f');
    drawSparkline('disk-spark', diskHistory, '#ff0');
  });
}

// 計算弧線路徑
function calcArcPath(percent) {
  const angle = (percent / 100) * Math.PI;
  const x = 75 - 60 * Math.cos(angle);
  const y = 75 - 60 * Math.sin(angle);
  const largeArc = percent > 50 ? 1 : 0;
  return `M 15 75 A 60 60 0 ${largeArc} 1 ${x} ${y}`;
}

// 更新儀表盤
function updateGauge(id, percent) {
  const gauge = document.getElementById(id);
  gauge.querySelector('.gauge-value').setAttribute('d', calcArcPath(percent));
  gauge.querySelector('.gauge-percent').textContent = percent.toFixed(1) + '%';
}

// 繪製走勢圖
function drawSparkline(id, data, color) {
  const svg = document.getElementById(id);
  if (data.length < 2) return;

  const w = 100, h = 30;
  const points = data.map((v, i) => {
    const x = (i / (SPARKLINE_POINTS - 1)) * w;
    const y = h - (v / 100) * h;
    return `${x},${y}`;
  }).join(' ');

  svg.querySelector('.spark-line').setAttribute('points', points);
  svg.querySelector('.spark-area').setAttribute('points', `0,${h} ${points} ${w},${h}`);

  // 更新最新點位置
  const lastX = w;
  const lastY = h - (data[data.length - 1] / 100) * h;
  svg.querySelector('.spark-dot').setAttribute('cx', lastX);
  svg.querySelector('.spark-dot').setAttribute('cy', lastY);
}
```

### CSS 樣式

```css
.gauges-container {
  display: flex;
  justify-content: space-around;
  flex-wrap: wrap;
  gap: 15px;
  margin-bottom: 15px;
}

.gauge-card {
  background: #111;
  border: 1px solid #333;
  border-radius: 4px;
  padding: 15px;
  text-align: center;
  min-width: 160px;
}

.gauge-percent {
  font-size: 20px;
  font-weight: bold;
  fill: #fff;
}

.gauge-label {
  font-size: 12px;
  fill: #888;
}

.gauge-value {
  transition: d 0.5s ease-out;
}

.sparkline {
  margin-top: 10px;
}

.spark-line {
  fill: none;
  stroke-width: 1.5;
}

.spark-area {
  opacity: 0.2;
}

.spark-dot {
  r: 3;
}

/* 響應式設計 */
@media (max-width: 600px) {
  .gauges-container {
    flex-direction: column;
    align-items: center;
  }
}
```

## 檔案變更

| 檔案 | 變更內容 |
|------|----------|
| `main.go` | 修改 `dashboardHTML` 常數，新增約 150 行 HTML/CSS/JS |

## 保留的元素

- 原有 CPU、Memory 大型歷史圖表
- 各核心使用率網格
- 主機資訊、磁碟詳情、溫度區塊
- MQTT 設定區
- Process 頁面連結
