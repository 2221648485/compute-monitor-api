# 项目结构说明与 Java 转 Go 分层指南

本文档基于当前项目 `compute-monitor-api` 的实际代码编写，说明当前 Go 项目结构是否符合企业级规范，以及 Java Controller 中的代码迁移到 Go 后应该放在哪一层。

## 结论

当前项目已经采用了 Go 后端服务常见的企业化基础结构：

```text
cmd/
configs/
deployments/
docs/
internal/
pkg/
scripts/
test/
```

业务代码采用“按业务模块聚合，模块内部再分层”的方式：

```text
internal/student/
  dto.go
  handler.go
  model.go
  repository.go
  router.go
  service.go
```

这比 Java 常见的全局分层更适合 Go 项目：

```text
controller/
service/
mapper/
entity/
```

Go 项目里不推荐把所有 controller 放一个包、所有 service 放一个包。更推荐按业务模块组织，例如后续可以扩展：

```text
internal/node/
internal/cluster/
internal/metric/
internal/alert/
```

每个业务模块内部再拆 `handler / service / repository / dto / model / router`。

## 当前项目是否使用 GORM

当前项目已经改为使用 GORM。

数据库初始化位置：

```text
internal/store/mysql/mysql.go
```

核心依赖：

```go
gorm.io/gorm
gorm.io/driver/mysql
```

初始化方式：

```go
db, err := gorm.Open(mysql.Open(opts.DSN), &gorm.Config{})
```

同时通过 GORM 底层的 `sql.DB` 配置连接池：

```go
sqlDB, err := db.DB()
sqlDB.SetMaxOpenConns(opts.MaxOpenConns)
sqlDB.SetMaxIdleConns(opts.MaxIdleConns)
sqlDB.SetConnMaxLifetime(opts.ConnMaxLifetime)
```

这是一种比较常见的企业级写法：业务层使用 GORM，提高开发效率；连接池仍然由底层 `database/sql` 管理。

## 当前项目结构

```text
compute-monitor-api/
  cmd/
    server/
      main.go
  configs/
    config.yaml
  deployments/
  docs/
    development-standard.md
    project-structure-guide.md
  internal/
    app/
      app.go
      app_test.go
    config/
      config.go
    response/
      response.go
    store/
      mysql/
        mysql.go
    student/
      dto.go
      handler.go
      model.go
      repository.go
      router.go
      service.go
      service_test.go
  pkg/
  scripts/
  test/
  go.mod
  go.sum
```

## 目录职责

### `cmd/server`

程序入口，只负责启动流程。

当前文件：

```text
cmd/server/main.go
```

职责：

- 加载配置
- 初始化 MySQL/GORM
- 创建 HTTP Router
- 启动 HTTP 服务
- 关闭资源

`main.go` 不应该写业务逻辑，也不应该注册大量业务路由。现在项目已经把路由和模块装配抽到了 `internal/app`，这是比把所有依赖都堆在 `main.go` 里更规范的方式。

### `internal/app`

应用装配层，也可以理解为 bootstrap/composition root。

当前文件：

```text
internal/app/app.go
```

职责：

- 创建 Gin Engine
- 注册基础路由，例如 `/ping`
- 创建 `/api/v1` 路由组
- 组装业务模块依赖
- 注册业务模块路由

当前依赖组装在这里：

```go
studentRepository := student.NewMySQLRepository(db)
studentService := student.NewService(studentRepository)
studentHandler := student.NewHandler(studentService)
student.RegisterRoutes(api, studentHandler)
```

这个依赖方向是规范的：

```text
repository -> service -> handler -> router
```

更准确地说，应该理解为：

```text
handler 依赖 service
service 依赖 repository 接口
repository 依赖 database
router 绑定 handler
```

企业级 Go 项目里，手动依赖注入是很常见的。Go 没有 Spring 那种默认自动注入容器，所以显式 `NewXxx` 不是不规范。真正需要注意的是：不要把所有业务模块的装配代码一直堆在 `main.go`。当模块变多后，应该放到 `internal/app`、`internal/bootstrap` 或 `internal/server` 这类装配层。

### `internal/config`

配置模块。

当前文件：

```text
internal/config/config.go
```

职责：

- 读取环境变量
- 组织应用配置结构
- 给启动流程提供统一配置对象

后续如果要读取 `configs/config.yaml`，可以继续在这里扩展，不建议业务模块直接读取环境变量。

### `internal/store/mysql`

MySQL 基础设施模块。

当前文件：

