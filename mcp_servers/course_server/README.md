# Course Server

示例 MCP 工具服务 - 课程管理

## 运行

```bash
go run main.go
```

服务将在 `http://localhost:8081` 启动。

## 接口

### 健康检查
```
GET /health
```

### 工具调用
```
POST /call
Content-Type: application/json

{
  "student_id": "20230001"
}
```

响应：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "student_id": "20230001",
    "courses": [
      {
        "id": "CS101",
        "name": "计算机科学导论",
        "teacher": "张教授",
        "time": "周一 09:00-11:00",
        "location": "教学楼A101"
      }
    ]
  }
}
```

## 注册到 MCP-Agent

通过管理 API 注册工具：

```bash
curl -X POST http://localhost:8000/api/v1/tools \
  -H "Authorization: Bearer <admin-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "get_schedule",
    "description": "获取学生课程表",
    "server_name": "course_server",
    "server_url": "http://localhost:8081",
    "schema": {
      "type": "object",
      "properties": {
        "student_id": {
          "type": "string",
          "description": "学生ID"
        }
      },
      "required": ["student_id"]
    },
    "version": "1.0.0"
  }'
```
