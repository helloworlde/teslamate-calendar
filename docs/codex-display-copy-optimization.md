# Codex 任务：优化 ICS 标题与 DESCRIPTION 可读性

目标：重构 teslamate-calendar 的事件标题、日报/周报/月报、行程、充电、更新文案，使其适合日历客户端阅读。要求逐项执行、写入检查文件、执行测试并确认。

## 1. 执行要求

必须在仓库根目录新增或更新：

- `docs/implementation-checklist.md`

把本任务拆成 checklist，逐项执行并勾选。每完成一个阶段都要更新该文件。最后必须执行：

```bash
go test ./...
```

最后在 checklist 中写入：

```md
## Final Verification
- [ ] go test ./... passed
- [ ] daily.ics sample checked
- [ ] drives.ics sample checked
- [ ] charges.ics sample checked
- [ ] all.ics sample checked
```

如果某项无法执行，必须在 checklist 中写明原因。

## 2. 当前问题

当前 DESCRIPTION 类似：

```text
日期：2026-04-11
车辆：Tesla

━━━━━━━━━━━━
📊 当日摘要
━━━━━━━━━━━━
行程：7 次
里程：135.56 km
行驶：225 min
最高速度：109 km/h
充电：1 次
充入：33.92 kWh
充电时长：288 min
费用：39.17
电量：43% → 66%

━━━━━━━━━━━━
🚗 行程明细
━━━━━━━━━━━━
- 10:32-10:42 蓟门里东区 → 苏宁易购, 海淀区 1.85km 10min 171Wh/km 43%→43%
- 11:20-11:34 苏宁易购, 海淀区 → 蓟门里东区 3.20km 14min 192Wh/km 42%→41%

━━━━━━━━━━━━
⚡ 充电明细
━━━━━━━━━━━━
- 21:01-01:49 +33.92kWh 11%→66% @蓟门里东区
地图：https://example.com/maps
```

问题：

1. 摘要字段堆叠，没有重点。
2. `225 min`、`288 min` 不适合中文阅读。
3. 行程明细一行塞入路线、距离、耗时、电耗、电量，可读性差。
4. 地点中英文标点混用，例如 `苏宁易购, 海淀区`。
5. 地图和看板链接层级不清晰。
6. 标题、摘要、明细缺少统一格式。

## 3. 总体原则

1. 标题短，详情清晰。
2. DESCRIPTION 先概览，后明细，最后链接。
3. 摘要区最多展示关键指标，不做流水账。
4. 行程/充电明细使用多行结构。
5. 没有数据的模块不显示。
6. 值为 0 且无业务意义时不显示。
7. 不显示内部 ID：CarID、行程ID、充电ID、更新ID。
8. 不显示 `null`、`<nil>`、`NaN`、`Inf`。
9. 所有用户可见时间使用 query timezone。
10. daily summary 不展示每条地图链接；地图链接主要放在 drives.ics / charges.ics 单事件里。

## 4. SUMMARY 标题规范

### daily.ics range=day

compact：

```text
📊 136km · 7🚗 · 1⚡
```

normal：

```text
📊 Model 3 · 136km · 7行程 · 1充电
```

detail：

```text
📊 Model 3 日报 · 135.6km · 7行程 · 1充电 · +33.9kWh
```

只有行程：

```text
📊 Model 3 · 20km · 2行程
```

只有充电：

```text
📊 Model 3 · 1充电 · +33.9kWh
```

### daily.ics range=week

```text
📊 Model 3 周报 · 286km · 18行程 · 2充电
```

### daily.ics range=month

```text
📊 Model 3 月报 · 1261km · 62行程 · 8充电
```

### drives.ics

```text
🚗 Model 3 · 天兴五街 → 蓟门里东区 · 64.5km
```

### charges.ics

```text
⚡ Model 3 · +33.9kWh · 11%→66%
```

## 5. daily DESCRIPTION 新格式

### 有行程 + 充电