```text
internal/store/mysql/mysql.go
```

职责：

- 初始化 GORM
- 配置底层连接池
- 关闭底层数据库连接

这里只处理“怎么连接数据库”，不处理具体业务表查询。

### `internal/response`

统一响应模块。

当前文件：

```text
internal/response/response.go
```

职责：

- 统一 JSON 响应结构
- 封装成功响应
- 封装常见错误响应

类似 Java 项目里的：

```text
Result<T>
ApiResponse<T>
R<T>
```

### `internal/student`

业务模块示例。

这个模块展示了后续新增业务模块时的标准写法。

## `student` 模块文件职责

### `router.go`

负责注册路由。

```go
func RegisterRoutes(router gin.IRouter, handler *Handler) {
    group := router.Group("/students")
    {
        group.POST("", handler.Create)
        group.GET("/:id", handler.GetByID)
    }
}
```

类似 Java Controller 上的：

```java
@RequestMapping("/students")
@PostMapping
@GetMapping("/{id}")
```

区别是 Gin 通常显式注册路由。

### `handler.go`

最接近 Java 的 Controller。

职责：

- 接收 HTTP 请求
- 解析 path/query/body 参数
- 调用 service
- 根据业务错误返回 HTTP 状态码
- 调用 `response` 输出统一 JSON

可以写：

```go
var req CreateStudentRequest
if err := c.ShouldBindJSON(&req); err != nil {
    response.BadRequest(c, err.Error())
    return
}
```

不应该写：

- SQL
- GORM 查询
- 复杂业务规则
- 事务编排
- 数据库连接初始化

### `service.go`

业务逻辑层。

职责：

- 数据清洗
- 业务校验
- 业务流程编排
- 调用 repository
- 把底层错误转换为业务错误

当前代码中，GORM 的未找到错误会在 service 层转换成业务错误：

```go
if errors.Is(err, gorm.ErrRecordNotFound) {
    return Student{}, ErrStudentNotFound
}
```

这比让 handler 直接识别数据库错误更好，因为 handler 只需要关心业务错误如何转成 HTTP 响应。

### `repository.go`

数据访问层。

职责：

- 定义 Repository 接口
- 使用 GORM 读写数据库
- 返回数据模型和错误

当前 GORM 写法：

```go
func (r *MySQLRepository) Create(ctx context.Context, student Student) (Student, error) {
    if err := r.db.WithContext(ctx).Create(&student).Error; err != nil {
        return Student{}, err
    }

    return student, nil
}
```

查询写法：

```go
func (r *MySQLRepository) FindByID(ctx context.Context, id int64) (Student, error) {
    var student Student
    if err := r.db.WithContext(ctx).First(&student, id).Error; err != nil {
        return Student{}, err
    }

    return student, nil
}
```

repository 不应该处理 HTTP 状态码，也不应该直接返回前端 DTO。

### `model.go`

领域模型或数据库模型。

当前：

