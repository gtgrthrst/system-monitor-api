# Dashboard Card Layout Design

Date: 2026-01-21

## Overview

將儀表板從 section 區塊佈局改為卡片式佈局，移除大型趨勢圖，保留頂部 gauges + sparklines。

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                  [ System Monitor ]                      │
│       racknerd-34a1e8f • Ubuntu • Uptime: 31d 12h       │
├─────────────────────────────────────────────────────────┤
│  ┌─────────┐      ┌─────────┐      ┌─────────┐          │
│  │ ◔ CPU   │      │ ◔ MEM   │      │ ◔ DISK  │          │
│  │  ~~~    │      │  ~~~    │      │  ~~~    │          │
│  └─────────┘      └─────────┘      └─────────┘          │
│              (保留現有 Gauges + Sparklines)              │
├─────────────────────────────────────────────────────────┤
│  ┌────────────┐ ┌────────────┐ ┌────────────┐ ┌───────┐ │
│  │    CPU     │ │   MEMORY   │ │    DISK    │ │ TEMP  │ │
│  │ ████░░ 45% │ │ ███░░░ 41% │ │ ██░░░░ 35% │ │(opt)  │ │
│  │ 1.8/4 GHz  │ │ 850M/2.0G  │ │ 11G/35G    │ │ 45°C  │ │
│  └────────────┘ └────────────┘ └────────────┘ └───────┘ │
├─────────────────────────────────────────────────────────┤
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐ ...                │
│  │ C0   │ │ C1   │ │ C2   │ │ C3   │   (CPU 核心卡)     │
│  │██ 15%│ │█  8% │ │███85%│ │█  5% │   色彩依使用率     │
│  └──────┘ └──────┘ └──────┘ └──────┘                    │
├─────────────────────────────────────────────────────────┤
│  ▶ MQTT Settings (點擊展開)                              │
└─────────────────────────────────────────────────────────┘
```

## Components

### 1. Host Subtitle

標題下方顯示系統資訊：

```
[ System Monitor ]
racknerd-34a1e8f • Ubuntu • Uptime: 31d 12h
```

- Hostname、Platform、Uptime 以 `•` 分隔
- 置中對齊，顏色較淡 (`#888`)

### 2. Metric Cards (CPU / Memory / Disk)

```
┌──────────────┐
│     CPU      │  ← 標題（白色）
│ ████████░░░░ │  ← 進度條
│     45.2%    │  ← 百分比（亮色）
│  1.8 / 4 GHz │  ← 詳細數值（灰色）
└──────────────┘
```

**CSS:**
```css
.metric-card {
  background: #111;
  border: 1px solid #333;
  border-radius: 4px;
  padding: 12px;
  min-width: 140px;
  text-align: center;
}
```

**內容：**
- CPU: 平均使用率 + 型號簡稱
- Memory: 已用/總量 (bytes formatted)
- Disk: 已用/總量 (bytes formatted)

### 3. CPU Core Cards

```
┌────────┐
│ Core 0 │
│ ████░░ │
│  15.2% │
└────────┘
```

**色彩閾值：**
- 0-50%: 綠色 `#0f0`
- 51-80%: 黃色 `#ff0`
- 81-100%: 紅色 `#f44`

### 4. Temperature Card (Conditional)

- 僅當 `temperature !== null` 時渲染
- 顯示最高溫度感測器數值
- 色彩依溫度：<60° 綠、60-80° 黃、>80° 紅

### 5. MQTT Collapsible Section

**收合狀態（預設）：**
```
┌─────────────────────────────────────────────────────────┐
│ ▶ MQTT Settings                          [Disabled] ○  │
└─────────────────────────────────────────────────────────┘
```

**展開狀態：**
```
┌─────────────────────────────────────────────────────────┐
│ ▼ MQTT Settings                          [Connected] ● │
├─────────────────────────────────────────────────────────┤
│  Broker    [ tcp://broker.example.com:1883 ]           │
│  Client ID [ sysinfo-monitor                 ]          │
│  Topic     [ metrics/system                  ]          │
│  Username  [ admin                           ]          │
│  Password  [ ••••••••                        ]          │
│                                                         │
│            [ Save & Connect ]  [ Disconnect ]           │
└─────────────────────────────────────────────────────────┘
```

**互動：**
- 點擊標題列切換展開/收合
- 狀態指示器：○ Disabled（灰）/ ● Connected（綠）/ ● Error（紅）
- 展開狀態保存至 localStorage

## Responsive Design

**桌面版（>768px）：**
```
┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐
│  CPU   │ │ MEMORY │ │  DISK  │ │  TEMP  │
└────────┘ └────────┘ └────────┘ └────────┘
┌────┐ ┌────┐ ┌────┐ ┌────┐ ┌────┐ ┌────┐
│ C0 │ │ C1 │ │ C2 │ │ C3 │ │ C4 │ │ C5 │
└────┘ └────┘ └────┘ └────┘ └────┘ └────┘
```

**手機版（≤768px）：**
```
┌────────┐ ┌────────┐
│  CPU   │ │ MEMORY │
└────────┘ └────────┘
┌────────┐ ┌────────┐
│  DISK  │ │  TEMP  │
└────────┘ └────────┘
┌────┐ ┌────┐ ┌────┐
│ C0 │ │ C1 │ │ C2 │
└────┘ └────┘ └────┘
```

## Removal List

1. **Canvas 趨勢圖**
   - `<canvas id="cpuChart">` 和 `<canvas id="memChart">`
   - `drawChart()` 函數
   - `.chart`, `.chart-label` CSS

2. **舊版 Section 佈局**
   - `.section` 區塊結構
   - 原有的 `.bar-container` 進度條

## Implementation Notes

- 保留現有 Gauges + Sparklines（頂部）
- 保留 sparkline 用的 history arrays
- 新增 `getColorByPercent(percent)` 函數處理色彩閾值
- MQTT 展開狀態使用 `localStorage.setItem('mqttExpanded', bool)`
