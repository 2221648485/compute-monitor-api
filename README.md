# Compute Monitor API

Compute Monitor API 是面向 Kubernetes 集群的算力监控与算力迁移后端系统。系统统一纳管多集群配置，持续同步 Kubernetes 资源，接入 Prometheus 指标，提供节点、Pod、GPU、工作负载、告警、审计、迁移调度和兼容调度器接口，支撑算力资源观测、容量评估、迁移决策和自动化运维。

## 核心能力

- **统一认证与权限控制**：支持账号密码登录、JWT access token、refresh token、Redis 会话、密码修改、当前用户信息查询和管理员角色控制。
- **后台用户管理**：支持用户创建、分页查询、详情查询、信息修改、启停状态控制和管理员重置密码。
- **多集群纳管**：支持 Kubernetes 集群配置的新增、查询、更新、删除和连接测试，按集群维护 kubeconfig 路径和 Prometheus 地址。
- **Kubernetes 资源同步**：通过 client-go 同步 Namespace、Node、Pod、Deployment、Service 等资源，并写入 MySQL 缓存表。
- **节点与工作负载查询**：提供 Namespace、Node、节点详情、节点 Pod、Deployment、Service、Pod 日志和 YAML 原生资源操作能力。
- **Prometheus 指标查询**：支持 CPU、内存、磁盘、网络、GPU 等时间序列指标查询，并支持 `start`、`end`、`step` 控制查询窗口和粒度。
- **GPU 资源管理**：基于 Kubernetes Node GPU 容量和 Prometheus DCGM 指标，提供 GPU 列表、节点 GPU 列表、GPU 汇总、Top 查询和节点 GPU 指标。
- **算力迁移调度接口**：提供 `/api/v2` 兼容接口，面向旧前端、外部调度器和算力迁移组件输出集群、节点、应用、实例、资源摘要和批量启停/删除能力。
- **后台定时任务**：支持按配置周期自动同步集群 Kubernetes 资源，保证数据库缓存与集群状态持续对齐。
- **统一响应与错误处理**：所有接口使用统一 JSON 响应结构，基础设施不可用时统一返回 503，避免启动期空指针异常。
- **Swagger API 文档**：非生产环境自动开放 Swagger UI，便于联调和接口验收。
- **容器化部署**：提供 Dockerfile 和 docker-compose，支持 MySQL、Redis 和后端服务一键启动。

## 技术栈

| 类别 | 技术 |
| --- | --- |
| 语言 | Go 1.26.3 |
| HTTP 框架 | Gin |
| ORM | GORM |
| 数据库 | MySQL |
| 缓存 / 会话 | Redis |
| 认证 | JWT + Redis refresh session |
| Kubernetes | client-go |
| 监控指标 | Prometheus Go client |
| API 文档 | swaggo / gin-swagger |
| 部署 | Docker / Docker Compose |

## 架构概览

```text
HTTP Client / Scheduler / Frontend
        |
        v
Gin Router
        |
        v
Middleware
  - RequestLog
  - Auth
  - RequireRole
        |
        v
Handler
        |
        v
Service
        |
        +--------------------+
        |                    |
        v                    v
Repository              External Clients
  - MySQL/GORM            - Kubernetes client-go
  - Redis Session         - Prometheus HTTP API
        |
        v
MySQL / Redis / Kubernetes / Prometheus
```

系统按业务模块组织代码，每个核心业务包采用 `dto.go`、`model.go`、`repository.go`、`service.go`、`handler.go`、`router.go` 的分层方式。`internal/app` 是应用装配层，只负责依赖初始化、模块装配和路由注册，不写业务逻辑。

## 目录结构

