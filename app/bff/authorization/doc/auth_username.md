# 自定义认证服务

## 概述

在原有手机号+验证码认证的基础上，新增三种认证方式，通过配置控制启用哪些方式：

| 认证方式 | 配置值 | 说明 |
|---------|--------|------|
| 用户名+密码 | `username_password` | 用户名注册登录 |
| 手机号+验证码 | `phone_sms_code` | 原有 MTProto 认证流程 |
| 手机号+密码 | `phone_password` | 手机号注册登录（密码认证） |

认证成功后返回 MTProto `auth_key`，客户端可直接使用该密钥建立 MTProto 连接，无需维护额外的网络协议。

## 架构

```
客户端 --> gRPC --> AuthUsernameService (authorization BFF)
                        |
                        ├── GetAuthMethods          → 读取 Config.AuthMethods
                        ├── CheckUsernameAvailable   → UsernameClient
                        ├── UsernameRegister/SignIn   → UsernameClient + UserClient + UserPasswordClient
                        ├── PhonePasswordRegister/SignIn → UserClient + UserPasswordClient
                        |
                        ├── UserPasswordClient ──RPC──> UserPasswordService (user 服务)
                        │                                   ├── SaveUserPassword
                        │                                   ├── GetUserPassword (by user_id)
                        │                                   └── GetUserPasswordByPhone (JOIN users + user_passwords)
                        ├── AuthsessionClient (创建 auth_key)
                        └── bcrypt (密码哈希/验证，在 BFF 层完成)
```

密码存取通过 RPC 调用 user 服务的 `UserPasswordService`，遵循微服务分层架构（BFF 不直接访问数据库）。

## API 接口

Proto 文件位置：`app/bff/authorization/proto/auth_username.proto`

### 0. GetAuthMethods - 获取支持的认证方式

**请求：** 无参数

**响应：**
| 字段 | 类型 | 说明 |
|------|------|------|
| auth_methods | repeated string | 认证方式列表 |

**示例返回：** `["username_password", "phone_sms_code", "phone_password"]`

客户端应先调用此接口，根据返回动态展示可用的登录/注册选项。

### 1. CheckUsernameAvailable - 检查用户名是否可用

**请求：**
| 字段 | 类型 | 说明 |
|------|------|------|
| username | string | 用户名 |

**响应：**
| 字段 | 类型 | 说明 |
|------|------|------|
| available | bool | 是否可用 |

**规则：** 用户名长度 3-32 字符。

### 2. UsernameRegister - 用户名+密码注册

**请求：**
| 字段 | 类型 | 说明 |
|------|------|------|
| username | string | 用户名（3-32 字符） |
| password | string | 密码（6-128 字符） |
| first_name | string | 名字 |
| last_name | string | 姓氏 |

**响应：** `AuthResp`（见下方）

**流程：**
1. 验证 username 和 password 长度
2. 检查用户名是否已被占用
3. bcrypt 哈希密码（BFF 层）
4. 创建用户（UserClient.UserCreateNewUser，phone 为空）
5. 保存密码哈希（UserPasswordClient.SaveUserPassword → user 服务）
6. 设置用户名（UserClient + UsernameClient）
7. 创建 MTProto auth_key 并绑定 user_id

### 3. UsernameSignIn - 用户名+密码登录

**请求：**
| 字段 | 类型 | 说明 |
|------|------|------|
| username | string | 用户名 |
| password | string | 密码 |

**响应：** `AuthResp`（见下方）

**流程：**
1. 通过用户名解析 user_id（UsernameClient.UsernameResolveUsername）
2. 获取密码哈希（UserPasswordClient.GetUserPassword → user 服务）
3. bcrypt 验证密码（BFF 层）
4. 获取用户信息（UserClient.UserGetImmutableUser）
5. 创建 MTProto auth_key 并绑定 user_id

### 4. PhonePasswordRegister - 手机号+密码注册

**请求：**
| 字段 | 类型 | 说明 |
|------|------|------|
| phone | string | 手机号 |
| password | string | 密码（6-128 字符） |
| first_name | string | 名字 |
| last_name | string | 姓氏 |

**响应：** `AuthResp`（见下方）

**流程：**
1. 验证 phone 和 password
2. bcrypt 哈希密码（BFF 层）
3. 创建用户（UserClient.UserCreateNewUser，带手机号）
4. 保存密码哈希（UserPasswordClient.SaveUserPassword → user 服务）
5. 创建 MTProto auth_key 并绑定 user_id

### 5. PhonePasswordSignIn - 手机号+密码登录

**请求：**
| 字段 | 类型 | 说明 |
|------|------|------|
| phone | string | 手机号 |
| password | string | 密码 |

**响应：** `AuthResp`（见下方）

