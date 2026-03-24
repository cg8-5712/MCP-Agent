# MCP-Agent

<p align="center">
  <strong>基于 Go 的 MCP Server 管理与工具调用系统</strong>
</p>

<p align="center">
  <a href="#特性">特性</a> •
  <a href="#开发计划">开发计划</a> •
  <a href="#快速开始">快速开始</a> •
  <a href="#架构">架构</a> •
  <a href="#api-文档">API 文档</a> •
  <a href="#贡献">贡献</a>
</p>

---

## 简介

MCP-Agent 是一个专注于系统化管理工具、调用权限控制及实时监控的 MCP Server 管理系统。通过向量检索、LLM 规划和工具执行，实现智能化的多步骤任务编排。

## 特性

- **智能工具检索**：基于向量数据库（FAISS/pgvector/Milvus）的语义检索
- **多步骤规划**：LLM 驱动的任务分解与执行计划生成
- **权限管理**：基于 RBAC 的细粒度权限控制
- **实时监控**：工具健康检查、调用日志、性能统计
- **热插拔架构**：工具注册无需修改核心代码
- **Web 管理界面**：可视化工具管理、日志查询、权限配置

## 开发计划

### Phase 1: 基础设施（已完成）

**目标**：搭建核心框架与基础服务

- [x] 项目结构设计与文档编写
- [x] 数据库模型设计（工具、日志、权限）
- [x] MCP Gateway 实现（路由分发、中间件）
- [x] JWT 认证模块
- [x] RBAC 权限校验模块
- [x] 基础 API 接口（工具 CRUD、健康检查）
- [x] 配置文件加载（Viper + YAML）

**已交付**：可运行的 MCP Gateway + 基础管理 API

---

### Phase 2: 工具管理与监控（已完成）

**目标**：实现工具注册、调用与监控

- [x] 工具注册与参数 Schema 校验
- [x] 工具调用代理（HTTP 转发）
- [x] 调用日志记录（SQLite 持久化）
- [x] 健康检查定时任务
- [x] 工具状态监控与统计
- [x] 日志查询 API（分页、筛选）

**已交付**：完整的工具管理与监控系统

---

### Phase 3: 向量检索与 LLM 集成（已完成）

**目标**：实现智能工具检索与任务规划

- [x] 工具描述向量化（集成 Embedding 模型）
- [x] 向量数据库集成（内存模式/Milvus）
- [x] Tool Embedding Search API
- [x] Planner LLM 集成（生成执行计划）
- [x] Executor 实现（多步骤任务执行）
- [x] Final LLM 集成（自然语言回答生成）
- [x] 任务链路追踪与错误回溯

**已交付**：端到端的 AI Agent 调用链路

---

### Phase 4: Web 管理界面

**目标**：提供可视化管理与监控界面

- [ ] 工具管理页面（增删改查、启用/禁用）
- [ ] 权限配置可视化编辑
- [ ] 调用日志查询与导出
- [ ] 工具状态仪表盘（在线率、调用统计、响应时间）
- [ ] 任务失败诊断面板
- [ ] 用户认证与角色管理

**预期交付**：完整的 Web 管理界面

---

### Phase 5: 优化与扩展

**目标**：性能优化与功能扩展

- [ ] 工具调用缓存机制
- [ ] 并发调用优化
- [ ] 分布式部署支持
- [ ] 工具版本管理
- [ ] Webhook 通知（工具失败、异常告警）
- [ ] 多租户支持
- [ ] API 限流与熔断
- [ ] 完整的单元测试与集成测试

**预期交付**：生产级可用的 MCP-Agent 系统

---

### 里程碑

| 阶段    | 预计完成时间 | 状态       |
| ------- | ------------ | ---------- |
| Phase 1 | 2026-03     | ✅ 已完成  |
| Phase 2 | 2026-03     | ✅ 已完成  |
| Phase 3 | 2026-03     | ✅ 已完成  |
| Phase 4 | 2026-07     | 未开始     |
| Phase 5 | 2026-08     | 未开始     |

## 技术栈

| 类别       | 选型                          |
| ---------- | ----------------------------- |
| 语言       | Go 1.21+                      |
| Web 框架   | Gin                           |
| 数据库     | GORM + PostgreSQL/SQLite/Mysql|
| 配置格式   | YAML                          |
| 认证       | JWT                           |
| 前端       | react                         |

## 快速开始

### 环境要求

- Go 1.21+
- PostgreSQL 或 SQLite / MySQL
- （可选）向量数据库：Milvus

### 安装

```bash
# 克隆仓库
git clone https://github.com/cg8-5712/MCP-Agent.git
cd MCP-Agent

# 安装依赖（国内用户推荐使用代理）
GOPROXY=https://goproxy.cn,direct go mod tidy

# 运行
go run cmd/server/main.go

# 或编译后运行
go build -o mcp-agent cmd/server/main.go
./mcp-agent
```