```text
compute-monitor-api/
  cmd/server/              # 服务启动入口
  configs/                 # 分环境配置文件
  deployments/             # docker-compose 等部署文件
  docs/                    # Swagger 和项目文档
  internal/
    app/                   # 应用装配层
    auth/                  # 登录、token、session、密码
    user/                  # 后台用户管理
    cluster/               # 集群配置管理
    k8s/                   # K8s client、资源模型和缓存仓储
    k8ssync/               # Kubernetes 资源同步
    node/                  # Namespace、Node、Pod 查询
    metrics/               # Prometheus 指标查询和缓存
    gpu/                   # GPU 资源和 GPU 指标
    compat/                # /api/v2 调度器兼容接口
    scheduler/             # 后台定时同步任务
    middleware/            # 鉴权、日志等中间件
    response/              # 统一响应结构
    store/                 # MySQL、Redis 基础设施
  Dockerfile
  go.mod
```

## 数据流

1. 管理员在后台创建集群配置，填写集群 ID、名称、kubeconfig 路径和 Prometheus 地址。
2. 系统通过 kubeconfig 创建 Kubernetes client-go 客户端。
3. `k8ssync` 或后台 scheduler 周期性读取 Kubernetes 资源。
4. Namespace、Node、Pod、Deployment、Service 等资源写入 MySQL 缓存表。
5. 前端和调度器查询资源时优先读取 MySQL 缓存，避免每次请求都实时访问 Kubernetes API。
6. CPU、内存、磁盘、网络、GPU 等动态指标从 Prometheus 查询。
7. 迁移调度组件通过 `/api/v2` 获取资源摘要、节点负载、应用实例和批量操作能力。

## 接口分组

### 基础接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/healthz` | 服务健康检查 |
| GET | `/swagger/index.html` | Swagger UI，非生产环境开放 |

### 认证接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/api/admin/auth/login` | 用户登录 |
| POST | `/api/admin/auth/refresh` | 刷新 token |
| GET | `/api/admin/auth/me` | 查询当前用户 |
| PUT | `/api/admin/auth/password` | 修改当前用户密码 |

### 用户管理

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/admin/users` | 分页查询用户 |
| POST | `/api/admin/users` | 创建用户 |
| GET | `/api/admin/users/{userId}` | 查询用户详情 |
| PUT | `/api/admin/users/{userId}` | 修改用户信息 |
| PUT | `/api/admin/users/{userId}/status` | 启用或禁用用户 |
| PUT | `/api/admin/users/{userId}/password` | 重置用户密码 |

### 集群管理

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/api/admin/clusters` | 创建集群配置 |
| GET | `/api/admin/clusters` | 分页查询集群 |
| GET | `/api/admin/clusters/{clusterId}` | 查询集群详情 |
| PUT | `/api/admin/clusters/{clusterId}` | 修改集群配置 |
| POST | `/api/admin/clusters/{clusterId}/test` | 测试集群连接 |
| DELETE | `/api/admin/clusters/{clusterId}` | 删除集群配置及关联缓存 |

### Kubernetes 同步

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/api/admin/clusters/{clusterId}/k8s/sync` | 同步全部资源 |
| POST | `/api/admin/clusters/{clusterId}/k8s/sync/namespaces` | 同步 Namespace |
| POST | `/api/admin/clusters/{clusterId}/k8s/sync/nodes` | 同步 Node |
| POST | `/api/admin/clusters/{clusterId}/k8s/sync/pods` | 同步 Pod |
| POST | `/api/admin/clusters/{clusterId}/k8s/sync/deployments` | 同步 Deployment |
| POST | `/api/admin/clusters/{clusterId}/k8s/sync/services` | 同步 Service |

### 节点与基础资源

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/admin/clusters/{clusterId}/namespaces` | 查询 Namespace |
| GET | `/api/admin/clusters/{clusterId}/nodes` | 查询 Node |
| GET | `/api/admin/clusters/{clusterId}/nodes/{nodeName}` | 查询 Node 详情 |
| GET | `/api/admin/clusters/{clusterId}/nodes/{nodeName}/pods` | 查询节点 Pod |

