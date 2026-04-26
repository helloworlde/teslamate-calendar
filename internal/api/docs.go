package api

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

const openAPIPlaceholderToken = "change-me-to-random-token"

func (h *Handlers) OpenAPIJSON(c *gin.Context) {
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.String(http.StatusOK, BuildOpenAPISpec())
}

func (h *Handlers) SwaggerIndex(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, swaggerIndexHTML())
}

func (h *Handlers) Scalar(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, `<!doctype html>
<html><head><meta charset="utf-8"><title>teslamate-calendar Scalar</title></head>
<body>
<script id="api-reference" data-url="/openapi.json"></script>
<script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
</body></html>`)
}

func swaggerIndexHTML() string {
	return `<!doctype html>
<html><head><meta charset="utf-8"><title>teslamate-calendar Swagger</title>
<link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css"></head>
<body><div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
<script>window.ui=SwaggerUIBundle({url:'/openapi.json',dom_id:'#swagger-ui'})</script>
</body></html>`
}

func BuildOpenAPISpec() string {
	spec := map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":       "teslamate-calendar API",
			"version":     "1.0.0",
			"description": "通过 teslamateapi 读取数据并输出 RFC 5545 iCalendar。环境变量 TESLAMATE_API_BASE_URL 仅需协议与主机（内部固定 /api/v1）。Grafana 链接由 TESLAMATE_DASHBOARD_URL_TEMPLATE 配置。日历 URL 路径中的 Token 必须与 CALENDAR_FEED_TOKEN 一致。OpenAPI 中的 token 仅为占位符，不会使用运行中真实值。",
		},
		"tags": []any{
			map[string]any{"name": "Health", "description": "存活与上游就绪探针"},
			map[string]any{"name": "Cars", "description": "车辆列表与详情（teslamateapi 代理）"},
			map[string]any{"name": "Calendar", "description": "ICS 订阅源"},
			map[string]any{"name": "Docs", "description": "OpenAPI 与文档 UI"},
		},
		"components": map[string]any{
			"schemas":    openAPIComponentSchemas(),
			"parameters": openAPICommonParameters(),
			"responses":  openAPICommonResponses(),
		},
		"paths": buildPaths(),
	}
	b, _ := json.Marshal(spec)
	return string(b)
}

func openAPIComponentSchemas() map[string]any {
	return map[string]any{
		"Error": map[string]any{
			"type":     "object",
			"required": []string{"error"},
			"properties": map[string]any{
				"error": map[string]any{"type": "string", "description": "错误信息"},
			},
		},
		"HealthStatus": map[string]any{
			"type":     "object",
			"required": []string{"status"},
			"properties": map[string]any{
				"status": map[string]any{"type": "string", "example": "ok"},
			},
		},
		"ReadyStatus": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"status": map[string]any{"type": "string", "example": "ready"},
				"error":  map[string]any{"type": "string"},
			},
		},
		"Car": map[string]any{
			"type":        "object",
			"description": "与 teslamateapi 返回结构一致，字段以实际为准。",
		},
		"ICalendar": map[string]any{
			"type":        "string",
			"description": "text/calendar 正文，RFC 5545。",
		},
	}
}

func openAPICommonParameters() map[string]any {
	return map[string]any{
		"CarID": map[string]any{
			"name": "CarID", "in": "path", "required": true,
			"schema": map[string]any{
				"type":        "string",
				"example":     "1",
				"description": "车辆 ID（数字字符串）",
			},
		},
		"Token": map[string]any{
			"name": "Token", "in": "path", "required": true,
			"schema": map[string]any{
				"type":        "string",
				"example":     openAPIPlaceholderToken,
				"description": "与部署环境变量 CALENDAR_FEED_TOKEN 一致；示例为占位符，非真实值。",
			},
		},
		"days": map[string]any{
			"name": "days", "in": "query",
			"schema": map[string]any{
				"type":    "integer",
				"minimum": 1,
				"example": 90,
			},
		},
		"startDate": map[string]any{
			"name": "startDate", "in": "query",
			"schema": map[string]any{"type": "string", "description": "区间起点（与 endDate 联用，UTC 或含时区）"},
		},
		"endDate": map[string]any{
			"name": "endDate", "in": "query",
			"schema": map[string]any{"type": "string", "description": "区间终点"},
		},
		"timezone": map[string]any{
			"name": "timezone", "in": "query",
			"schema": map[string]any{
				"type":    "string",
				"example": "Asia/Shanghai",
			},
		},
		"range": map[string]any{
			"name": "range", "in": "query",
			"schema": map[string]any{
				"type":    "string",
				"enum":    []string{"day", "week", "month"},
				"default": "day",
			},
		},
		"view": map[string]any{
			"name": "view", "in": "query",
			"schema": map[string]any{
				"type":    "string",
				"enum":    []string{"compact", "normal", "detail"},
				"default": "normal",
			},
		},
		"detail": map[string]any{
			"name": "detail", "in": "query",
			"schema": map[string]any{
				"type":    "boolean",
				"default": true,
			},
		},
		"vehicleName": map[string]any{
			"name": "vehicleName", "in": "query",
			"schema": map[string]any{"type": "string", "description": "覆盖日历中显示的车辆名称"},
		},
		"minDistance": map[string]any{
			"name": "minDistance", "in": "query",
			"schema": map[string]any{"type": "string", "description": "转发至 teslamateapi 的行程筛选"},
		},
		"maxDistance": map[string]any{
			"name": "maxDistance", "in": "query",
			"schema": map[string]any{"type": "string", "description": "转发至 teslamateapi 的行程筛选"},
		},
	}
}

