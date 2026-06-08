# 配置文件使用说明

## 推荐方案

当前项目不使用 `.env` 文件。

环境配置直接写在 YAML 中：

- `configs/config.dev.yaml`：本地开发环境
- `configs/config.test.yaml`：测试环境
- `configs/config.prod.yaml`：正式运行环境

程序通过 `APP_ENV` 选择读取哪个配置文件：

```text
APP_ENV=dev  -> configs/config.dev.yaml
APP_ENV=test -> configs/config.test.yaml
APP_ENV=prod -> configs/config.prod.yaml
```

如果不设置 `APP_ENV`，默认使用 `dev`。

## 本地开发

本地直接运行时，不设置 `APP_ENV` 就会读取：

```text
configs/config.dev.yaml
```

也可以显式设置：

```powershell
$env:APP_ENV="dev"
```

## Docker 镜像

Dockerfile 中已经设置：

```dockerfile
ENV APP_ENV=prod
```

所以镜像启动时默认读取：

```text
configs/config.prod.yaml
```

不需要打包前手动修改配置文件。

构建镜像：

```powershell
docker build -t compute-monitor-api:1.0.0 .
```

运行镜像：

```powershell
docker run -p 8080:8080 compute-monitor-api:1.0.0
```

如果临时想用镜像跑 dev 配置，可以运行时覆盖：

```powershell
docker run -e APP_ENV=dev -p 8080:8080 compute-monitor-api:1.0.0
```

## Docker Compose

`deployments/docker-compose/docker-compose.yml` 是本地联调用的，所以显式设置：

```yaml
APP_ENV: dev
```

它会读取 `configs/config.dev.yaml`，并启动一个本地 MySQL 容器。

启动：

```powershell
docker compose -f deployments/docker-compose/docker-compose.yml up -d
```

## Kubernetes

Kubernetes 中推荐同一个镜像部署到不同环境，只通过 `APP_ENV` 选择配置：

```yaml
env:
  - name: APP_ENV
    value: prod
```

如果你决定把生产数据库连接也写死在 `configs/config.prod.yaml`，那么 K8s 不需要额外注入 `MYSQL_DSN`。

更安全的企业级做法是：`config.prod.yaml` 只放非敏感默认值，数据库密码通过 K8s Secret 注入。但如果你的目标是先跑通项目，直接写在 `config.prod.yaml` 也可以。