### 指标与 GPU

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/admin/clusters/{clusterId}/metrics/node/{nodeName}/cpu` | 查询节点 CPU 指标 |
| GET | `/api/admin/clusters/{clusterId}/metrics/node/{nodeName}/memory` | 查询节点内存指标 |
| GET | `/api/admin/clusters/{clusterId}/metrics/node/{nodeName}/disk` | 查询节点磁盘指标 |
| GET | `/api/admin/clusters/{clusterId}/metrics/node/{nodeName}/network` | 查询节点网络指标 |
| GET | `/api/admin/clusters/{clusterId}/gpus` | 查询集群 GPU 列表 |
| GET | `/api/admin/clusters/{clusterId}/nodes/{nodeName}/gpus` | 查询节点 GPU 列表 |
| GET | `/api/admin/clusters/{clusterId}/gpus/summary` | 查询 GPU 汇总 |
| GET | `/api/admin/clusters/{clusterId}/gpus/top` | 查询 GPU Top 列表 |
| GET | `/api/admin/clusters/{clusterId}/metrics/node/{nodeName}/gpu` | 查询节点 GPU 指标 |

指标接口支持以下 query 参数：

| 参数 | 说明 |
| --- | --- |
| `start` | 查询开始时间，Unix 秒 |
| `end` | 查询结束时间，Unix 秒 |
| `step` | 查询粒度，单位秒 |

### 调度器兼容接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/v2/clusters/exclusive/list` | 查询调度器可见集群 |
| GET | `/api/v2/clusters/{clusterId}/summary/static` | 查询静态容量摘要 |
| GET | `/api/v2/clusters/{clusterId}/summary/dynamic` | 查询动态使用率摘要 |
| GET | `/api/v2/clusters/{clusterId}/nodes` | 查询节点列表 |
| GET | `/api/v2/clusters/{clusterId}/nodes/resource-consumption` | 查询节点资源使用率 |
| GET | `/api/v2/clusters/{clusterId}/apps` | 查询应用列表 |
| GET | `/api/v2/clusters/{clusterId}/instances` | 查询实例列表 |
| GET | `/api/v2/clusters/{clusterId}/native/Deployment` | 查询 Deployment 缓存 |
| GET | `/api/v2/clusters/{clusterId}/metric/nodes/{nodeName}/metrics` | 查询节点指标 |
| POST | `/api/v2/clusters/{clusterId}/native` | 通过 YAML 创建原生资源 |
| DELETE | `/api/v2/clusters/{clusterId}/native/Deployment/{name}` | 删除 Deployment |
| PUT | `/api/v2/clusters/{clusterId}/apps/batch-start` | 批量启动应用 |
| PUT | `/api/v2/clusters/{clusterId}/apps/batch-stop` | 批量停止应用 |
| POST | `/api/v2/clusters/{clusterId}/apps/batch-delete` | 批量删除应用 |

## 配置说明

系统通过 `APP_ENV` 选择配置文件：

| APP_ENV | 配置文件 |
| --- | --- |
| `dev` | `configs/config.dev.yaml` |
| `test` | `configs/config.test.yaml` |
| `prod` | `configs/config.prod.yaml` |

核心配置项：

```yaml
app:
  env: dev
  port: 8080

mysql:
  dsn: root:123456@tcp(127.0.0.1:3308)/compute_monitor_dev?charset=utf8mb4&parseTime=True&loc=Local
  max_open_conns: 20
  max_idle_conns: 10

redis:
  addr: 127.0.0.1:6379
  password: ""
  db: 0

auth:
  jwt:
    secret: dev-compute-monitor-secret
    issuer: compute-monitor-api
    access_token_ttl_seconds: 7200
    refresh_token_ttl_seconds: 604800

scheduler:
  k8s_sync:
    enabled: true
    interval_seconds: 60
    timeout_seconds: 30
    namespace: ""
```