```text
2026-04-11 · Model 3

━━━━━━━━━━━━
📌 今日概览
━━━━━━━━━━━━
🚗 行程 7 次 · 135.6 km · 3小时45分钟
⚡ 充电 1 次 · +33.9 kWh · 4小时48分钟
🔋 电量 43% → 66%
🏁 最高速度 109 km/h
💰 费用 39.17

━━━━━━━━━━━━
🚗 行程
━━━━━━━━━━━━
1. 10:32-10:42 · 10分钟 · 1.9 km
   蓟门里东区 → 苏宁易购（海淀区）
   171 Wh/km · 43%→43%

2. 11:20-11:34 · 14分钟 · 3.2 km
   苏宁易购（海淀区） → 蓟门里东区
   192 Wh/km · 42%→41%

━━━━━━━━━━━━
⚡ 充电
━━━━━━━━━━━━
1. 21:01-次日 01:49 · 4小时48分钟
   蓟门里东区
   +33.9 kWh · 11%→66% · 39.17

━━━━━━━━━━━━
🔗 相关链接
━━━━━━━━━━━━
TeslaMate 看板：
https://example.com
```

### 只有行程

```text
2026-04-24 · Model 3

━━━━━━━━━━━━
📌 今日概览
━━━━━━━━━━━━
🚗 行程 2 次 · 15.9 km · 40分钟
🏁 最高速度 80 km/h
🔋 电量 83% → 80%

━━━━━━━━━━━━
🚗 行程
━━━━━━━━━━━━
1. 09:33-10:00 · 26分钟 · 9.8 km
   蓟门里东区 → 城奥大厦附近
   127 Wh/km · 83%→81%
```

### 只有充电

```text
2026-04-11 · Model 3

━━━━━━━━━━━━
📌 今日概览
━━━━━━━━━━━━
⚡ 充电 1 次 · +33.9 kWh · 4小时48分钟
🔋 电量 11% → 66%
💰 费用 39.17

━━━━━━━━━━━━
⚡ 充电
━━━━━━━━━━━━
1. 21:01-次日 01:49 · 4小时48分钟
   蓟门里东区
   +33.9 kWh · 11%→66% · 39.17
```

## 6. drives.ics DESCRIPTION 格式

```text
Model 3 · 行程

━━━━━━━━━━━━
🚗 路线
━━━━━━━━━━━━
蓟门里东区 → 城奥大厦附近

━━━━━━━━━━━━
📊 数据
━━━━━━━━━━━━
时间：09:33-10:00
耗时：26分钟
距离：9.8 km
电耗：127 Wh/km
电量：83%→81%
最高速度：80 km/h

━━━━━━━━━━━━
📍 位置
━━━━━━━━━━━━
起点：蓟门里东区
终点：城奥大厦附近

━━━━━━━━━━━━
🔗 相关链接
━━━━━━━━━━━━
地图：
https://example.com/maps

TeslaMate 看板：
https://example.com/dashboard
```

要求：

- 不显示行程ID。
- 有轨迹时显示：`轨迹：已记录，可在 TeslaMate 看板查看`。
- 不输出 polyline、coordinates、geojson 原文。
- 没有地图链接时不显示地图字段。
- 没有看板链接时不显示 TeslaMate 看板字段。

## 7. charges.ics DESCRIPTION 格式

```text
Model 3 · 充电

━━━━━━━━━━━━
⚡ 充电
━━━━━━━━━━━━
地点：蓟门里东区
时间：21:01-次日 01:49
耗时：4小时48分钟

━━━━━━━━━━━━
📊 数据
━━━━━━━━━━━━
充入：33.9 kWh
电量：11%→66%
费用：39.17
最高功率：11 kW

━━━━━━━━━━━━
🔗 相关链接
━━━━━━━━━━━━
地图：
https://example.com/maps

TeslaMate 看板：
https://example.com/dashboard
```

要求：

- 不显示充电ID。
- 没有费用不显示费用。
- 没有功率不显示功率。
- 没有地图链接不显示地图字段。
- 没有看板链接不显示 TeslaMate 看板字段。

## 8. week/month DESCRIPTION 格式

周报：

```text
2026-04-13 ~ 2026-04-19 · Model 3

━━━━━━━━━━━━
📌 本周概览
━━━━━━━━━━━━
🚗 行程 18 次 · 286.3 km · 9小时42分钟
⚡ 充电 2 次 · +82.4 kWh
🏁 最高速度 109 km/h
💰 费用 99.58

━━━━━━━━━━━━
📅 每日概览
━━━━━━━━━━━━
04-13 · 19.7 km · 2行程
04-14 · 20.3 km · 3行程
04-19 · 13.1 km · 5行程 · 1充电
```