func openAPICommonResponses() map[string]any {
	ref := func(s string) map[string]any {
		return map[string]any{"$ref": "#/components/schemas/" + s}
	}
	return map[string]any{
		"BadRequest": map[string]any{
			"description": "参数错误",
			"content": map[string]any{
				"application/json": map[string]any{"schema": ref("Error")},
			},
		},
		"GatewayTimeout": map[string]any{
			"description": "上游请求超时",
			"content": map[string]any{
				"application/json": map[string]any{"schema": ref("Error")},
			},
		},
		"BadGateway": map[string]any{
			"description": "teslamateapi 不可用",
			"content": map[string]any{
				"application/json": map[string]any{"schema": ref("Error")},
			},
		},
		"Unauthorized": map[string]any{
			"description": "Token 与 CALENDAR_FEED_TOKEN 不一致",
			"content": map[string]any{
				"application/json": map[string]any{"schema": ref("Error")},
			},
		},
	}
}

func refParam(name string) map[string]any {
	return map[string]any{"$ref": "#/components/parameters/" + name}
}

func buildPaths() map[string]any {
	upErr := map[string]any{"$ref": "#/components/responses/BadGateway"}
	jsonSchema := func(s string) map[string]any {
		return map[string]any{"schema": map[string]any{"$ref": "#/components/schemas/" + s}}
	}
	paths := map[string]any{
		"/healthz": map[string]any{
			"get": map[string]any{
				"tags":    []string{"Health"},
				"summary": "Liveness",
				"responses": map[string]any{
					"200": map[string]any{
						"description": "服务存活",
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{"$ref": "#/components/schemas/HealthStatus"},
							},
						},
					},
				},
			},
		},
		"/readyz": map[string]any{
			"get": map[string]any{
				"tags":    []string{"Health"},
				"summary": "Readiness",
				"responses": map[string]any{
					"200": map[string]any{
						"content": map[string]any{"application/json": jsonSchema("ReadyStatus")},
					},
					"502": map[string]any{
						"content": map[string]any{"application/json": jsonSchema("Error")},
					},
				},
			},
		},
		"/ping": map[string]any{
			"get": map[string]any{
				"tags":    []string{"Health"},
				"summary": "ping",
				"responses": map[string]any{
					"200": map[string]any{
						"content": map[string]any{
							"text/plain": map[string]any{
								"schema": map[string]any{"type": "string", "example": "pong"},
							},
						},
					},
				},
			},
		},
		"/openapi.json": map[string]any{
			"get": map[string]any{
				"tags":      []string{"Docs"},
				"summary":   "OpenAPI JSON",
				"responses": map[string]any{"200": map[string]any{"description": "本规范"}},
			},
		},
		"/cars": map[string]any{
			"get": map[string]any{
				"tags":    []string{"Cars"},
				"summary": "车辆列表",
				"responses": map[string]any{
					"200": map[string]any{
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{
									"type":  "array",
									"items": map[string]any{"$ref": "#/components/schemas/Car"},
								},
							},
						},
					},
					"502": upErr,
				},
			},
		},
		"/cars/{CarID}": map[string]any{
			"get": map[string]any{
				"tags":       []string{"Cars"},
				"summary":    "单车详情",
				"parameters": []any{refParam("CarID")},
				"responses": map[string]any{
					"200": map[string]any{"content": map[string]any{"application/json": jsonSchema("Car")}},
					"502": upErr,
				},
			},
		},
		"/calendar/token/{Token}/cars/{CarID}/all.ics":     icsGetOp("all.ics", true, true, true, true, true),
		"/calendar/token/{Token}/cars/{CarID}/daily.ics":   icsGetOp("daily.ics 汇总（range=day|week|month）", true, true, true, true, true),
		"/calendar/token/{Token}/cars/{CarID}/drives.ics":  icsGetOp("行程", true, true, true, true, true),
		"/calendar/token/{Token}/cars/{CarID}/charges.ics": icsGetOp("充电", true, true, true, true, true),
		"/calendar/token/{Token}/cars/{CarID}/updates.ics": icsGetOp("软件更新", true, true, true, true, true),
	}
	return paths
}

func icsGetOp(desc string, days, se, tz, rng, vdet bool) map[string]any {
	params := []any{refParam("Token"), refParam("CarID")}
	if days {
		params = append(params, refParam("days"))
	}
	if se {
		params = append(params, refParam("startDate"), refParam("endDate"))
	}
	if tz {
		params = append(params, refParam("timezone"))
	}
	if rng {
		params = append(params, refParam("range"))
	}
	params = append(params, refParam("view"))
	if vdet {
		params = append(params, refParam("detail"), refParam("vehicleName"), refParam("minDistance"), refParam("maxDistance"))
	}
	return map[string]any{
		"get": map[string]any{
			"tags":        []string{"Calendar"},
			"summary":     "ICS: " + desc,
			"description": "路径 Token 与 CALENDAR_FEED_TOKEN 必须一致。",
			"parameters":  params,
			"responses": map[string]any{
				"200": map[string]any{
					"description": "iCalendar",
					"content": map[string]any{
						"text/calendar": map[string]any{
							"schema": map[string]any{"$ref": "#/components/schemas/ICalendar"},
						},
					},
				},
				"400": map[string]any{"$ref": "#/components/responses/BadRequest"},
				"401": map[string]any{"$ref": "#/components/responses/Unauthorized"},
				"504": map[string]any{"$ref": "#/components/responses/GatewayTimeout"},
				"502": map[string]any{"$ref": "#/components/responses/BadGateway"},
			},
		},
	}
}
