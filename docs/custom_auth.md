# 自定义认证（用户名+密码）实现文档

## 概述

在 MTProto 协议层新增了用户名+密码和手机号+密码的认证方式，复用原有的 MTProto 通道（端口 10443），不需要额外的端口或服务。

## 新增的 5 个 API 方法

| 方法 | CRC32 | 功能 |
|------|-------|------|
| `auth.getAuthMethods` | `0x3F83D1B` | 获取服务器支持的认证方式列表 |
| `auth.usernameRegister` | `0x96A42C40` | 用户名+密码注册 |
| `auth.usernameSignIn` | `0x6B8B4F99` | 用户名+密码登录 |
| `auth.phonePasswordRegister` | `0x4E5D6A32` | 手机号+密码注册 |
| `auth.phonePasswordSignIn` | `0x770CBD31` | 手机号+密码登录 |

## 请求流程图

```
iOS 客户端
    │  MTProto (TCP 10443)
    ▼
Gateway ──── 解析 CRC32，识别方法
    │  gRPC
    ▼
Session ──── 检查白名单（允许未登录调用）
    │  gRPC
    ▼
BFF Authorization Handler ──── 业务逻辑
    │  gRPC                         │  gRPC
    ▼                               ▼
UsernameService              UserPasswordService
(检查/设置用户名)             (存储/验证密码哈希)
    │  gRPC
    ▼
UserService
(创建/获取用户)
```

## 以 auth.usernameRegister 为例的完整调用链

```
1. iOS 发送 TLAuthUsernameRegister(username:"test", password:"123456", firstName:"Test")
2. Gateway 根据 CRC32 识别为 auth.usernameRegister
3. Session 白名单放行（未登录可调用）
4. BFF handler (auth.usernameRegister_handler.go):
   a. 验证用户名长度(3-32)和密码长度(6-128)
   b. UsernameClient.CheckUsername → 确认用户名未被占用
   c. bcrypt 哈希密码（BFF 端完成，不传明文给下游）
   d. UserClient.CreateNewUser → 创建用户，获得 userId
   e. UserPasswordClient.SaveUserPassword(userId, hash) → 写入 user_passwords 表
   f. UserClient.UpdateUsername + UsernameClient.UpdateUsername → 绑定用户名
   g. AuthsessionClient.BindAuthKeyUser → 绑定 auth_key 到用户
   h. 返回 Auth_Authorization(User) 给客户端
5. 客户端收到 Auth_Authorization，登录完成
```

---

## 新增/修改的文件清单

### 一、MTProto 协议层（proto/mtproto/）

这些文件定义了客户端和服务器之间的通信协议。

#### 新增文件

| 文件 | 作用 | 是否必须 |
|------|------|---------|
| `custom_auth.go` | 6 个新消息类型的 TL 编解码（Encode/Decode）+ init() 路由注册 | 必须。没有它服务器无法解析客户端发来的二进制数据 |
| `schema.tl.custom_auth.pb.go` | 6 个 protobuf 消息的 Go 结构体（由 schema.tl.custom_auth.proto 生成） | 必须。custom_auth.go 依赖这些结构体 |

#### 修改的文件

| 文件 | 改了什么 | 是否必须 |
|------|---------|---------|
| `schema.tl.crc32.pb.go` | 新增 6 个 CRC32 常量 | 必须。MTProto 靠 CRC32 识别每个方法 |
| `schema.tl.sync_service.pb.go` | RPCAuthorizationServer 接口新增 5 个方法 | 必须。gRPC 路由需要这些方法定义 |

#### proto 源文件（仅开发时需要，运行时不需要）

| 文件 | 位置 | 说明 |
|------|------|------|
| `schema.tl.custom_auth.proto` | proto_sources/ | 生成 schema.tl.custom_auth.pb.go 的源文件，已生成过，修改后需重新 protoc |

### 二、UserPasswordService 密码存储服务（app/service/biz/user/）

这是一个内部 gRPC 微服务，负责密码哈希的存取。和 UserService 运行在同一个进程（biz 服务）。

#### 新增文件

