# 阶段 0-1 API 开发文档

项目后端：`compute-monitor-api`

## 配置读取

当前项目使用 Viper 读取分环境 YAML：

```text
configs/config.dev.yaml
configs/config.test.yaml
configs/config.prod.yaml
```

通过 `APP_ENV` 选择配置：

```text
APP_ENV=dev  -> configs/config.dev.yaml
APP_ENV=test -> configs/config.test.yaml
APP_ENV=prod -> configs/config.prod.yaml
```

不设置 `APP_ENV` 时默认使用 `dev`。

Dockerfile 默认设置 `APP_ENV=prod`，所以镜像默认使用生产配置。

## 健康检查

```http
GET /healthz
```

示例：

```bash
curl http://localhost:8080/healthz
```

## API 前缀

正式接口统一使用：

```text
/api/v2
```

## 后续替换方向

后续从 mock 数据切换到真实数据时，按模块逐步替换：

1. 集群列表从本地配置或 MySQL 读取。
2. 节点数据从 Kubernetes `client-go` 读取。
3. 工作负载数据从 Kubernetes Deployment/Pod 读取。
4. 指标数据从 Prometheus 查询。

