# Auth 与 User 模块设计说明

本文档说明后台登录认证和后台用户管理的目录结构、接口设计、分层职责，以及 `auth` 和 `user` 两个模块之间的关系。

## 结论

建议拆成两个业务模块：

```text
internal/auth
internal/user
```

职责划分：

```text
auth:
  负责登录认证、token 签发、token 解析、退出登录、当前登录态。

user:
  负责后台用户管理，例如用户列表、创建用户、更新用户、启用禁用用户、重置密码。
```

`User` 模型应该归 `internal/user` 所有。  
`auth` 登录时只复用 `user.User` 和 `user.Repository`，不要在 `auth` 中再维护另一份用户模型。

## 推荐目录结构

最终推荐结构如下：

```text
internal/
  auth/
    dto.go
    errors.go
    handler.go
    password.go
    router.go
    service.go
    token.go

  user/
    dto.go
    errors.go
    handler.go
    model.go
    repository.go
    router.go
    service.go

  middleware/
    auth.go
    request_log.go

  app/
    app.go

  response/
    response.go

  store/
    mysql/
      mysql.go
```

当前项目中 `internal/auth/model.go` 还存在 `User` 和 `LoginLog`。  
后续应该把它们迁移到 `internal/user/model.go`，然后让 `auth` 引用 `user` 模块。

## Auth 模块职责

`auth` 只处理认证相关能力。

推荐文件职责：

```text
internal/auth/dto.go
  LoginRequest
  LoginResponse
  RefreshTokenRequest
  ChangePasswordRequest

internal/auth/errors.go
  ErrInvalidCredential
  ErrTokenInvalid
  ErrTokenExpired
  ErrPermissionDenied
  IsAuthError
  ErrorMessage

internal/auth/password.go
  PasswordHasher
  BcryptPasswordHasher
  NewPasswordHasher
  HashPassword
  ComparePassword

internal/auth/token.go
  TokenOptions
  TokenClaims
  TokenManager
  JWTManager
  NewTokenManager
  Generate
  Parse

internal/auth/service.go
  Service
  NewService
  Login
  GetCurrentUser
  Logout
  ChangePassword

internal/auth/handler.go
  Handler
  NewHandler
  Login
  Me
  Logout
  ChangePassword

internal/auth/router.go
  RegisterRoutes
  RegisterPublicRoutes
  RegisterPrivateRoutes
```

### Auth 接口

后台认证接口建议如下：

```text
POST /api/admin/auth/login
GET  /api/admin/auth/me
POST /api/admin/auth/logout
POST /api/admin/auth/password
```

接口说明：

| 方法 | 路径 | 是否需要登录 | Handler 方法 | 说明 |
| --- | --- | --- | --- | --- |
| POST | `/api/admin/auth/login` | 否 | `Login` | 登录并返回 token |
| GET | `/api/admin/auth/me` | 是 | `Me` | 获取当前登录用户 |
| POST | `/api/admin/auth/logout` | 是 | `Logout` | 退出登录 |
| POST | `/api/admin/auth/password` | 是 | `ChangePassword` | 修改当前用户密码 |

### Login 请求

```json
{
  "username": "admin",
  "password": "12345678"
}
```

说明：

- 前端提交密码时可以是明文，但生产环境必须使用 HTTPS。
- 后端数据库只保存 `password_hash`。
- 后端用 bcrypt 校验前端提交的密码和数据库中的哈希是否匹配。

### Login 响应

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "access_token": "jwt-token",
    "token_type": "Bearer",
    "expires_in": 7200,
    "user": {
      "id": 1,
      "username": "admin",
      "nickname": "管理员",
      "role": "admin"
    }
  }
}
```

## User 模块职责

`user` 负责后台用户管理。

推荐文件职责：

```text
internal/user/model.go
  User
  LoginLog
  RoleAdmin
  RoleOperator
  RoleViewer
  UserStatusEnabled
  UserStatusDisabled

internal/user/dto.go
  CreateUserRequest
  UpdateUserRequest
  UpdateUserStatusRequest
  UserListRequest
  UserResponse
  UserListResponse
  ToResponse
  ToResponses

internal/user/errors.go
  ErrUserNotFound
  ErrUsernameExists
  ErrCannotDisableSelf
  ErrInvalidUserStatus
  ErrInvalidUserRole
  ErrPasswordTooWeak

internal/user/repository.go
  Repository
  MySQLRepository
  NewMySQLRepository
  Create
  FindByID
  FindByUsername
  List
  Count
  Update
  UpdateStatus
  UpdatePassword
  UpdateLastLoginAt
  CreateLoginLog

internal/user/service.go
  Service
  NewService
  Create
  GetByID
  List
  Update
  UpdateStatus
  ResetPassword

internal/user/handler.go
  Handler
  NewHandler
  Create
  List
  GetByID
  Update
  UpdateStatus
  ResetPassword