**默认账号**：
- 用户名: `admin`
- 密码: `admin123`

### 配置

编辑 `configs/config.yaml`：

```yaml
server:
  port: 8000
  mode: "release"

jwt:
  secret: "${JWT_SECRET}"
  expire_hours: 24

database:
  driver: "sqlite"
  dsn: "./data/mcp_agent.db"

mcp_servers:
  - name: "course_server"
    url: "http://localhost:8081"

# LLM 配置
llm:
  provider: "openai"
  api_key: "${LLM_API_KEY}"
  model:
    planner: "gpt-4o-mini"
    final: "gpt-4o"

# Embedding 配置
embedding:
  provider: "openai"
  api_key: "${EMBEDDING_API_KEY}"
  model: "text-embedding-3-small"

# 向量数据库配置
vector_db:
  use_memory: true  # 开发环境使用内存模式
```

权限配置 `configs/permissions.yaml`：

```yaml
permissions:
  - role: "student"
    tools: ["get_schedule", "search_course"]
  - role: "admin"
    tools: ["all"]
```

## 架构

### AI Agent 调用链路

```
API Gateway → Tool Embedding Search → Planner LLM → Executor → Final LLM
```

1. **Tool Embedding Search**：向量检索最相关工具
2. **Planner LLM**：生成多步骤执行计划（JSON）
3. **Executor**：调用工具并追踪执行链路
4. **Final LLM**：生成自然语言回答

### 服务层架构

```
AI Agent ←→ MCP Gateway ←→ Tool Servers (课程/通知/管理)
                ↓
          管理模块 (工具管理/权限/日志/监控)
```

详细架构请参考 [CLAUDE.md](./CLAUDE.md)。

## API 文档

### 统一响应格式

```json
{
  "code": 200,
  "message": "success",
  "data": {}
}
```

### 核心接口

| 方法   | 路径                        | 说明                   | 鉴权   |
| ------ | --------------------------- | ---------------------- | ------ |
| POST   | `/api/v1/tools/:name/call`  | 调用指定工具           | 需要   |
| GET    | `/api/v1/tools`             | 获取所有工具列表       | 需要   |
| POST   | `/api/v1/tools`             | 注册新工具             | 管理员 |
| GET    | `/api/v1/health`            | 全局健康检查           | 公开   |
| GET    | `/api/v1/logs`              | 查询调用日志           | 管理员 |
| POST   | `/api/v1/auth/login`        | 登录获取 Token         | 公开   |
| POST   | `/api/v1/agent/execute`     | AI Agent 执行任务      | 需要   |
| POST   | `/api/v1/agent/search-tools`| 向量检索工具           | 需要   |
| POST   | `/api/v1/agent/index-tools` | 重新索引工具向量       | 管理员 |

完整 API 文档请参考 [CLAUDE.md#API接口设计](./CLAUDE.md#api-接口设计)。

## 项目结构

```
MCP-Agent/
├── cmd/server/main.go    # 主入口
├── internal/
│   ├── config/           # 配置加载（Viper）
│   ├── constants/        # 常量定义
│   ├── model/            # 数据模型（User/Tool/CallLog）
│   ├── repository/       # 数据访问层（GORM）
│   ├── service/          # 业务逻辑层
│   ├── handler/          # HTTP 处理层（Gin）
│   ├── middleware/       # 中间件（Auth/CORS/Logger/Permission）
│   └── permission/       # RBAC 权限管理
├── pkg/
│   ├── logger/           # Zap 日志封装
│   ├── llm/              # LLM 客户端封装
│   ├── embedding/        # Embedding 客户端封装
│   ├── vectordb/         # 向量数据库接口与实现
│   └── utils/            # 工具函数
├── configs/
│   ├── config.yaml       # 主配置
│   └── permissions.yaml  # 权限配置
├── mcp_servers/          # MCP 工具服务（待实现）
└── web/                  # Web 前端（git submodule）
```

## 开发

### 添加新工具

1. 在 `mcp_servers/` 下创建工具服务目录
2. 实现工具的 HTTP 接口（需包含 `/health` 端点）
3. 在 `configs/config.yaml` 中注册工具
4. 通过管理 API 添加工具描述与参数 Schema

### 运行测试

```bash
go test ./...
```

### 构建

```bash
go build -o mcp-agent cmd/server/main.go
```

## 贡献

欢迎贡献！请先阅读 [CODE_OF_CONDUCT.md](./CODE_OF_CONDUCT.md) 和 [CLAUDE.md](./CLAUDE.md)。

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](./LICENSE) 文件。

## 联系方式

项目链接：[https://github.com/cg8-5712/MCP-Agent](https://github.com/cg8-5712/MCP-Agent)

---

<p align="center">Made with ❤️ by MCP-Agent Team</p>