**流程：**
1. 通过手机号获取 user_id + 密码哈希（UserPasswordClient.GetUserPasswordByPhone → user 服务，一次 RPC，JOIN 查询）
2. bcrypt 验证密码（BFF 层）
3. 获取用户信息（UserClient.UserGetImmutableUser）
4. 创建 MTProto auth_key 并绑定 user_id

### AuthResp - 统一认证响应（所有认证方式共用）

| 字段 | 类型 | 说明 |
|------|------|------|
| user_id | int64 | 用户 ID |
| username | string | 用户名（用户名认证时返回） |
| phone | string | 手机号（手机号认证时返回） |
| first_name | string | 名字 |
| last_name | string | 姓氏 |
| auth_key | bytes | MTProto 授权密钥（256 字节） |
| auth_key_id | int64 | 授权密钥 ID |

## 内部 RPC：UserPasswordService

Proto 文件位置：`app/service/biz/user/proto/user_password.proto`

运行在 user 服务进程中，供 authorization BFF 通过 RPC 调用。

| 方法 | 说明 |
|------|------|
| SaveUserPassword | 保存 user_id → password_hash |
| GetUserPassword | 按 user_id 查 password_hash |
| GetUserPasswordByPhone | 按 phone JOIN 查 user_id + password_hash |

## 数据库

`user_passwords` 表定义在 `teamgramd/sql/1_teamgram.sql` 中：

```sql
CREATE TABLE `user_passwords` (
  `id` bigint(20) NOT NULL,
  `user_id` bigint(20) NOT NULL COMMENT '用户ID',
  `password_hash` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '密码哈希（bcrypt）',
  `deleted` tinyint(1) NOT NULL DEFAULT '0',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

如果数据库已初始化，需手动执行建表语句。

## 文件结构

```
app/bff/authorization/                     # BFF 层（认证入口）
├── proto/
│   └── auth_username.proto                # Proto 源文件
├── auth_username/
│   ├── auth_username.pb.go                # protoc 生成
│   └── auth_username_grpc.pb.go           # protoc 生成
├── doc/
│   └── auth_username.md                   # 本文档
└── internal/
    ├── config/config.go                   # AuthMethods 配置 + 常量
    ├── dao/dao.go                         # 嵌入 UserPasswordClient
    ├── core/
    │   └── auth_username_handler.go       # 业务逻辑（5 个认证方法 + createAuthKeyAndBind 公共方法）
    └── server/grpc/
        ├── grpc.go                        # 注册 AuthUsernameService
        └── service/
            └── auth_username_service_impl.go

app/service/biz/user/                      # User 服务（密码存取）
├── proto/
│   └── user_password.proto                # Proto 源文件
├── user_password/
│   ├── user_password.pb.go                # protoc 生成
│   └── user_password_grpc.pb.go           # protoc 生成
├── client/
│   └── user_password_client.go            # UserPasswordClient 接口（3 个方法）
└── internal/
    ├── dao/password.go                    # 密码存取 DAO（含 JOIN 查询）
    ├── core/
    │   └── user.password_handler.go       # Core handler
    └── server/grpc/
        ├── grpc.go                        # 注册 UserPasswordService
        └── service/
            └── user_password_service_impl.go
```

## 部署配置

### 认证方式配置

在 `authorization.yaml` 中配置启用的认证方式（可选，默认 `["username_password", "phone_sms_code"]`）：

```yaml
AuthMethods:
  - username_password
  - phone_sms_code
  - phone_password
```

### 其他

- authorization BFF 不需要数据库配置，密码数据通过 user 服务 RPC 访问
- UserPasswordClient 复用 UserClient 的 RPC 连接（同一个 user 服务地址）
- user 服务本身已有 MySQL 连接，会复用 `user_passwords` 表

## 依赖

需要添加 `golang.org/x/crypto` 依赖（bcrypt）：

```bash
cd teamgram-server
go mod tidy
```

## 重新生成 Proto

如果修改了 proto 文件，使用以下命令重新生成 Go 代码：

```bash
# authorization BFF 的 auth_username proto
cd app/bff/authorization
protoc --go_out=./auth_username --go_opt=paths=source_relative \
       --go-grpc_out=./auth_username --go-grpc_opt=paths=source_relative \
       proto/auth_username.proto
# 注意：生成后需将 auth_username/proto/*.go 移到 auth_username/ 下

# user 服务的 user_password proto
cd app/service/biz/user
protoc --go_out=./user_password --go_opt=paths=source_relative \
       --go-grpc_out=./user_password --go-grpc_opt=paths=source_relative \
       proto/user_password.proto
# 注意：生成后需将 user_password/proto/*.go 移到 user_password/ 下
```

需要安装的工具（注意版本，项目使用 Go 1.19 + grpc v1.59.0）：
- `protoc-gen-go@v1.31.0`
- `protoc-gen-go-grpc@v1.3.0`
