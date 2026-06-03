# Go Gin 企业级开发规范示例

这个项目中的 `student` 模块演示了一个后端接口常见的企业级分层写法。

## 分层职责

- `cmd/server`: 程序入口，只负责加载配置、初始化依赖、注册路由、启动服务。
- `internal/config`: 配置读取，环境变量、配置文件都应集中在这里处理。
- `internal/store`: 数据库、缓存等基础设施初始化。
- `internal/student/model.go`: 数据库领域模型，表示业务实体。
- `internal/student/dto.go`: HTTP 请求和响应结构，避免把数据库模型直接暴露给前端。
- `internal/student/repository.go`: 数据访问层，只写 SQL 和数据库读写逻辑。
- `internal/student/service.go`: 业务层，处理参数清洗、业务规则、事务编排。
- `internal/student/handler.go`: 控制器层，处理 HTTP 参数、状态码和响应。
- `internal/student/router.go`: 路由注册，避免把路由散落在入口文件里。
- `internal/response`: 统一响应格式。

## 接口示例

创建学生：

```http
POST /api/v1/students
Content-Type: application/json

{
  "name": "Zhang San",
  "age": 20,
  "email": "zhangsan@example.com"
}
```

查询学生：

```http
GET /api/v1/students/1
```

## MySQL 表结构参考

当前代码只是规范示例，不要求你现在已经建表。后续真的连接 MySQL 时，可以参考：

```sql
CREATE TABLE students (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(64) NOT NULL,
  age INT NOT NULL,
  email VARCHAR(128) NOT NULL UNIQUE,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
);
```

## 推荐规范

- 控制器不写 SQL，不写复杂业务逻辑。
- service 不依赖 Gin，方便单元测试。
- repository 不处理 HTTP 状态码，只返回数据和错误。
- 请求参数使用 DTO，并用 `binding` 做基础校验。
- 响应格式统一，前端更容易处理。
- 依赖从入口处组装，避免在业务代码里到处 new 数据库连接。
- 包名简短小写，文件名按职责命名。
- SQL 参数使用占位符，避免 SQL 注入。