月报：

```text
2026-04 · Model 3

━━━━━━━━━━━━
📌 本月概览
━━━━━━━━━━━━
🚗 行程 62 次 · 1260.5 km · 38小时20分钟
⚡ 充电 8 次 · +318.2 kWh
🏁 最高速度 129 km/h
💰 费用 386.20

━━━━━━━━━━━━
📅 每周概览
━━━━━━━━━━━━
第 1 周 · 156.2 km · 9行程 · 1充电
第 2 周 · 286.3 km · 18行程 · 2充电
第 3 周 · 341.5 km · 20行程 · 3充电
```

## 9. 格式化规则

时间：

- `40分钟`
- `3小时45分钟`
- `21:01-次日 01:49`

不要输出：

- `225 min`
- `288 min`
- `1 h 22 min`
- `10min`

距离：

- 标题：`136km`
- 详情：`135.6 km`

电量：

- `43%→66%`
- 摘要：`🔋 电量 43% → 66%`

电耗：

- `171 Wh/km`

位置名称规范化：

- `苏宁易购, 海淀区` => `苏宁易购（海淀区）`
- `天兴五街, 廊坊市` => `天兴五街（廊坊市）`

## 10. 代码实现建议

建议新增或重构：

- `internal/calendar/copy.go`
- `internal/calendar/format.go`
- `internal/calendar/descriptions.go`

建议实现：

```go
func FormatDurationZH(seconds float64) string
func FormatDistanceTitle(km float64) string
func FormatDistanceDetail(km float64) string
func FormatKWhTitle(kwh float64) string
func FormatKWhDetail(kwh float64) string
func FormatBatteryRange(start, end *float64, spaced bool) string
func FormatTimeRange(start, end time.Time, loc *time.Location) string
func NormalizePlaceName(name string) string
func Section(title string, lines ...string) []string
func LinksSection(mapURL, dashboardURL string) []string
```

## 11. 测试要求

必须新增或更新测试：

1. `FormatDurationZH(225*60)` => `3小时45分钟`
2. `FormatDurationZH(40*60)` => `40分钟`
3. 跨天时间范围：`21:01-次日 01:49`
4. `NormalizePlaceName("苏宁易购, 海淀区")` => `苏宁易购（海淀区）`
5. daily 有行程+充电的 DESCRIPTION 格式正确
6. daily 只有行程时不出现 `⚡ 充电`
7. daily 只有充电时不出现 `🚗 行程`
8. 不出现 `行程ID`、`充电ID`、`更新ID`
9. 不出现 `225 min`、`10min`、`33.92kWh`
10. drives DESCRIPTION 包含 `🚗 路线`、`📊 数据`、`📍 位置`、`🔗 相关链接`
11. charges DESCRIPTION 包含 `⚡ 充电`、`📊 数据`
12. 有轨迹时显示 `轨迹：已记录，可在 TeslaMate 看板查看`
13. 不输出 polyline、coordinates、geojson 原文

## 12. 验收标准

执行：

```bash
go test ./...
```

必须通过。

人工检查生成的 daily.ics，不允许再出现：

```text
行驶：225 min
- 10:32-10:42 蓟门里东区 → 苏宁易购, 海淀区 1.85km 10min 171Wh/km 43%→43%
- 21:01-01:49 +33.92kWh 11%→66% @蓟门里东区
```

必须改成类似：

```text
2026-04-11 · Model 3

━━━━━━━━━━━━
📌 今日概览
━━━━━━━━━━━━
🚗 行程 7 次 · 135.6 km · 3小时45分钟
⚡ 充电 1 次 · +33.9 kWh · 4小时48分钟
🔋 电量 43% → 66%
🏁 最高速度 109 km/h
💰 费用 39.17

━━━━━━━━━━━━
🚗 行程
━━━━━━━━━━━━
1. 10:32-10:42 · 10分钟 · 1.9 km
   蓟门里东区 → 苏宁易购（海淀区）
   171 Wh/km · 43%→43%
```

最终 `docs/implementation-checklist.md` 必须全部勾选，并明确写明测试和样例检查结果。