## 本地运行

### 1. 准备依赖

需要本地可访问：

- Go 1.26.3
- MySQL 8.x
- Redis 7.x
- Kubernetes kubeconfig
- Prometheus HTTP API

### 2. 初始化数据库

创建数据库：

```sql
CREATE DATABASE compute_monitor_dev DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
```

系统启动时会自动执行 GORM AutoMigrate，补齐用户、集群、Kubernetes 缓存、指标、告警和审计相关表结构。

### 3. 配置环境

```powershell
$env:APP_ENV = "dev"
```

按实际环境修改：

```text
configs/config.dev.yaml
```

### 4. 启动服务

```powershell
D:\tools\environment\go1.26.3\bin\go.exe run .\cmd\server
```

服务启动后访问：

```text
http://localhost:8080/healthz
http://localhost:8080/swagger/index.html
```

## Docker 部署

### 构建镜像

```powershell
docker build -t compute-monitor-api:1.0.0 .
```

### 使用 docker-compose 启动

```powershell
docker compose -f deployments/docker-compose/docker-compose.yml up -d
```

启动后包含：

- `compute-monitor-api`：后端服务，端口 `8080`
- `mysql`：MySQL 数据库
- `redis`：Redis 会话缓存

## 认证流程

1. 调用 `/api/admin/auth/login`，提交用户名和密码。
2. 服务返回 access token、refresh token、过期时间和当前用户信息。
3. 后续请求在 Header 中携带：

```http
Authorization: Bearer <access_token>
```

4. access token 过期后调用 `/api/admin/auth/refresh` 刷新 token。
5. 用户修改密码后，旧 token 通过 `token_version` 机制失效。

## 分页规范

所有列表接口默认支持分页：

| 参数 | 默认值 | 说明 |
| --- | --- | --- |
| `page` | `1` | 页码 |
| `size` | `20` | 每页数量，最大 `100` |

服务内部需要全量查询时使用 `page.All()`，HTTP query 不能绕过分页限制。

## 迁移调度场景

系统面向算力迁移场景提供完整的数据闭环：

1. **资源发现**：同步集群 Namespace、Node、Pod、Deployment、Service。
2. **容量建模**：基于 Node capacity 计算 CPU、内存、GPU 静态容量。
3. **动态观测**：通过 Prometheus 查询实时 CPU、内存、磁盘、网络和 GPU 使用率。
4. **应用视图**：将 Deployment 和 Pod 映射为调度器可识别的应用与实例。
5. **迁移执行**：通过兼容接口执行 YAML 创建、Deployment 删除、批量启动、批量停止和批量删除。
6. **状态回写**：资源同步任务持续刷新数据库缓存，供前端和调度器查询。

## 开发规范

- `cmd/server` 只负责启动流程。
- `internal/app` 只负责装配，不写业务逻辑。
- `handler.go` 只处理 HTTP 参数和响应。
- `service.go` 负责业务校验、流程编排和跨仓储调用。
- `repository.go` 负责 GORM 数据访问。
- `dto.go` 定义请求体和响应体。
- `model.go` 定义数据库模型和领域模型。
- 新增或修改 Go 代码后运行：

```powershell
D:\tools\environment\go1.26.3\bin\gofmt.exe -w <files>
D:\tools\environment\go1.26.3\bin\go.exe test ./...
```

## 常用命令

```powershell
# 运行测试
D:\tools\environment\go1.26.3\bin\go.exe test ./...

# 生成 Swagger 文档
swag init -g cmd/server/main.go -o docs/swagger

# 本地启动
$env:APP_ENV = "dev"
D:\tools\environment\go1.26.3\bin\go.exe run .\cmd\server

# 构建镜像
docker build -t compute-monitor-api:1.0.0 .
```

## 许可证

本项目用于算力监控、资源调度与迁移决策平台建设，可按团队内部规范进行二次开发和部署。
