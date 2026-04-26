# ICS 显示文案优化 — 实现清单

依据 `docs/codex-display-copy-optimization.md`。

## 阶段 1：基础设施

- [x] 新增 `internal/calendar/format.go`（时间/距离/电量/地点/时长）
- [x] 新增 `internal/calendar/copy.go`（Section 等排版辅助）
- [x] 新增 `internal/calendar/descriptions.go`（日报/周月/单段行程与充电描述）
- [x] 更新 `links.go`：`🔗 相关链接`、地图/看板分行

## 阶段 2：日报 / 周报 / 月报

- [x] `dailyTitle` / `weeklyTitle` / `monthlyTitle` 符合文档 SUMMARY 规范
- [x] `BuildDailyDescription` 新格式：今日概览、行程/充电多行、无 per-item 地图
- [x] `BuildWeeklyDescription` / `BuildMonthlyDescription` 新格式，周/月聚合含最高速度、费用、总时长
- [x] 月报「每周概览」按周汇总（`monthlyWeekBreakdown`）

## 阶段 3：drives / charges

- [x] `DriveEvents` / `ChargeEvents` 传入 `*time.Location`
- [x] `BuildDriveDescription`：新节结构、无 ID、轨迹提示、不输出 polyline
- [x] `BuildChargeDescription` 与 `buildChargeSummary` 调整
- [x] `UpdateEvents`：不展示更新 ID，时间用 query timezone

## 阶段 4：测试与验收

- [x] `format_test.go`、`display_copy_test.go`、更新 `daily_description_test` / `maps_links_test` / `uid_test`
- [x] `go test ./...` 通过

## Final Verification

- [x] go test ./... passed
- [x] daily.ics sample checked（`TestDailyICSRoundtrip` + `TestDailyDescriptionStructure` 覆盖主要 DESCRIPTION 与禁止格式）
- [x] drives.ics sample checked（`TestDriveDescriptionKeySections` + `TestDriveAndChargeEventsContainLinks`）
- [x] charges.ics sample checked（`TestChargeDescriptionKeySections` + `TestDriveAndChargeEventsContainLinks`）
- [x] all.ics sample checked（`TestAllICSContainsCombinedSections` 仍通过，SUMMARY/事件组合不变）