| 文件 | 层级 | 作用 | 是否必须 |
|------|------|------|---------|
| `proto/user_password.proto` | proto 源文件 | 定义 3 个 RPC 接口 | 开发时需要，运行时不需要 |
| `user_password/user_password.pb.go` | 生成代码 | Go 消息结构体 | 必须 |
| `user_password/user_password_grpc.pb.go` | 生成代码 | gRPC 客户端/服务端接口 | 必须 |
| `client/user_password_client.go` | 客户端封装 | BFF 调用密码服务的入口 | 必须 |
| `internal/server/grpc/service/user_password_service_impl.go` | gRPC 实现 | 接收 gRPC 请求，转发给 core | 必须 |
| `internal/core/user.password_handler.go` | 业务逻辑 | 调用 DAO 层 | 必须 |
| `internal/dao/password.go` | 数据访问 | SQL 操作 user_passwords 表 | 必须 |

#### 修改的文件

| 文件 | 改了什么 |
|------|---------|
| `helper.go` | 新增 `NewUserPasswordService()` 暴露给 biz 聚合服务 |
| `internal/server/grpc/grpc.go` | 注册 UserPasswordService 到 gRPC server |

#### 三个 RPC 接口

```protobuf
service UserPasswordService {
    rpc SaveUserPassword(userId, passwordHash) → {}
    rpc GetUserPassword(userId) → {passwordHash}
    rpc GetUserPasswordByPhone(phone) → {userId, passwordHash}
}
```

### 三、BFF Authorization（app/bff/authorization/）

BFF 是面向客户端的业务逻辑层，把 MTProto 请求翻译成对各个微服务的调用。

#### 新增文件

| 文件 | 作用 |
|------|------|
| `internal/core/auth.usernameRegister_handler.go` | 用户名注册逻辑 |
| `internal/core/auth.usernameSignIn_handler.go` | 用户名登录逻辑 |
| `internal/core/auth.phonePasswordRegister_handler.go` | 手机号+密码注册逻辑 |
| `internal/core/auth.phonePasswordSignIn_handler.go` | 手机号+密码登录逻辑 |
| `internal/core/auth.getAuthMethods_handler.go` | 获取支持的认证方式 |

#### 修改的文件

| 文件 | 改了什么 |
|------|---------|
| `internal/server/grpc/service/authorization_service_impl.go` | 新增 5 个 dispatch 方法 |
| `internal/dao/dao.go` | Dao 结构体嵌入 UserPasswordClient |

### 四、Session 白名单（app/interface/session/）

#### 修改的文件

| 文件 | 改了什么 |
|------|---------|
| `internal/service/check_api_request_type.go` | 5 个新方法加入未登录可调用白名单 |

### 五、Biz 聚合服务（app/service/biz/biz/）

#### 修改的文件

| 文件 | 改了什么 |
|------|---------|
| `internal/server/server.go` | 注册 UserPasswordService 到 gRPC server |

### 六、数据库

#### 修改的文件

| 文件 | 改了什么 |
|------|---------|
| `teamgramd/sql/1_teamgram.sql` | 末尾新增 user_passwords 表 |

### 七、iOS 客户端

| 文件 | 改了什么 |
|------|---------|
| `TelegramApi/Api23.swift` | 新增 `Api.auth.AuthMethods` 类型 |
| `TelegramApi/Api0.swift` | 注册 AuthMethods 解析器 |
| `TelegramApi/Api30.swift` | 新增 5 个 `Api.functions.auth.*` 方法 |
| `TelegramCore/Authorization.swift` | 新增登录/注册函数 |
| `AuthorizationUI/AuthorizationSequenceController.swift` | 自定义认证 UI 入口 |
| `CustomAuth/CustomAuthViewController.swift` | 用户名密码输入界面 |

---

## 密码安全

- 密码在 BFF 端用 **bcrypt** 哈希后再传给 UserPasswordService
- UserPasswordService 存储的是哈希值，不是明文
- debug 日志只打印 userId，不打印密码哈希
- 数据库表 `user_passwords` 存储 bcrypt hash

## 部署注意

1. `1_teamgram.sql` 已包含 `user_passwords` 建表语句，全新部署会自动创建
2. `go.mod` 中 `replace github.com/teamgram/proto => ./proto`，proto 在项目内部
3. Docker 构建时 proto 会被 `COPY . .` 自动复制进去
4. 所有 `.pb.go` 文件已生成好，不需要额外执行 protoc
