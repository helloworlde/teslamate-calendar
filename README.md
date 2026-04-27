# teslamate-calendar

[Module](https://github.com/helloworlde/teslamate-calendar)：`github.com/helloworlde/teslamate-calendar`

`teslamate-calendar` 通过 `teslamateapi` HTTP API 读取数据并输出 **RFC 5545** iCalendar（`text/calendar`）格式。无数据库、无后台任务、无 Redis，轻量级部署。

**镜像：**

- `ghcr.io/helloworlde/teslamate-calendar:latest`（`main` 构建）
- 其他标签见 [Packages](https://github.com/helloworlde/teslamate-calendar/pkgs/container/teslamate-calendar)

## 功能特性

- 📅 生成符合 RFC 5545 标准的 iCalendar 格式
- 🚗 支持行程、充电、软件更新、日报/周报/月报
- 🔒 内置缓存机制，支持防缓存击穿（singleflight）
- 🌍 支持时区配置和多种视图模式
- 🔗 集成 Google Maps 和 Grafana 看板链接

## 环境变量

| 变量 | 说明 |
|------|------|
| `LISTEN_ADDR` | 监听地址，默认 `:8080` |
| `TESLAMATE_API_BASE_URL` | **必填**，teslamateapi 根 URL（仅协议+主机+端口，内部使用 `/api/v1`） |
| `TESLAMATE_API_TOKEN` | 上游 Bearer token（可空） |
| `TESLAMATE_API_AUTH_HEADER` / `TESLAMATE_API_AUTH_SCHEME` | 上游鉴权，默认 `Authorization` + `Bearer` |
| `CALENDAR_FEED_TOKEN` | **必填**，无默认值；与订阅 URL 路径中的 Token 必须一致。请使用长随机串，**不要使用示例字面量作为生产值** |
| `DEFAULT_DAYS` / `MAX_DAYS` | 回溯与上限天数 |
| `DEFAULT_TIMEZONE` | IANA 时区，用于日界与事件展示 |
| `DEFAULT_VIEW` | `compact` / `normal` / `detail` |
| `CACHE_TTL_SECONDS` | 日历响应缓存 |
| `REQUEST_TIMEOUT_SECONDS` | 请求上游超时 |
| `LOG_LEVEL` | 日志级别 |
| `TESLAMATE_DASHBOARD_URL_TEMPLATE` | 可选。Grafana 看板 URL 模板。占位符：`{from}` `{to}`（Unix 毫秒）、`{car_id}`、`{range}`、`{event_id}`。空则不附加「TeslaMate 看板」链接。 |

不读取 TeslaMate `globalsettings` 中的 Grafana 地址，避免隐式行为。

## HTTP 接口

- `GET /healthz` `GET /readyz` `GET /ping`
- `GET /cars` `GET /cars/{CarID}`（不校验 `CALENDAR_FEED_TOKEN`；上游仍由 teslamateapi 保护）
- `GET /calendar/token/{Token}/cars/{CarID}/daily.ics`（`range=day\|week\|month`）
- `GET .../drives.ics` `.../charges.ics` `.../updates.ics` `.../all.ics`
- `GET /openapi.json` `GET /swagger/index.html` `GET /scalar`

路径中的 `Token` 须与 `CALENDAR_FEED_TOKEN` 一致。OpenAPI/Scalar 中 token 仅展示占位符 `change-me-to-random-token`，不会预填本机真实 secret。

**不建议**默认在日历客户端中订阅 `all.ics`（体量大、更新频繁）；可按需使用单日/单类日历。

## 富文本

- 主字段为纯文本 `DESCRIPTION`（推荐所有客户端以可读性为准）。
- 若未来在事件中填写 `HTMLDescription`，会多输出一行 `X-ALT-DESC;FMTTYPE=text/html`；**不保证** Apple / Google 等一定按 HTML 显示，多数环境仍按纯文本处理。
- 当前业务逻辑**不填充** `HTMLDescription`。

## 地图链接

- 有坐标时使用 HTTPS Google Maps（路线或点位）；无坐标时 HTTPS 搜索。不使用 `maps://`、Amap 等深链。
- 各平台由浏览器或系统策略打开；**不保证**直接唤起「系统地图 App」。
- 历史轨迹的完整展开应在 TeslaMate/Grafana 看板中查看，不会塞进短 URL。

## Docker

```bash
docker run -d \
  --name teslamate-calendar \
  -p 8088:8080 \
  -e TESLAMATE_API_BASE_URL="http://teslamateapi:8080" \
  -e CALENDAR_FEED_TOKEN="your-random-token" \
  ghcr.io/helloworlde/teslamate-calendar:latest
```

`docker compose` 示例见 [docker-compose.yml](docker-compose.yml)。

本地构建：

```bash
docker build -t teslamate-calendar:local .
```

## 开发

### 本地运行

```bash
export TESLAMATE_API_BASE_URL=http://localhost:4000
export CALENDAR_FEED_TOKEN=dev-only-token
go run ./cmd/teslamate-calendar
```

### 运行测试

```bash
go test ./...
```

### 代码结构

- `cmd/teslamate-calendar/` - 主程序入口
- `internal/api/` - HTTP API 处理器和路由
- `internal/calendar/` - iCalendar 生成逻辑
- `internal/client/` - TeslaMate API 客户端
- `internal/config/` - 配置管理
- `internal/model/` - 数据模型
- `internal/service/` - 业务服务层
- `internal/util/` - 工具函数

## License

MIT
