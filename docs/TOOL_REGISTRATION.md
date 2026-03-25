# 工具注册 JSON 格式说明

**接口：** `POST /api/v1/tools`

工具分三种类型，通过 `tool_type` 字段区分：

| tool_type | 说明 |
|-----------|------|
| `native`  | 使用内置模板，MCP-Agent 直接执行，无需部署外部服务 |
| `custom`  | 用户自定义 HTTP endpoint，MCP-Agent 作为代理转发 |
| `remote`  | 注册到外部 MCP Server（原有方式）|

---

## Type A：原生内置工具 (`tool_type: "native"`)

### 模板 1 — `http_get`：GET 请求获取 API 内容

```json
{
  "name": "get_weather",
  "description": "获取指定城市的实时天气信息",
  "tool_type": "native",
  "version": "1.0.0",
  "native_config": {
    "template": "http_get",
    "url": "https://api.weatherapi.com/v1/current.json?key=YOUR_KEY&q={city}",
    "headers": {
      "Accept": "application/json"
    }
  },
  "schema": {
    "type": "object",
    "required": ["city"],
    "properties": {
      "city": {
        "type": "string",
        "description": "城市名称，如 Beijing"
      }
    }
  }
}
```

> `url` 中的 `{city}` 会被调用时传入的 `arguments.city` 自动替换。

---

### 模板 2 — `http_post`：POST 请求

```json
{
  "name": "translate_text",
  "description": "调用翻译 API 将文本翻译为目标语言",
  "tool_type": "native",
  "version": "1.0.0",
  "native_config": {
    "template": "http_post",
    "url": "https://api.deepl.com/v2/translate",
    "headers": {
      "Authorization": "DeepL-Auth-Key YOUR_KEY",
      "Content-Type": "application/json"
    },
    "body_template": "{\"text\": [\"{text}\"], \"target_lang\": \"{target_lang}\"}"
  },
  "schema": {
    "type": "object",
    "required": ["text", "target_lang"],
    "properties": {
      "text": {
        "type": "string",
        "description": "需要翻译的文本"
      },
      "target_lang": {
        "type": "string",
        "description": "目标语言代码，如 ZH、EN、JA"
      }
    }
  }
}
```

> `body_template` 是字符串模板，`{text}` 和 `{target_lang}` 会被 `arguments` 中对应字段替换。

---

### 模板 3 — `read_doc`：读取文档/网页内容

```json
{
  "name": "read_confluence_page",
  "description": "读取 Confluence 文档页面内容",
  "tool_type": "native",
  "version": "1.0.0",
  "native_config": {
    "template": "read_doc",
    "url": "https://your-domain.atlassian.net/wiki/rest/api/content/{page_id}?expand=body.storage",
    "headers": {
      "Authorization": "Basic YOUR_BASE64_TOKEN",
      "Accept": "application/json"
    }
  },
  "schema": {
    "type": "object",
    "required": ["page_id"],
    "properties": {
      "page_id": {
        "type": "string",
        "description": "Confluence 页面 ID"
      }
    }
  }
}
```

> `read_doc` 与 `http_get` 行为相同，语义上用于文档读取场景，便于 AI 检索时区分用途。

---

## Type C：完全自定义工具 (`tool_type: "custom"`)

用户已有内部 HTTP 服务，不需要改造为 MCP Server，MCP-Agent 将 `arguments` 作为 JSON body 直接 POST 到 `server_url`。

```json
{
  "name": "internal_order_query",
  "description": "查询内部订单系统中指定用户的订单列表",
  "tool_type": "custom",
  "server_url": "http://internal-erp.company.com/api/orders/query",
  "version": "2.1.0",
  "schema": {
    "type": "object",
    "required": ["user_id"],
    "properties": {
      "user_id": {
        "type": "string",
        "description": "用户 ID"
      },
      "status": {
        "type": "string",
        "description": "订单状态筛选：pending / completed / cancelled"
      },
      "page": {
        "type": "integer",
        "description": "页码，默认 1"
      }
    }
  }
}
```

> `custom` 类型无需 `native_config`，调用时 MCP-Agent 将整个 `arguments` 对象作为 JSON body POST 到 `server_url`。

---

## 工具调用格式

**接口：** `POST /api/v1/tools/{name}/call`

```json
{
  "arguments": {
    "city": "Beijing"
  }
}
```

**响应：**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "status": 200,
    "body": "{\"location\": {\"name\": \"Beijing\"}, \"current\": {\"temp_c\": 22}}"
  }
}
```

---

## AI Agent 执行（自动工具选择）

**接口：** `POST /api/v1/agent/execute`

无需指定工具名，Agent 会根据任务描述自动检索合适的工具并执行：

```json
{
  "task": "帮我查一下北京今天的天气，然后翻译成英文",
  "max_steps": 5
}
```

---

## 字段速查表

| 字段 | 类型 | 适用类型 | 必填 | 说明 |
|------|------|---------|------|------|
| `name` | string | 所有 | ✅ | 唯一工具名 |
| `description` | string | 所有 | 推荐 | 工具描述，用于 AI 向量检索，描述越详细检索越准确 |
| `tool_type` | string | 所有 | ✅ | `native` / `custom` / `remote` |
| `native_config.template` | string | native | ✅ | `http_get` / `http_post` / `read_doc` |
| `native_config.url` | string | native | ✅ | 请求 URL，支持 `{param}` 占位符 |
| `native_config.headers` | object | native | ❌ | HTTP 请求头键值对 |
| `native_config.body_template` | string | native (post) | ❌ | POST body 字符串模板，支持 `{param}` |
| `server_url` | string | custom / remote | ✅ | 目标服务 HTTP 地址 |
| `schema` | JSON Schema | 所有 | 推荐 | 参数校验规则，遵循 JSON Schema 规范 |
| `version` | string | 所有 | ❌ | 工具版本号，默认 `1.0.0` |
| `server_name` | string | remote | ❌ | 所属 MCP Server 名称（仅 remote 使用）|

---

## 占位符替换规则

`native` 类型支持在 `url` 和 `body_template` 中使用 `{参数名}` 占位符：

- 占位符格式：`{key}`
- 替换来源：调用时 `arguments` 中对应的字段值
- 未匹配的占位符保持原样
- `body_template` 中如需嵌套 JSON，需对内部引号转义：`\"`