```go
type Student struct {
    ID        int64
    Name      string
    Age       int
    Email     string
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

GORM 会按默认规则把 `Student` 映射到 `students` 表，并识别 `ID`、`CreatedAt`、`UpdatedAt`。

如果后续表名或字段名不符合默认规则，可以增加：

```go
func (Student) TableName() string {
    return "students"
}
```

或使用 GORM tag。

### `dto.go`

请求和响应对象。

职责：

- 定义请求参数
- 定义响应结构
- 使用 `binding` tag 做基础校验
- 把 model 转换成 response DTO

不要直接用数据库 model 接收前端请求。

推荐流程：

```text
CreateStudentRequest -> Student -> StudentResponse
```

## Java Controller 中的代码应该放哪里

| Java Controller 里的代码 | Go 中的位置 |
| --- | --- |
| `@RequestBody` 接收 JSON | `handler.go` |
| `@PathVariable` 路径参数 | `handler.go` |
| `@RequestParam` 查询参数 | `handler.go` |
| 返回 HTTP 状态码 | `handler.go` |
| 返回统一响应 | `handler.go` 调用 `internal/response` |
| 请求对象 | `dto.go` |
| 响应对象 / VO | `dto.go` |
| 简单字段校验 | `dto.go` 的 `binding` tag |
| 复杂业务校验 | `service.go` |
| 数据清洗 | `service.go` |
| 事务编排 | `service.go` |
| 调用多个 repository | `service.go` |
| GORM 查询 | `repository.go` |
| 数据库模型 | `model.go` |
| 路由映射 | `router.go` |

一句话：

```text
Go 里的 handler 像 Java Controller，但只保留 HTTP 相关代码。
Controller 里原来顺手写的业务逻辑，放到 service。
Controller 里原来顺手写的 SQL 或 ORM 查询，放到 repository。
```

## Java 和当前 Go 结构对应关系

| Java/Spring 概念 | 当前 Go 项目位置 | 说明 |
| --- | --- | --- |
| `Application` 启动类 | `cmd/server/main.go` | 程序入口 |
| Spring 容器装配 | `internal/app/app.go` | 手动依赖注入和路由装配 |
| `@RestController` | `internal/<module>/handler.go` | HTTP 控制器 |
| `@RequestMapping` | `internal/<module>/router.go` | 路由注册 |
| `Service` | `internal/<module>/service.go` | 业务逻辑 |
| `Mapper` / `DAO` / `Repository` | `internal/<module>/repository.go` | GORM 数据访问 |
| `Entity` / `Domain` | `internal/<module>/model.go` | 领域模型或数据库模型 |
| `Request DTO` | `internal/<module>/dto.go` | 请求参数 |
| `VO` / `Response DTO` | `internal/<module>/dto.go` | 响应对象 |
| `Result<T>` | `internal/response/response.go` | 统一响应 |
| `application.yml` | `configs/config.yaml` + `internal/config` | 配置文件和配置读取 |
| `DataSourceConfig` | `internal/store/mysql/mysql.go` | GORM 和连接池初始化 |

## 关于手动 new 是否规范

这段代码本身是规范的：

```go
studentRepository := student.NewMySQLRepository(db)
studentService := student.NewService(studentRepository)
studentHandler := student.NewHandler(studentService)
student.RegisterRoutes(api, studentHandler)
```

它体现的是依赖注入，而不是在业务代码里到处 new。

不推荐的是下面这些写法：

```go
func (h *Handler) Create(c *gin.Context) {
    db := mysql.New(...)
    repo := student.NewMySQLRepository(db)
    service := student.NewService(repo)
    ...
}
```

也不推荐在 service 里自己创建 repository：

```go
func NewService() *Service {
    db := mysql.New(...)
    return &Service{
        repository: NewMySQLRepository(db),
    }
}
```

企业级 Go 更推荐：

```text
启动层统一创建依赖
业务层通过构造函数接收依赖
业务代码不关心依赖如何创建
```

当前项目已经把这个装配逻辑从 `main.go` 移到了 `internal/app`，后续模块变多时可以继续拆成：

```text
internal/app/app.go
internal/app/modules.go
internal/app/middleware.go
```

或者：

```text
internal/server/server.go
internal/bootstrap/bootstrap.go
```

不需要一开始就引入复杂 DI 框架。

## 新增业务模块规范

假设新增算力节点模块 `node`：

```text
internal/node/
  dto.go
  handler.go
  model.go
  repository.go
  router.go
  service.go
```

在 `internal/app` 中装配：

```go
nodeRepository := node.NewMySQLRepository(db)
nodeService := node.NewService(nodeRepository)
nodeHandler := node.NewHandler(nodeService)
node.RegisterRoutes(api, nodeHandler)
```

当模块越来越多时，可以拆出：

```go
func registerStudentModule(api gin.IRouter, db *gorm.DB) {
    studentRepository := student.NewMySQLRepository(db)
    studentService := student.NewService(studentRepository)
    studentHandler := student.NewHandler(studentService)
    student.RegisterRoutes(api, studentHandler)
}
```

这样 `registerModules` 不会变得太长。

## 推荐请求处理流程

```text
HTTP 请求
  -> router.go 匹配路由
  -> handler.go 解析参数
  -> dto.go 承载请求结构和基础校验
  -> service.go 执行业务逻辑
  -> repository.go 使用 GORM 访问数据库
  -> service.go 转换业务错误
  -> handler.go 转换 HTTP 响应
  -> response.go 输出统一 JSON
```

## 后续企业级演进建议

当前结构可以继续补充：

- `internal/middleware`：鉴权、CORS、请求日志、trace id
- `internal/apperror`：统一业务错误码
- `internal/logger`：日志组件
- `internal/validator`：复杂参数校验
- `internal/store/redis`：Redis 初始化
- `deployments/migrations`：数据库迁移脚本
- `internal/<module>/*_test.go`：单元测试
- `test`：集成测试和测试数据

不要过早引入太多目录。先保持：

```text
cmd/server -> internal/app -> internal/<module> -> internal/store
```

这个主线清晰即可。

