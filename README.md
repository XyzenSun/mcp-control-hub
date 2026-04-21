# MCP Control Hub

MCP Control Hub 是一个 MCP (Model Context Protocol) 网关服务，用于集中管理和代理多个上游 MCP 服务器。

## 功能特性

- **服务器管理** - 注册、启用/禁用、同步多个上游 MCP 服务器
- **工具发现** - 自动发现并注册上游服务器提供的工具
- **命名空间隔离** - 通过命名空间组织工具，支持灵活的路由配置
- **API Key 认证** - 安全的 API 访问控制
- **多数据库支持** - SQLite、MySQL、PostgreSQL
- **Web UI** - 内置管理界面
- **Docker 部署** - 开箱即用的容器化支持

## 快速开始

### 使用 Docker Compose (推荐)

```bash
# 创建数据目录
mkdir -p data

# 启动服务
docker-compose up -d
```

服务将在 `http://localhost:8080` 启动。

### 使用 Docker

```bash
docker run -d \
  --name mcp-control-hub \
  -p 8080:8080 \
  -e BOOTSTRAP_API_KEY=sk-your-api-key \
  -e GATEWAY_DATABASE_DRIVER=sqlite \
  -e GATEWAY_DATABASE_DSN=/app/data/gateway.db \
  -v ./data:/app/data \
  ghcr.io/xyzensun/mcp-control-hub:latest
```

### 从源码构建

```bash
# 克隆仓库
git clone https://github.com/xyzensun/mcp-control-hub.git
cd mcp-control-hub

# 构建
go build -o mcp-control-hub ./cmd/gateway

# 运行
./mcp-control-hub
```

## 配置

### 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `BOOTSTRAP_API_KEY` | 初始 API Key (首次启动时创建) | - |
| `GATEWAY_SERVER_HOST` | 服务监听地址 | `0.0.0.0` |
| `GATEWAY_SERVER_PORT` | 服务监听端口 | `8080` |
| `GATEWAY_SERVER_MODE` | 运行模式 (`debug`/`release`) | `release` |
| `GATEWAY_DATABASE_DRIVER` | 数据库类型 (`sqlite`/`mysql`/`postgres`) | `sqlite` |
| `GATEWAY_DATABASE_DSN` | 数据库连接字符串 | `gateway.db` |
| `GATEWAY_LOGGING_LEVEL` | 日志级别 | `info` |
| `GATEWAY_LOGGING_FORMAT` | 日志格式 (`json`/`text`) | `json` |

### 配置文件

参考 `configs/config.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  mode: "release"

database:
  driver: "sqlite"
  dsn: "gateway.db"

logging:
  level: "info"
  format: "json"

mcp:
  session_ttl: 1h
  health_check_interval: 30s

security:
  api_key_length: 32
  rate_limit: 100
```

## API 接口

### 认证

所有 API 请求需要在 Header 中携带 API Key：

```
X-API-Key: sk-your-api-key
```

### 服务器管理

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/api/v1/servers` | 创建服务器 |
| `GET` | `/api/v1/servers` | 获取服务器列表 |
| `GET` | `/api/v1/servers/:id` | 获取服务器详情 |
| `PUT` | `/api/v1/servers/:id` | 更新服务器 |
| `DELETE` | `/api/v1/servers/:id` | 删除服务器 |
| `POST` | `/api/v1/servers/:id/enable` | 启用服务器 |
| `POST` | `/api/v1/servers/:id/disable` | 禁用服务器 |
| `POST` | `/api/v1/servers/:id/sync` | 同步服务器工具 |

### 工具管理

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/api/v1/tools` | 获取工具列表 |
| `GET` | `/api/v1/tools/:id` | 获取工具详情 |
| `PUT` | `/api/v1/tools/:id` | 更新工具 |
| `POST` | `/api/v1/tools/:id/enable` | 启用工具 |
| `POST` | `/api/v1/tools/:id/disable` | 禁用工具 |
| `POST` | `/api/v1/tools/refresh` | 刷新所有工具 |

### 命名空间管理

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/api/v1/namespaces` | 创建命名空间 |
| `GET` | `/api/v1/namespaces` | 获取命名空间列表 |
| `GET` | `/api/v1/namespaces/:id` | 获取命名空间详情 |
| `PUT` | `/api/v1/namespaces/:id` | 更新命名空间 |
| `DELETE` | `/api/v1/namespaces/:id` | 删除命名空间 |
| `POST` | `/api/v1/namespaces/:id/tools` | 添加工具到命名空间 |
| `DELETE` | `/api/v1/namespaces/:id/tools/:tool_id` | 从命名空间移除工具 |

### API Key 管理

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/api/v1/apikeys` | 创建 API Key |
| `GET` | `/api/v1/apikeys` | 获取 API Key 列表 |
| `DELETE` | `/api/v1/apikeys/:id` | 删除 API Key |

### MCP 端点

MCP 客户端通过以下端点连接：

```
POST /mcp/:apikey/:namespace
GET /mcp/:apikey/:namespace
```

## 支持的上游服务器协议

| 协议 | 说明 |
|------|------|
| `stdio` | 标准输入输出协议 |
| `sse` | Server-Sent Events |
| `streamable` | 可流式传输协议 |

## 开发

### 项目结构

```
.
├── cmd/
│   └── gateway/main.go     # 入口文件
├── configs/
│   └── config.yaml         # 配置示例
├── internal/
│   ├── api/                # HTTP API 层
│   ├── config/             # 配置加载
│   ├── database/           # 数据库初始化
│   ├── mcp/                # MCP 协议实现
│   ├── models/             # 数据模型
│   ├── repository/         # 数据访问层
│   └── service/            # 业务逻辑层
├── pkg/
│   ├── logger/             # 日志工具
│   └── utils/              # 工具函数
├── Dockerfile
├── docker-compose.yml
└── go.mod
```

### 构建

```bash
go build -o mcp-control-hub ./cmd/gateway
```

## 许可证

MIT License