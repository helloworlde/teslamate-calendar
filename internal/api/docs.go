package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"teslamate-calendar/internal/config"
)

func (h *Handlers) OpenAPIJSON(c *gin.Context) {
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.String(http.StatusOK, BuildOpenAPISpec(h.cfg))
}

func (h *Handlers) SwaggerDocJSON(c *gin.Context) {
	h.OpenAPIJSON(c)
}

func (h *Handlers) SwaggerIndex(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, swaggerIndexHTML(h.cfg))
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

func openAPIExampleFeedToken(cfg config.Config) string {
	s := strings.TrimSpace(cfg.CalendarFeedToken)
	if s == "" {
		return "tesla"
	}
	return s
}

func swaggerIndexHTML(cfg config.Config) string {
	full := "Bearer " + openAPIExampleFeedToken(cfg)
	js, _ := json.Marshal(full)
	return fmt.Sprintf(`<!doctype html>
<html><head><meta charset="utf-8"><title>teslamate-calendar Swagger</title>
<link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css"></head>
<body><div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
<script>window.ui=SwaggerUIBundle({url:'/openapi.json',dom_id:'#swagger-ui',persistAuthorization:true,onComplete:function(){var v=%s;try{if(window.ui&&window.ui.preauthorizeApiKey)window.ui.preauthorizeApiKey('BearerAuth',v)}catch(e){}}})</script>
</body></html>`, string(js))
}

func BuildOpenAPISpec(cfg config.Config) string {
	feed := openAPIExampleFeedToken(cfg)
	spec := map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":   "teslamate-calendar API",
			"version": "1.0.0",
			"description": "通过 teslamateapi 读取数据并输出 RFC 5545 iCalendar 订阅。" +
				"环境变量 TESLAMATE_API_BASE_URL 仅需协议与主机（端口），内部固定访问 /api/v1；" +
				"Grafana 看板根地址优先 TESLAMATE_DASHBOARD_BASE_URL，否则使用 globalsettings 中的 grafana_url。" +
				"日历 URL 使用路径 token 与配置项 CALENDAR_FEED_TOKEN 一致时通过校验；/cars 是否要求 Bearer 由部署时的 REQUIRE_TOKEN_FOR_CARS 决定。" +
				" 下方参数 default/example 来自当前服务配置，便于在 Swagger/Scalar 中直接调试。",
		},
		"tags": []any{
			map[string]any{"name": "Health", "description": "存活与上游就绪探针"},
			map[string]any{"name": "Cars", "description": "车辆列表与详情（teslamateapi 代理）"},
			map[string]any{"name": "Calendar", "description": "ICS 订阅源"},
			map[string]any{"name": "Docs", "description": "OpenAPI 与文档 UI"},
		},
		"components": map[string]any{
			"securitySchemes": map[string]any{
				"BearerAuth": map[string]any{
					"type":        "apiKey",
					"in":          "header",
					"name":        "Authorization",
					"description": fmt.Sprintf("与 CALENDAR_FEED_TOKEN 相同；与日历 URL 中 Token 一致。完整头示例：Bearer %s。Swagger 页打开后会尝试预填该值。", feed),
				},
			},
			"schemas":    openAPIComponentSchemas(),
			"parameters": openAPICommonParameters(cfg),
			"examples":   openAPIExamples(),
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
			"format":      "binary",
			"description": "text/calendar 正文，RFC 5545。",
		},
	}
}

