# 项目目录结构与 API 模块映射

## 当前结构

```text
compute-monitor-api/
  cmd/server/main.go
  configs/
    config.dev.yaml
    config.test.yaml
    config.prod.yaml
  deployments/
    docker-compose/docker-compose.yml
  internal/
    app/
    config/
    middleware/
    response/
    store/mysql/
    student/
  docs/
  scripts/
```

## 模块职责

| 目录 | 职责 |
| --- | --- |
| `cmd/server` | 程序入口，加载配置、初始化依赖、启动服务 |
| `internal/app` | Gin Engine、路由分组、模块装配 |
| `internal/config` | 按 `APP_ENV` 读取分环境 YAML 配置 |
| `internal/store/mysql` | GORM 和 MySQL 连接池初始化 |
| `internal/response` | 统一响应结构 |
| `internal/middleware` | 请求日志、鉴权、CORS 等中间件 |
| `internal/student` | Go 分层示例模块 |

## 配置读取

当前项目不使用 `.env` 文件，也不通过 `MYSQL_DSN` 等环境变量覆盖配置。

程序只使用 `APP_ENV` 选择配置文件：

```text
APP_ENV=dev  -> configs/config.dev.yaml
APP_ENV=test -> configs/config.test.yaml
APP_ENV=prod -> configs/config.prod.yaml
```

如果不设置 `APP_ENV`，默认读取 `configs/config.dev.yaml`。

Docker 镜像中默认设置：

```text
APP_ENV=prod
```

所以镜像默认读取 `configs/config.prod.yaml`。