internal/user/router.go
  RegisterRoutes
```

## User 管理接口

你当前提到的用户管理接口建议如下：

```text
GET  /api/admin/users
POST /api/admin/users
GET  /api/admin/users/{userId}
PUT  /api/admin/users/{userId}
PUT  /api/admin/users/{userId}/status
```

可以再补一个重置密码接口：

```text
PUT /api/admin/users/{userId}/password
```

接口表：

| 方法 | 路径 | 是否需要登录 | 建议 Handler 方法 | 说明 |
| --- | --- | --- | --- | --- |
| GET | `/api/admin/users` | 是 | `List` | 分页查询后台用户 |
| POST | `/api/admin/users` | 是 | `Create` | 创建后台用户 |
| GET | `/api/admin/users/{userId}` | 是 | `GetByID` | 查看用户详情 |
| PUT | `/api/admin/users/{userId}` | 是 | `Update` | 更新用户基础信息 |
| PUT | `/api/admin/users/{userId}/status` | 是 | `UpdateStatus` | 启用或禁用用户 |
| PUT | `/api/admin/users/{userId}/password` | 是 | `ResetPassword` | 重置用户密码 |

## 请求和响应示例

### 查询用户列表

```http
GET /api/admin/users?page=1&page_size=20&keyword=admin&status=enabled
Authorization: Bearer <token>
```

响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "username": "admin",
        "nickname": "管理员",
        "role": "admin",
        "status": "enabled"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 20
  }
}
```

### 创建用户

```http
POST /api/admin/users
Authorization: Bearer <token>
Content-Type: application/json
```

```json
{
  "username": "operator01",
  "password": "12345678",
  "nickname": "操作员 01",
  "role": "operator"
}
```

### 更新用户

```http
PUT /api/admin/users/2
Authorization: Bearer <token>
Content-Type: application/json
```

```json
{
  "nickname": "新的昵称",
  "role": "viewer"
}
```

### 更新用户状态

```http
PUT /api/admin/users/2/status
Authorization: Bearer <token>
Content-Type: application/json
```

```json
{
  "status": "disabled"
}
```

## Auth 与 User 的依赖关系

推荐依赖方向：

```text
auth -> user
```

原因：

- 登录认证需要查询用户。
- token 中需要写入用户 ID、用户名、角色。
- 当前用户接口需要通过用户 ID 查询用户详情。

不推荐：

```text
user -> auth
```

因为用户管理不应该依赖 HTTP 登录认证逻辑。

如果用户管理创建用户或重置密码需要密码哈希，可以有两个选择：

### 方案一：暂时复用 auth.PasswordHasher

```text
user service -> auth.PasswordHasher
```

优点：

- 当前阶段简单直接。

缺点：

- `user` 会依赖 `auth`，边界不够干净。

### 方案二：抽出 security 包

```text
internal/security/password.go
```

推荐结构：

```text
internal/security/
  password.go
```

然后：

```text
auth -> security
user -> security
```

优点：

- `auth` 和 `user` 都不互相依赖密码实现。
- 后续扩展加密、签名、密码策略更清晰。

当前阶段可以先用方案一，等 `user` 真正实现创建用户和重置密码时，再考虑抽 `internal/security`。

## 推荐迁移顺序

当前项目中 `auth` 已有完整代码，`user` 还是注释骨架。后续迁移建议：

1. 在 `internal/user/model.go` 中实现 `User`、`LoginLog`、角色常量、状态常量。
2. 从 `internal/auth/model.go` 删除 `User`、`LoginLog`。
3. 修改 `auth/token.go`，把参数类型从 `auth.User` 改成 `user.User`。
4. 修改 `auth/service.go`，让它依赖用户查询接口。
5. 在 `internal/user/repository.go` 中实现 `FindByUsername`、`FindByID`、`UpdateLastLoginAt`、`CreateLoginLog`。
6. 在 `internal/app/app.go` 中把 auth repository 替换为 user repository。
7. 实现 `internal/user` 的管理接口。
8. 注册 `user.RegisterRoutes(privateAPI, userHandler)`。

最终目标：

```text
auth:
  只负责认证流程。

user:
  拥有 User 模型和用户管理能力。
```

## App 接入方式

`internal/app/app.go` 建议保持依赖装配职责：

```text
config -> token manager
db -> user repository
repository -> auth service
service -> auth handler
handler -> auth router
```

后台接口分两类：

```text
publicAPI:
  不需要登录，例如 /api/admin/auth/login

privateAPI:
  需要登录，挂 middleware.Auth(tokenManager)
```

后续用户管理路由应该挂到 `privateAPI`：

```text
user.RegisterRoutes(privateAPI, userHandler)
```

高权限接口再继续叠加角色中间件：

```text
RequireRole("admin")
```