func openAPICommonParameters(cfg config.Config) map[string]any {
	feed := openAPIExampleFeedToken(cfg)
	carEx := "1"
	dayEx := cfg.DefaultDays
	if dayEx < 1 {
		dayEx = 90
	}
	tzEx := strings.TrimSpace(cfg.DefaultTimezone)
	if tzEx == "" {
		tzEx = "Asia/Shanghai"
	}
	viewDef := strings.TrimSpace(cfg.DefaultView)
	if viewDef != "compact" && viewDef != "normal" && viewDef != "detail" {
		viewDef = "normal"
	}
	dayExStr := strconv.Itoa(dayEx)
	return map[string]any{
		"CarID": map[string]any{
			"name": "CarID", "in": "path", "required": true,
			"schema": map[string]any{
				"type":        "string",
				"default":     carEx,
				"example":     carEx,
				"description": "车辆 ID（数字字符串）",
			},
		},
		"Token": map[string]any{
			"name": "Token", "in": "path", "required": true,
			"schema": map[string]any{
				"type":        "string",
				"default":     feed,
				"example":     feed,
				"description": "与 CALENDAR_FEED_TOKEN 一致",
			},
		},
		"days": map[string]any{
			"name": "days", "in": "query",
			"schema": map[string]any{
				"type":        "integer",
				"minimum":     1,
				"default":     dayEx,
				"example":     dayEx,
				"description": "回溯天数，受 MAX_DAYS 限制（default 为当前 DEFAULT_DAYS=" + dayExStr + "）",
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
				"type":        "string",
				"default":     tzEx,
				"example":     tzEx,
				"description": "IANA 时区，用于日界与 all-day 显示",
			},
		},
		"range": map[string]any{
			"name": "range", "in": "query",
			"schema": map[string]any{
				"type":    "string",
				"enum":    []string{"day", "week", "month"},
				"default": "day", "description": "仅 daily.ics / all.ics 中日报类汇总：按天、按周、按月",
			},
		},
		"view": map[string]any{
			"name": "view", "in": "query",
			"schema": map[string]any{
				"type":    "string",
				"enum":    []string{"compact", "normal", "detail"},
				"default": viewDef, "example": viewDef, "description": "事件标题与展示密度",
			},
		},
		"detail": map[string]any{
			"name": "detail", "in": "query",
			"schema": map[string]any{
				"type": "boolean", "default": true,
				"description": "是否包含扩展描述；周报/月报时控制是否带分日/分周明细",
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
		"lang": map[string]any{
			"name": "lang", "in": "query",
			"schema": map[string]any{"type": "string", "default": "zh-CN", "description": "预留语言"},
		},
	}
}

func openAPIExamples() map[string]any {
	return map[string]any{
		"ICalendarHeaders": map[string]any{
			"value": map[string]any{
				"Content-Type":        "text/calendar; charset=utf-8",
				"Content-Disposition": "inline; filename=\"teslamate-calendar.ics\"",
				"Cache-Control":       "public, max-age=<CACHE_TTL_SECONDS>",
			},
		},
	}
}

func openAPICommonResponses() map[string]any {
	ref := func(s string) map[string]any {
		return map[string]any{"$ref": "#/components/schemas/" + s}
	}
	return map[string]any{
		"BadRequest": map[string]any{
			"description": "参数错误（如 car id、时间区间、days 越界等）",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": ref("Error"),
				},
			},
		},
		"Forbidden": map[string]any{
			"description": "日历订阅未启用等",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": ref("Error"),
				},
			},
		},
		"GatewayTimeout": map[string]any{
			"description": "上游请求超时",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": ref("Error"),
				},
			},
		},
		"BadGateway": map[string]any{
			"description": "teslamateapi 不可用或非成功响应",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": ref("Error"),
				},
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
				"summary": "Readiness（探测上游）",
				"responses": map[string]any{
					"200": map[string]any{
						"description": "上游可用",
						"content":     map[string]any{"application/json": jsonSchema("ReadyStatus")},
					},
					"502": map[string]any{
						"description": "上游未就绪",
						"content":     map[string]any{"application/json": jsonSchema("Error")},
					},
				},
			},
		},
		"/ping": map[string]any{
			"get": map[string]any{
				"tags":    []string{"Health"},
				"summary": "轻量探针，返回纯文本",
				"responses": map[string]any{
					"200": map[string]any{
						"description": "pong",
						"content": map[string]any{
							"text/plain": map[string]any{
								"schema": map[string]any{
									"type":    "string",
									"example": "pong",
								},
							},
						},
					},
				},
			},
		},
		"/openapi.json": map[string]any{
			"get": map[string]any{
				"tags":    []string{"Docs"},
				"summary": "OpenAPI 3.0 文档 JSON",
				"responses": map[string]any{
					"200": map[string]any{
						"description": "本规范",
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{
									"type":                 "object",
									"additionalProperties": true,
								},
							},
						},
					},
				},
			},
		},
		"/cars": map[string]any{
			"get": map[string]any{
				"tags":     []string{"Cars"},
				"summary":  "车辆列表",
				"security": []any{map[string]any{"BearerAuth": []any{}}},
				"responses": map[string]any{
					"200": map[string]any{
						"description": "成功",
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{
									"type":  "array",
									"items": map[string]any{"$ref": "#/components/schemas/Car"},
								},
							},
						},
					},
					"401": map[string]any{
						"description": "未授权（在启用 REQUIRE_TOKEN_FOR_CARS 时）",
					},
					"502": upErr,
				},
			},
		},
		"/cars/{CarID}": map[string]any{
			"get": map[string]any{
				"tags":       []string{"Cars"},
				"summary":    "单车详情",
				"security":   []any{map[string]any{"BearerAuth": []any{}}},
				"parameters": []any{refParam("CarID")},
				"responses": map[string]any{
					"200": map[string]any{
						"description": "成功",
						"content":     map[string]any{"application/json": jsonSchema("Car")},
					},
					"401": map[string]any{
						"description": "未授权",
					},
					"502": upErr,
				},
			},
		},
		"/calendar/token/{Token}/cars/{CarID}/all.ics":     icsPath("合并行程+充电+日报汇总+更新（all.ics 中日报是否按周/月由 range 决定）", true, true, true, true, true),
		"/calendar/token/{Token}/cars/{CarID}/daily.ics":   icsPath("按日/周/月汇总（range）", true, true, true, true, true),
		"/calendar/token/{Token}/cars/{CarID}/drives.ics":  icsPath("行程", true, true, true, true, true),
		"/calendar/token/{Token}/cars/{CarID}/charges.ics": icsPath("充电", true, true, true, true, true),
		"/calendar/token/{Token}/cars/{CarID}/updates.ics": icsPath("软件更新", true, true, true, true, true),
	}
	return paths
}

func icsPath(desc string, days, se, tz, rng, vdet bool) map[string]any {
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
		params = append(params, refParam("detail"), refParam("vehicleName"), refParam("minDistance"), refParam("maxDistance"), refParam("lang"))
	}
	op := map[string]any{
		"tags":        []string{"Calendar"},
		"summary":     "ICS: " + desc,
		"description": "路径中的 Token 与请求头/查询中的 token 需与 `CALENDAR_FEED_TOKEN` 一致。",
		"parameters":  params,
		"responses": map[string]any{
			"200": map[string]any{
				"description": "iCalendar 正文",
				"content": map[string]any{
					"text/calendar": map[string]any{
						"schema": map[string]any{"$ref": "#/components/schemas/ICalendar"},
					},
				},
			},
			"400": map[string]any{"$ref": "#/components/responses/BadRequest"},
			"403": map[string]any{"$ref": "#/components/responses/Forbidden"},
			"504": map[string]any{"$ref": "#/components/responses/GatewayTimeout"},
			"502": map[string]any{"$ref": "#/components/responses/BadGateway"},
		},
	}
	return map[string]any{"get": op}
}
