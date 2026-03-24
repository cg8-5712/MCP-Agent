# MCP-Agent API 使用示例

## 认证

### 登录获取 Token

```bash
curl -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'
```

响应:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "username": "admin",
      "role": "admin"
    }
  }
}
```

## 工具管理

### 注册新工具

```bash
curl -X POST http://localhost:8000/api/v1/tools \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "get_weather",
    "description": "Get current weather information for a city",
    "server_name": "weather_server",
    "server_url": "http://localhost:8081",
    "schema": {
      "type": "object",
      "properties": {
        "city": {
          "type": "string",
          "description": "City name"
        }
      },
      "required": ["city"]
    }
  }'
```

### 获取工具列表

```bash
curl -X GET http://localhost:8000/api/v1/tools \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 调用工具

```bash
curl -X POST http://localhost:8000/api/v1/tools/get_weather/call \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "arguments": {
      "city": "Beijing"
    }
  }'
```

## AI Agent 功能

### 执行 AI Agent 任务

```bash
curl -X POST http://localhost:8000/api/v1/agent/execute \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "What is the weather in Beijing and Shanghai?"
  }'
```

响应:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "answer": "Based on the weather data, Beijing is currently sunny with 25°C, while Shanghai is cloudy with 22°C.",
    "plan": {
      "steps": [
        {
          "step_id": 1,
          "tool_name": "get_weather",
          "arguments": {"city": "Beijing"},
          "description": "Get weather for Beijing"
        },
        {
          "step_id": 2,
          "tool_name": "get_weather",
          "arguments": {"city": "Shanghai"},
          "description": "Get weather for Shanghai"
        }
      ]
    },
    "executions": [
      {
        "step_id": 1,
        "tool_name": "get_weather",
        "arguments": {"city": "Beijing"},
        "result": {"temperature": 25, "condition": "sunny"},
        "duration_ms": 120
      },
      {
        "step_id": 2,
        "tool_name": "get_weather",
        "arguments": {"city": "Shanghai"},
        "result": {"temperature": 22, "condition": "cloudy"},
        "duration_ms": 115
      }
    ],
    "total_time_ms": 1250
  }
}
```

### 向量检索工具

```bash
curl -X POST http://localhost:8000/api/v1/agent/search-tools \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "weather information",
    "top_k": 3
  }'
```

响应:
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "name": "get_weather",
      "description": "Get current weather information for a city",
      "score": 0.92,
      "schema": {
        "type": "object",
        "properties": {
          "city": {"type": "string"}
        }
      }
    }
  ]
}
```

### 重新索引工具（管理员）

```bash
curl -X POST http://localhost:8000/api/v1/agent/index-tools \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 监控与日志

### 健康检查

```bash
curl -X GET http://localhost:8000/api/v1/health
```

### 查询调用日志（管理员）

```bash
curl -X GET "http://localhost:8000/api/v1/logs?tool_name=get_weather&page=1&page_size=20" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 获取工具统计信息

```bash
curl -X GET http://localhost:8000/api/v1/tools/get_weather/stats \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 配置说明

### 环境变量

在运行前设置以下环境变量:

```bash
export LLM_API_KEY="your-openai-api-key"
export EMBEDDING_API_KEY="your-openai-api-key"
export JWT_SECRET="your-secret-key"
```

### 配置文件

编辑 `configs/config.yaml`:

```yaml
llm:
  provider: "openai"
  api_key: "${LLM_API_KEY}"
  base_url: "https://api.openai.com/v1"
  model:
    planner: "gpt-4o-mini"
    final: "gpt-4o"
  temperature: 0.7
  max_tokens: 2000

embedding:
  provider: "openai"
  api_key: "${EMBEDDING_API_KEY}"
  base_url: "https://api.openai.com/v1"
  model: "text-embedding-3-small"
  dimension: 1536

vector_db:
  use_memory: true  # 开发环境使用内存模式
```

## 完整示例：创建天气查询工具

1. 启动 MCP-Agent:
```bash
./mcp-agent
```

2. 登录获取 Token:
```bash
TOKEN=$(curl -s -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | jq -r '.data.token')
```

3. 注册天气工具:
```bash
curl -X POST http://localhost:8000/api/v1/tools \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "get_weather",
    "description": "Get current weather information for a city",
    "server_url": "http://localhost:8081",
    "schema": {
      "type": "object",
      "properties": {
        "city": {"type": "string"}
      },
      "required": ["city"]
    }
  }'
```

4. 使用 AI Agent 查询天气:
```bash
curl -X POST http://localhost:8000/api/v1/agent/execute \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"query": "What is the weather in Beijing?"}'
```

## 错误处理

所有错误响应遵循统一格式:

```json
{
  "code": 400,
  "message": "error description"
}
```

常见错误码:
- 400: 请求参数错误
- 401: 未认证或 Token 无效
- 403: 权限不足
- 404: 资源不存在
- 500: 服务器内部错误
