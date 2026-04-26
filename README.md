# teslamate-calendar

`teslamate-calendar` 是一个独立 Go 服务，只通过 `teslamateapi` HTTP API 读取数据并输出 RFC5545 兼容 ICS 订阅源。

## 架构

Calendar Client -> teslamate-calendar -> teslamateapi -> TeslaMate data

## 环境变量

- `LISTEN_ADDR` 默认 `:8080`
- `TESLAMATE_API_BASE_URL` 必填，**只需协议 + 主机（含端口）**，例如 `http://teslamateapi:8080`；**不要**写路径（如 `/api`），服务会固定请求 `.../api/v1/...`
- `TESLAMATE_API_TOKEN` 可选
- `TESLAMATE_API_AUTH_HEADER` 默认 `Authorization`
- `TESLAMATE_API_AUTH_SCHEME` 默认 `Bearer`
- `CALENDAR_FEED_TOKEN` 默认 `tesla`，可通过环境变量覆盖
- `DEFAULT_DAYS` 默认 `90`
- `MAX_DAYS` 默认 `365`
- `DEFAULT_TIMEZONE` 默认 `Asia/Shanghai`
- `DEFAULT_VIEW` 默认 `normal`，支持 `compact|normal|detail`
- `CACHE_TTL_SECONDS` 默认 `1800`
- `REQUEST_TIMEOUT_SECONDS` 默认 `10`
- `LOG_LEVEL` 默认 `info`
- `MAP_PROVIDER` 默认 `google`
- `TESLAMATE_DASHBOARD_BASE_URL` 可选；**未设置时**从 `GET /api/v1/globalsettings` 返回的 `data.settings.teslamate_urls.grafana_url` 读取作为 Grafana 看板根地址
- `TESLAMATE_DRIVE_DASHBOARD_PATH` 等 path 项可选，与上述根地址拼接

## Docker 部署

```bash
docker build -t teslamate-calendar:latest .
docker run --rm -p 8080:8080 \
  -e TESLAMATE_API_BASE_URL="http://teslamateapi:8080" \
  -e CALENDAR_FEED_TOKEN="tesla" \
  teslamate-calendar:latest
```

## docker-compose 部署

```bash
docker compose up -d
```

## 推荐订阅方式

- 日报（按天）：`http://localhost:8080/calendar/token/{token}/cars/1/daily.ics?range=day&view=normal&timezone=Asia/Shanghai`
- 周报：`http://localhost:8080/calendar/token/{token}/cars/1/daily.ics?range=week&view=normal&timezone=Asia/Shanghai`
- 月报：`http://localhost:8080/calendar/token/{token}/cars/1/daily.ics?range=month&view=normal&timezone=Asia/Shanghai`
- 行程：`http://localhost:8080/calendar/token/{token}/cars/1/drives.ics?view=normal&timezone=Asia/Shanghai`
- 充电：`http://localhost:8080/calendar/token/{token}/cars/1/charges.ics?view=normal&timezone=Asia/Shanghai`
- 更新：`http://localhost:8080/calendar/token/{token}/cars/1/updates.ics?view=normal&timezone=Asia/Shanghai`
- 全部：`http://localhost:8080/calendar/token/{token}/cars/1/all.ics?range=day&view=normal&timezone=Asia/Shanghai`

`all.ics` 仅作为高级订阅源，不推荐默认订阅（月视图会更杂乱）。

## 车辆名称

服务会自动从 `GET /api/v1/cars/:CarID` 推断显示名称，也支持 query 覆盖：

- `/calendar/token/{token}/cars/1/daily.ics?vehicleName=Model%203`

## 地图链接

- 行程：优先生成 Google 路线链接（起点->终点）
- 充电：优先生成 Google 单点位置链接
- 坐标缺失时回退到地点搜索链接

## TeslaMate 看板链接

Grafana 根地址：优先 `TESLAMATE_DASHBOARD_BASE_URL`，否则使用 teslamateapi 全局设置中的 `grafana_url`。  
支持用各 `TESLAMATE_*_DASHBOARD_PATH` 与根地址拼接，并自动附加 `from`/`to`/`var-car_id` 参数。

## 文档地址

- Swagger: `http://localhost:8080/swagger/index.html`
- Scalar: `http://localhost:8080/scalar`
- OpenAPI JSON: `http://localhost:8080/openapi.json`
