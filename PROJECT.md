# Teamgram Server 技术文档

> Telegram MTProto 协议的开源 Go 服务端实现（API Layer 166），兼容官方 Telegram 客户端。
> 版本：v0.96.0-teamgram-server | 协议：Apache 2.0

---

## 目录

- [技术栈](#技术栈)
- [微服务架构](#微服务架构)
- [服务启动顺序与端口](#服务启动顺序与端口)
- [请求流转路径](#请求流转路径)
- [Gateway 详解](#gateway-详解)
- [Session 详解](#session-详解)
- [BFF 详解](#bff-详解)
- [Messenger 详解（msg / sync）](#messenger-详解msg--sync)
- [Service 层详解](#service-层详解)
- [Kafka 消息流](#kafka-消息流)
- [目录结构](#目录结构)
- [代码组织规范](#代码组织规范)
- [数据库设计](#数据库设计)
- [配置文件详解](#配置文件详解)
- [Docker 部署](#docker-部署)
- [关键外部依赖](#关键外部依赖)
- [公共工具包 pkg/](#公共工具包-pkg)
- [核心业务流程](#核心业务流程)
- [开发备忘](#开发备忘)

---

## 技术栈

| 组件         | 技术                                                           |
|:-------------|:---------------------------------------------------------------|
| 语言         | Go 1.19                                                       |
| 框架         | go-zero v1.6.0（teamgram fork 版）                             |
| 数据库       | MySQL 5.7（InnoDB, utf8mb4）                                   |
| 缓存         | Redis（AOF 持久化）                                            |
| 消息队列     | Apache Kafka（通过 Zookeeper）                                  |
| 服务发现     | etcd v3.5.2                                                    |
| 对象存储     | MinIO（API 9000, Console 9001）                                |
| 媒体处理     | FFmpeg                                                         |
| RPC          | gRPC v1.59.0 + Protocol Buffers                               |
| 序列化       | Protobuf (gogo) + 自定义 TL Schema（mtprotoc 代码生成器）       |
| 可观测性     | OpenTelemetry（Jaeger/Zipkin/OTLP）、Prometheus                 |
| 容器化       | Docker + Docker Compose                                        |
| ID 生成      | Snowflake                                                      |
| 构建标签     | `-tags=jsoniter`（使用 jsoniter 替代标准 JSON）                 |

---

## 微服务架构

共 **11 个独立微服务**，分为 4 层：

```
┌──────────────────────────────────────────────────────────────┐
│                       Interface 层                            │
│  gateway (TCP 网关, 3 端口)  ──gRPC──→  session (会话/路由)   │
├──────────────────────────────────────────────────────────────┤
│                         BFF 层                                │
│  bff (聚合 22 个 gRPC 服务: account, auth, messages, ...)    │
├──────────────────────────────────────────────────────────────┤
│                      Messenger 层                             │
│  msg (消息发送/收件箱, Kafka 生产+消费)                        │
│  sync (实时同步推送, Kafka 消费, 协程池)                       │
├──────────────────────────────────────────────────────────────┤
│                       Service 层                              │
│  idgen    │ status      │ authsession │ dfs   │ media │ biz  │
│ Snowflake │ Redis 90s   │ 密钥/会话    │ MinIO │ 元数据 │ 聚合 │
│           │ TTL         │ 管理        │ +HTTP │       │ 7子服务│
└──────────────────────────────────────────────────────────────┘
```

---

## 服务启动顺序与端口

| 顺序 | 服务          | etcd 键                   | gRPC 端口 | 额外端口                     | 说明                     |
|:-----|:-------------|:--------------------------|:----------|:-----------------------------|:-------------------------|
| 1    | idgen        | `service.idgen`           | 20660     |                              | Snowflake 分布式 ID 生成  |
| 2    | status       | `service.status`          | 20670     |                              | 在线状态（Redis, 90s TTL）|
| 3    | authsession  | `service.authsession`     | 20450     |                              | 认证密钥/会话管理         |
| 4    | dfs          | `service.dfs`             | 20640     | HTTP 11701（文件下载）        | 分布式文件存储（MinIO）   |
| 5    | media        | `service.media`           | 20650     |                              | 媒体元数据管理            |
| 6    | biz          | `service.biz_service`     | 20020     |                              | 核心业务逻辑聚合（7 子服务）|
| 7    | msg          | `messenger.msg`           | 20030     |                              | 消息发送/收件箱处理       |
| 8    | sync         | `messenger.sync`          | 20420     |                              | 实时更新同步              |
| 9    | bff          | `bff.bff`                 | 20010     |                              | BFF 聚合层（22 服务）     |
| 10   | session      | `interface.session`       | 20120     |                              | MTProto 会话/RPC 路由     |
| 11   | gateway      | `interface.gateway`       | 20110     | TCP 10443, 5222, 8801        | MTProto TCP 网关           |

---

## 请求流转路径

```
iOS/Android/Desktop 客户端
    │  MTProto (TCP, 端口 10443/5222/8801)
    ▼
 Gateway ──── 传输编解码（自动检测 5 种编码）
    │         DH 密钥交换（RSA + AES-IGE）
    │         AES-IGE 加密/解密
    │         10MB LRU auth_key 缓存
    │         一致性哈希选择 Session 节点
    │  gRPC (SessionSendDataToSession)
    ▼
 Session ──── 剥离 invokeWithLayer/initConnection 等包装
    │         白名单校验（未登录可调用的 RPC）
    │         根据 BFFProxyClients.IDMap 路由到 BFF
    │  gRPC
    ▼
   BFF   ──── 22 个注册的 gRPC 服务
    │         调用下游 Service 层
    │         发布 Sync-T 到 Kafka
    │  gRPC
    ▼
 Service 层 (biz / msg / media / dfs / ...)
    │
    ├──→ MySQL (持久化, 消息以 JSON 存储)
    ├──→ Redis (缓存/状态/去重/auth_key)
    ├──→ MinIO (文件对象存储)
    └──→ Kafka (异步消息)
              │
              ├── Inbox-T ──→ msg 消费(Inbox-MainCommunity-S) ──→ 收件箱处理
              └── Sync-T  ──→ sync 消费(Sync-MainCommunity-S) ──→ 查询在线状态
                                                                ──→ Session
                                                                ──→ Gateway
                                                                ──→ 客户端
```

---

## Gateway 详解

位置：`app/interface/gateway/`

### 核心数据结构

```go
type Server struct {
    c              *config.Config
    server         *net2.TcpServer2          // TCP 服务器
    cache          *cache.LRUCache           // 10MB auth_key 内存缓存
    handshake      *handshake                // MTProto 密钥交换处理器
    session        *Session                  // Session 集群客户端（一致性哈希）
    authSessionMgr *authSessionManager       // authKeyId → sessionId → connId 映射
    timer          *timer2.Timer             // 定时器轮（1024 槽）
}
```

### 传输编解码（自动检测）

Gateway 通过 peek 连接首字节自动判断传输编码：

| 首字节               | 传输类型                    |
|:---------------------|:---------------------------|
| `0xef`               | Abridged（1 字节长度头）    |
| `0xeeeeeeee`         | Intermediate                |
| `0xdddddddd`         | Padded Intermediate         |
| 第 2 个 4 字节 = 0   | Full                        |
| HTTP 动词（HEAD/POST）| HTTP Proxy                  |
| 64 字节混淆头         | Obfuscated（AES-CTR-128 包裹内层编码）|

注册协议名：`mtproto`，通过 `net2.RegisterProtocol("mtproto", ...)` 注册。

### MTProto 握手流程（DH 密钥交换）

状态机：`STATE_CONNECTED2` → `STATE_pq_res` → `STATE_DH_params_res` → `STATE_dh_gen_res` → `STATE_AUTH_KEY`

```
客户端                          Gateway
  │                               │
  │── req_pq / req_pq_multi ────→ │  返回 server_nonce + PQ 挑战
  │←── resPQ ─────────────────── │  (PQ = 0x17ED48941A08F981, 硬编码)
  │                               │
  │── req_DH_params ─────────→   │  RSA 解密, 支持 4 种 InnerData:
  │                               │  PQInnerData / PQInnerDataDc
  │                               │  PQInnerDataTemp / PQInnerDataTempDc
  │←── server_DH_inner_data ─── │  DH 参数 (g=3, 2048-bit prime)
  │                               │
  │── set_client_DH_params ───→  │  计算 authKey = g_b^a mod p
  │←── dh_gen_ok ───────────── │  保存 authKey 到 authsession 服务
  │                               │
```

### 消息处理

- **AuthKeyId == 0**：未加密消息 → 走握手流程
- **AuthKeyId != 0**：已加密消息 → AES-IGE 解密 → 提取 sessionId → 通过一致性哈希转发至 Session 服务
- **反向推送**：Session 调用 `GatewaySendDataToGateway` → 通过 `authSessionMgr` 查找连接 → AES-IGE 加密写回客户端

### 一致性哈希

`Session` 结构使用 `hash.ConsistentHash` + etcd `discov.NewSubscriber` 实现 Session 节点的一致性路由，确保同一 authKeyId 粘滞到同一 Session 实例。

---

## Session 详解

位置：`app/interface/session/`

### 核心数据结构

```go
type Service struct {
    ac              config.Config
    mu              sync.RWMutex
    sessionsManager map[int64]*authSessions    // authKeyId → authSessions
    eGateServers    map[string]*Gateway         // gatewayId → gRPC 客户端
    reqCache        *RequestManager
    serverId        string
    *dao.Dao
}
```

### MTProto 包装剥离

Session 负责递归剥离 MTProto 的 `invoke*` 包装层：

| 包装类型                  | 处理                                    |
|:-------------------------|:----------------------------------------|
| `InvokeWithLayer`        | 记录 API Layer → 递归处理内层            |
| `InitConnection`         | 记录客户端信息（设备、系统、App 版本）→ 递归 |
| `InvokeAfterMsg`         | 处理消息依赖 → 递归                      |
| `InvokeAfterMsgs`        | 处理多消息依赖 → 递归                    |
| `InvokeWithoutUpdates`   | 标记不需要更新 → 递归                    |
| `InvokeWithMessagesRange` | 标记消息范围 → 递归                     |
| `InvokeWithTakeout`      | 标记数据导出 → 递归                      |

### 未登录白名单

以下 RPC 方法不需要登录即可调用：
- `auth.sendCode`, `auth.signIn`, `auth.signUp`, `auth.bindTempAuthKey`
- `help.getConfig`, `help.getNearestDc`
- `langpack.*`（语言包相关所有方法）

其他 RPC 在 `userId == 0` 时返回 401 错误。

### RPC 路由表（IDMap）

Session 根据 `session.yaml` 中的 `BFFProxyClients.IDMap` 将 gRPC 服务路径路由到 BFF：

**已启用的路由**（22 个）：
```
/mtproto.RPCAccount          /mtproto.RPCAuthorization     /mtproto.RPCAutoDownload
/mtproto.RPCChatInvites      /mtproto.RPCChats             /mtproto.RPCConfiguration
/mtproto.RPCContacts         /mtproto.RPCDialogs           /mtproto.RPCDrafts
/mtproto.RPCFiles            /mtproto.RPCMessages          /mtproto.RPCMiscellaneous
/mtproto.RPCNotification     /mtproto.RPCNsfw              /mtproto.RPCPhotos
/mtproto.RPCQrCode           /mtproto.RPCSponsoredMessages /mtproto.RPCTos
/mtproto.RPCUpdates          /mtproto.RPCUsernames         /mtproto.RPCUsers
/mtproto.RPCPremium (注册但存根)
```

**已注释（未实现）的路由**（~20 个）：
```
RPCChannels, RPCVoipCalls, RPCSecretChats, RPCBots, RPCInlineBot, RPCStickers,
RPCReactions, RPCGroupCalls, RPCPayments, RPCPolls, RPCScheduledMessages,
RPCThemes, RPCWallpapers, RPCFolders, RPCEmoji, RPCGames, RPCWebPage,
RPCLangpack, RPCStatistics, RPCTwoFa, RPCGdpr, RPCGifs, RPCDeepLinks ...
```

---

## BFF 详解

位置：`app/bff/`

BFF（Backend-for-Frontend）是一个单体聚合服务，将 22 个 gRPC 服务注册到同一个 `zrpc.RpcServer`：

```go
// app/bff/bff/internal/server/server.go
mtproto.RegisterRPCTosServer(s, tos_helper.New(...))
mtproto.RegisterRPCConfigurationServer(s, configuration_helper.New(...))
mtproto.RegisterRPCQrCodeServer(s, qrcode_helper.New(...))
mtproto.RegisterRPCMiscellaneousServer(s, miscellaneous_helper.New(...))
mtproto.RegisterRPCAuthorizationServer(s, authorization_helper.New(...))
mtproto.RegisterRPCPremiumServer(s, premium_helper.New(...))
mtproto.RegisterRPCChatInvitesServer(s, chatinvites_helper.New(...))
mtproto.RegisterRPCChatsServer(s, chats_helper.New(...))
mtproto.RegisterRPCFilesServer(s, files_helper.New(...))
mtproto.RegisterRPCUpdatesServer(s, updates_helper.New(...))
mtproto.RegisterRPCContactsServer(s, contacts_helper.New(...))
mtproto.RegisterRPCDialogsServer(s, dialogs_helper.New(...))
mtproto.RegisterRPCDraftsServer(s, drafts_helper.New(...))
mtproto.RegisterRPCAutoDownloadServer(s, autodownload_helper.New(...))
mtproto.RegisterRPCMessagesServer(s, messages_helper.New(...))
mtproto.RegisterRPCNotificationServer(s, notification_helper.New(...))
mtproto.RegisterRPCUsersServer(s, users_helper.New(...))
mtproto.RegisterRPCNsfwServer(s, nsfw_helper.New(...))
mtproto.RegisterRPCSponsoredMessagesServer(s, sponsoredmessages_helper.New(...))
mtproto.RegisterRPCAccountServer(s, account_helper.New(...))
mtproto.RegisterRPCPhotosServer(s, photos_helper.New(...))
mtproto.RegisterRPCUsernamesServer(s, usernames_helper.New(...))
```

### BFF 下游依赖

BFF 通过 gRPC 和 Kafka 连接所有下游服务：

| 下游服务          | etcd 键 / 连接方式         | 用途                    |
|:-----------------|:--------------------------|:------------------------|
| biz              | `service.biz_service`     | 核心业务数据（用户/群/会话/消息/更新/用户名/验证码）|
| authsession      | `service.authsession`     | 认证密钥和会话管理       |
| media            | `service.media`           | 媒体元数据              |
| idgen            | `service.idgen`           | ID 生成                 |
| msg              | `messenger.msg`           | 消息发送                |
| dfs              | `service.dfs`             | 文件上传/下载            |
| status           | `service.status`          | 在线状态查询            |
| sync (Kafka)     | Kafka `Sync-T`            | 发布实时更新             |
| Redis            | `localhost:6379`          | 缓存                    |

### Handler 模式

每个 BFF 子模块的 handler 遵循统一模式：

```go
// core.go
type XxxCore struct {
    ctx    context.Context
    svcCtx *svc.ServiceContext     // 持有所有下游客户端
    logx.Logger
    MD *metadata.RpcMetadata       // authKeyId, sessionId, userId, layer, clientAddr
}
```

每个 TL 方法对应一个 `<method>_handler.go` 文件，包含一个处理函数。

---

## Messenger 详解（msg / sync）

### msg 服务

位置：`app/messenger/msg/`

msg 服务同时运行两个子系统：
1. **gRPC 服务器**：注册 `RPCMsgServer`，处理同步消息操作（如 `MsgSendMessage`）
2. **Kafka 消费者**：消费 `Inbox-T` Topic（consumer group: `Inbox-MainCommunity-S`），处理异步收件箱投递

#### msg 的 DAO 层组合

```go
type Dao struct {
    *Mysql                                    // 直连 MySQL（消息表）
    idgen_client.IDGenClient2                 // ID 生成
    user_client.UserClient                    // 用户信息
    chat_client.ChatClient                    // 群聊信息
    inbox_client.InboxClient                  // 收件箱 Kafka 生产者
    SyncClient    sync_client.SyncClient      // 同步 Kafka 生产者
    BotSyncClient sync_client.SyncClient      // Bot 同步
    dialog_client.DialogClient                // 会话操作
    plugin.MsgPlugin                          // 插件扩展
    deduplication.MessageDeDuplicate          // Redis 消息去重
}
```

#### 消息存储

消息以 JSON 格式序列化存储在 `messages` 表的 `message_data` 字段中（`mtproto.Message` proto 转 JSON）。

### sync 服务

位置：`app/messenger/sync/`

sync 是纯 Kafka 消费者（gRPC 部分已注释掉），消费 `Sync-T` Topic 并将更新推送给客户端。

#### 同步类型

| 类型              | 值 | 说明                         |
|:-----------------|:---|:-----------------------------|
| `syncTypeUser`   | 1  | 推送给用户的所有设备          |
| `syncTypeUserNotMe` | 2 | 推送给除发送者外的所有设备   |
| `syncTypeUserMe` | 3  | 只推送给发送者的设备          |

#### 推送流程

1. 从 Kafka 消费更新消息
2. `processUpdates` 分析更新类型（`updateNewMessage`, `updateDeleteMessages`, `updateReadHistoryInbox`, `updateEditMessage` 等）
3. 加入 PTS 队列（`AddToPtsQueue`），确定是否需要推送
4. 查询 status 服务获取用户所有在线 session
5. 按 gateway serverId 分组
6. 通过 `PushUpdatesToSession` gRPC 推送到各 Session 节点
7. 对于 `syncTypeUser`，还会触发离线推送（`PushClient.SyncPushUpdatesIfNot`）

#### 协程池配置

```yaml
Routine:
  Chan: 16    # 通道数
  Size: 100   # 每通道协程数
```

---

## Service 层详解

### biz — 业务逻辑聚合

位置：`app/service/biz/biz/`

单个 gRPC 服务器聚合 **7 个核心数据服务**：

```go
chat.RegisterRPCChatServer(s, ...)         // 群聊 CRUD
code.RegisterRPCCodeServer(s, ...)         // 验证码
dialog.RegisterRPCDialogServer(s, ...)     // 会话 CRUD
message.RegisterRPCMessageServer(s, ...)   // 消息 CRUD
updates.RegisterRPCUpdatesServer(s, ...)   // 更新 CRUD
user.RegisterRPCUserServer(s, ...)         // 用户 CRUD
username.RegisterRPCUsernameServer(s, ...) // 用户名 CRUD
```

每个子服务有独立的 `internal/` 目录，共享同一 MySQL 数据库和 Redis。依赖 `service.media` 和 `service.idgen`。

### dfs — 分布式文件存储

位置：`app/service/dfs/`

| 功能       | 说明                                          |
|:-----------|:----------------------------------------------|
| gRPC 端口  | 20640                                         |
| HTTP 端口  | 11701（MiniHttp，文件下载用）                  |
| 对象存储   | MinIO（localhost:9000, minio/miniostorage）    |
| 文件缓存   | Redis（SSDB 接口兼容）                         |

#### 上传流程（以图片为例）

1. 验证文件分片完整性
2. 从 SSDB 缓存读取分片数据（`OpenFile` + `ReadAll`）
3. 校验 MD5
4. 通过 IDGen 生成 photo_id
5. 图片 resize 多尺寸（s/m/x）
6. 每个尺寸存入 MinIO（`PutPhotoFile`）
7. 返回 `mtproto.Photo` 包含所有尺寸信息

#### 下载流程

根据 `InputFileLocation` 类型分发：
- `inputEncryptedFileLocation` → `GetCacheFile("encryptedfiles", ...)`
- `inputDocumentFileLocation` → `GetCacheFile/GetFile("documents", ...)`
- `inputPhotoFileLocation` → `GetFile("photos"/"videos", ...)`
- `inputPeerPhotoFileLocation` → `GetFile("photos", "a/..."或"c/...")`

### 其他 Service

| 服务         | 关键细节                                                        |
|:-------------|:--------------------------------------------------------------|
| idgen        | Snowflake 算法，通过 etcd 保证 worker_id 唯一                  |
| status       | Redis 存储在线状态，90 秒 TTL，支持查询用户所有在线 session     |
| authsession  | 管理 auth_key 生命周期（permanent/temp/media-temp），MySQL + Redis |
| media        | 图片/文档元数据（photo_sizes/documents 表），依赖 dfs 做实际存储  |

---

## Kafka 消息流

| Topic     | 分区 | 副本 | 生产者            | 消费者组                    | 用途           |
|:----------|:-----|:-----|:-----------------|:---------------------------|:---------------|
| `Inbox-T` | 1    | 0    | msg 服务          | `Inbox-MainCommunity-S`    | 收件箱消息投递  |
| `Sync-T`  | 1    | 0    | msg, bff 服务     | `Sync-MainCommunity-S`     | 实时推送给客户端 |

Topic 由 Kafka 容器启动时自动创建（`KAFKA_CREATE_TOPICS` 环境变量）。

---

## 目录结构

```
teamgram-server/
├── app/                          # 应用代码（所有微服务）
│   ├── bff/                      # Backend-for-Frontend 层
│   │   ├── account/              #   账号（getAccountTTL, updateProfile, ...）
│   │   ├── authorization/        #   登录（sendCode, signIn, signUp, logOut, ...）
│   │   ├── autodownload/         #   自动下载设置
│   │   ├── chatinvites/          #   群邀请链接
│   │   ├── chats/                #   群聊管理（create, edit, addUser, ...）
│   │   ├── configuration/        #   客户端配置（help.getConfig, ...）
│   │   ├── contacts/             #   联系人（importContacts, getContacts, ...）
│   │   ├── dialogs/              #   会话列表（getDialogs, togglePinned, ...）
│   │   ├── drafts/               #   消息草稿
│   │   ├── files/                #   文件上传/下载
│   │   ├── messages/             #   消息操作（sendMessage, getHistory, ...）
│   │   ├── miscellaneous/        #   杂项（help.getSupport, ...）
│   │   ├── notification/         #   推送通知
│   │   ├── nsfw/                 #   NSFW 内容
│   │   ├── photos/               #   图片（uploadProfilePhoto, ...）
│   │   ├── premium/              #   Premium（存根实现）
│   │   ├── qrcode/               #   二维码登录
│   │   ├── sponsoredmessages/    #   赞助消息（存根实现）
│   │   ├── tos/                  #   服务条款
│   │   ├── updates/              #   增量更新（getState, getDifference, ...）
│   │   ├── usernames/            #   用户名（checkUsername, updateUsername）
│   │   ├── users/                #   用户信息（getUsers, getFullUser）
│   │   └── bff/                  #   BFF 聚合服务（注册以上所有子服务）
│   ├── interface/                # 网络接入层
│   │   ├── gateway/              #   TCP 网关
│   │   │   ├── internal/server/  #     TCP 服务、握手、加解密、codec
│   │   │   │   └── codec/        #     传输编解码（abridged/intermediate/padded/full/obfuscated）
│   │   │   └── gateway/          #     生成的 protobuf 代码
│   │   └── session/              #   会话管理
│   │       └── internal/service/ #     RPC 路由、invoke* 剥离、白名单
│   ├── messenger/                # 消息子系统
│   │   ├── msg/                  #   消息发送/收件箱（gRPC + Kafka consumer）
│   │   └── sync/                 #   实时同步（Kafka consumer + 协程池推送）
│   └── service/                  # 核心后端服务
│       ├── authsession/          #   认证会话（auth_key 生命周期管理）
│       ├── biz/                  #   业务逻辑聚合
│       │   ├── biz/              #     聚合器（注册以下 7 个子服务）
│       │   ├── chat/             #     群聊 CRUD
│       │   ├── code/             #     验证码生成/校验
│       │   ├── dialog/           #     会话 CRUD
│       │   ├── message/          #     消息 CRUD
│       │   ├── updates/          #     更新序列管理（PTS）
│       │   ├── user/             #     用户 CRUD
│       │   └── username/         #     用户名 CRUD
│       ├── dfs/                  #   分布式文件存储（MinIO + HTTP 下载）
│       ├── geoip/                #   GeoIP 服务
│       ├── idgen/                #   分布式 ID 生成（Snowflake）
│       ├── media/                #   媒体元数据（photos/documents 表管理）
│       └── status/               #   在线状态（Redis, 90s TTL）
├── pkg/                          # 公共工具包
│   ├── code/                     #   短信/验证码抽象（策略模式接口）
│   ├── conf/                     #   配置辅助
│   ├── deduplication/            #   Redis 消息去重（60s TTL）
│   ├── env2/                     #   环境常量（应用名等）
│   ├── goffmpeg/                 #   FFmpeg Go 封装
│   ├── hashx/                    #   哈希工具
│   ├── httpx/                    #   HTTP 渲染辅助
│   ├── mention/                  #   @提及解析（UTF-16 感知）
│   └── phonenumber/              #   手机号标准化
├── teamgramd/                    # 部署产物
│   ├── bin/                      #   二进制 + 启动脚本 + RSA 密钥 + config.json
│   ├── docker/                   #   Docker 入口脚本
│   ├── etc/                      #   各服务 YAML 配置（11 个）
│   ├── etc2/                     #   Docker 环境生成配置
│   ├── logs/                     #   日志输出
│   ├── sql/                      #   数据库 schema（32 表）+ 18 个迁移文件
│   └── third_party/              #   第三方安装说明（etcd/ffmpeg/kafka/minio/mysql/redis）
├── docker-compose.yaml           # 应用容器编排（单容器暴露所有端口）
├── docker-compose-env.yaml       # 基础设施（7 容器）
├── Makefile                      # 构建 11 个服务
├── build.sh                      # Shell 构建脚本（Docker 用）
└── Dockerfile                    # 多阶段构建（golang:1.19 → ubuntu）
```

---

## 代码组织规范

每个微服务遵循统一的 go-zero 风格目录结构：

```
<service>/
├── client/                    # gRPC 客户端封装（供其他服务调用）
├── cmd/<name>/main.go         # 入口: commands.Run(server.New())
├── internal/
│   ├── config/config.go       # 配置结构体（嵌入 zrpc.RpcServerConf + 自定义字段）
│   ├── core/                  # 业务逻辑处理器
│   │   ├── core.go            #   Core 上下文结构体（ctx, svcCtx, Logger, MD）
│   │   └── <method>_handler.go #  每个 TL 方法一个文件
│   ├── dal/                   # 数据访问层
│   │   ├── dao/mysql_dao/     #   MySQL DAO（CRUD 操作）
│   │   └── dataobject/        #   数据对象（Go struct 映射数据库行）
│   ├── dao/                   # DAO 聚合 + Redis 操作
│   ├── server/                # 服务启动
│   │   └── grpc/service/      #   gRPC 服务注册实现
│   └── svc/                   # 服务上下文（依赖注入容器）
│       └── service_context.go #   初始化所有 client、dao、config
└── plugin/                    # 插件接口（可选，用于扩展）
```

### Handler 函数签名

```go
func (c *XxxCore) MethodName(in *mtproto.TLXxx) (*mtproto.YYY, error) {
    // 1. 从 c.MD 获取 RPC 元数据（userId, authKeyId, ...）
    // 2. 参数校验
    // 3. 调用 c.svcCtx.Dao/Client 执行业务逻辑
    // 4. 返回 mtproto 响应
}
```

### 编码约定

- 入口统一为 `commands.Run(server.New())`
- RPC 方法实现命名为 `<method>_handler.go`（如 `auth.sendCode_handler.go`）
- 数据库操作在 `internal/dal/dao/mysql_dao/` 下
- 由 `mtprotoc` 从 `scheme.tl` 自动生成的文件标记有 `WARNING! All changes made in this file will be lost!`
- 使用 go-zero 的 `mr.Finish` 做并行查询（MapReduce 模式）
- 插件接口支持扩展认证流程（如 `Plugin.CheckPhoneNumberBanned`, `Plugin.CheckSessionPasswordNeeded`）

---

## 数据库设计

数据库名：`teamgram` | 引擎：InnoDB | 字符集：utf8mb4_unicode_ci
Schema 文件：`teamgramd/sql/1_teamgram.sql`（32 张表）

### 认证相关（4 表）

| 表                  | 核心字段                                      | 说明                     |
|:--------------------|:----------------------------------------------|:-------------------------|
| `auth_keys`         | auth_key_id(UNIQUE), body(varchar 512)         | MTProto 授权密钥（base64, 256 字节）|
| `auth_key_infos`    | auth_key_id, auth_key_type, perm/temp/media_id | 密钥类型映射（永久/临时/媒体）|
| `auths`             | auth_key_id, layer, api_id, device_model, ...  | 认证会话元数据（设备、版本、语言、IP）|
| `auth_users`        | auth_key_id, user_id, hash, device_model, ip, country | 密钥与用户绑定关系 |

### 用户相关（9 表）

| 表                           | 核心字段 | 说明                     |
|:-----------------------------|:---------|:-------------------------|
| `users`                      | phone(UNIQUE), first_name, last_name, username, user_type, is_bot, account_days_ttl(默认 180) | 用户账号 |
| `user_contacts`              | owner_user_id, contact_user_id, mutual | 联系人关系 |
| `user_privacies`             | user_id, key_type, rules(JSON) | 隐私规则 |
| `user_presences`             | user_id(UNIQUE), last_seen_at, expires | 最后上线 |
| `user_notify_settings`       | user_id, peer_type, peer_id, show_previews, silent, mute_until | 通知设置 |
| `user_peer_blocks`           | user_id, peer_type, peer_id | 拉黑关系 |
| `user_peer_settings`         | user_id, peer_type, peer_id, report_spam, add_contact, block_contact, ... | 对 peer 设置 |
| `user_profile_photos`        | user_id, photo_id, date2(排序用) | 头像历史 |
| `user_global_privacy_settings` | user_id, archive_and_mute_new_noncontact_peers | 全局隐私 |
| `user_settings`              | user_id, key2, value | 通用 KV 设置 |

### 消息相关（4 表）

| 表                  | 核心字段 | 说明                     |
|:--------------------|:---------|:-------------------------|
| `messages`          | user_id + user_message_box_id(UNIQUE), sender_user_id, peer_type, peer_id, random_id, message_filter_type, **message_data(JSON)**, message(text), mentioned, pinned | 消息（每用户一份 message box）|
| `dialogs`           | user_id + peer_type + peer_id(UNIQUE), pinned, top_message, read_inbox/outbox_max_id, unread_count, unread_mentions_count, draft_message_data(JSON), folder_id | 会话列表 |
| `dialog_filters`    | user_id, dialog_filter_id, dialog_filter(JSON), order_value | 会话文件夹 |
| `hash_tags`         | user_id, hash_tag, hash_tag_message_id | Hashtag 索引 |

### 群聊相关（3 表）

| 表                         | 核心字段 | 说明                     |
|:---------------------------|:---------|:-------------------------|
| `chats`                    | creator_user_id, title, participant_count, photo_id, default_banned_rights, version | 基础群组 |
| `chat_participants`        | chat_id + user_id(UNIQUE), participant_type, admin_rights, inviter_user_id, state | 群成员 |
| `chat_invites`             | chat_id, admin_id, link(UNIQUE), permanent, revoked, expire_date, usage_limit | 邀请链接 |
| `chat_invite_participants` | link + user_id(UNIQUE) | 通过链接加入的用户 |

### 更新/状态（2 表）

| 表                  | 核心字段 | 说明                     |
|:--------------------|:---------|:-------------------------|
| `user_pts_updates`  | user_id, pts, pts_count, update_type, update_data(JSON) | PTS 更新序列 |
| `auth_seq_updates`  | auth_id, user_id, seq, update_type, update_data(JSON) | 每会话序列更新 |

### 媒体/文件（5 表）

| 表                  | 核心字段 | 说明                     |
|:--------------------|:---------|:-------------------------|
| `photos`            | photo_id(UNIQUE), access_hash, has_stickers, has_video, size_id, video_size_id | 图片元数据 |
| `photo_sizes`       | photo_size_id + size_type(UNIQUE), width, height, file_size, file_path, stripped_bytes | 图片尺寸变体 |
| `video_sizes`       | video_size_id + size_type(UNIQUE), width, height, file_size, video_start_ts, file_path | 视频缩略图 |
| `documents`         | document_id(KEY), access_hash, file_path, file_size, mime_type, attributes(JSON) | 文档/文件 |
| `encrypted_files`   | encrypted_file_id, access_hash, key_fingerprint, file_path | 加密文件 |

### 其他（5 表）

| 表                    | 说明                     |
|:----------------------|:-------------------------|
| `devices`             | 推送通知设备 token（按 auth_key_id + user_id + token_type 唯一）|
| `phone_books`         | 客户端通讯录同步         |
| `imported_contacts`   | 联系人导入记录           |
| `unregistered_contacts` | 未注册联系人信息       |
| `popular_contacts`    | 热门联系人排名（按导入次数）|
| `predefined_users`    | 预定义/系统用户（含固定验证码）|
| `username`            | 用户名 → peer_type + peer_id 映射 |
| `bots` / `bot_commands` | 机器人账号和命令       |

### 种子数据

`z_init.sql` 插入系统用户：
- ID: 777000
- 名称: "Teamgram"
- 手机号: "42777"
- 用途: 发送验证码的系统消息

### 迁移记录

18 个迁移文件：`teamgramd/sql/migrate-*.sql`（2022-03 至 2023-07）

---

## 配置文件详解

### gateway.yaml

```yaml
KeyFile: "./server_pkcs1.key"          # RSA 私钥路径
KeyFingerprint: "12240908862933197005"  # RSA 密钥指纹（uint64 字符串）
MaxProc: 4                             # GOMAXPROCS
Server:
  Addrs: [0.0.0.0:10443, 0.0.0.0:5222, 0.0.0.0:8801]  # 3 个 TCP 监听
  ProtoName: mtproto                   # 注册的协议名
  SendBuf: 65536                       # 发送缓冲区
  ReceiveBuf: 65536                    # 接收缓冲区
  Keepalive: false
  SendChanSize: 1024                   # 发送通道大小
```

### session.yaml

关键配置：`BFFProxyClients.IDMap`（完整 RPC 路由表），以及 `Cache`（Redis）、`AuthSession`、`StatusClient`、`GatewayClient` 的 etcd 发现配置。

### bff.yaml

```yaml
Code:
  Name: "none"       # 验证码实现: "none" = 开发模式（只接受 "12345"）
  SendCodeUrl: ""    # 生产环境: 配置 SMS 发送 URL
  VerifyCodeUrl: ""  # 生产环境: 配置验证 URL
```

所有下游服务通过 etcd 发现。`SyncClient` 直连 Kafka（Brokers: 127.0.0.1:9092, Topic: Sync-T）。

### msg.yaml

直连 MySQL（`root:@tcp(127.0.0.1:3306)/teamgram`）和 Redis。同时配置 Kafka 消费者（InboxConsumer）和两个 Kafka 生产者（InboxClient, SyncClient）。

### sync.yaml

```yaml
Routine:
  Chan: 16    # 推送处理通道数
  Size: 100   # 每通道协程数
```

消费 Kafka `Sync-T`，依赖 idgen、status、session 和 biz_service。

### 完整配置文件列表

| 文件                 | 服务        | 关键配置项                          |
|:--------------------|:-----------|:------------------------------------|
| `gateway.yaml`      | gateway    | RSA 密钥, TCP 端口, Session 发现     |
| `session.yaml`      | session    | RPC 路由表, Redis, 4 个服务发现      |
| `bff.yaml`          | bff        | SMS 配置, 7 个服务发现, Kafka Sync   |
| `biz.yaml`          | biz        | MySQL DSN, Redis, media/idgen 发现   |
| `msg.yaml`          | msg        | MySQL DSN, Redis, 2 Kafka 生产+1 消费 |
| `sync.yaml`         | sync       | 协程池, Kafka 消费, MySQL, 4 个服务发现 |
| `idgen.yaml`        | idgen      | etcd 发现                            |
| `status.yaml`       | status     | Redis                                |
| `authsession.yaml`  | authsession| MySQL, Redis                         |
| `dfs.yaml`          | dfs        | MinIO, HTTP 端口 11701, SSDB/Redis   |
| `media.yaml`        | media      | MySQL, Redis, dfs 发现               |

---

## Docker 部署

### 基础设施容器（docker-compose-env.yaml）

7 个容器，网络 `teamgram_net`（172.20.0.0/16）：

| 容器       | 镜像                          | 端口              | 数据卷                     | 关键配置 |
|:-----------|:-----------------------------|:-----------------|:--------------------------|:---------|
| zookeeper  | wurstmeister/zookeeper       | 2181              | `./data/zookeeper/data`   | |
| kafka      | wurstmeister/kafka           | 9092              | `./data/kafka/data`       | 自动创建 Topic: `Inbox-T:1:0,Sync-T:1:0` |
| etcd       | quay.io/coreos/etcd:v3.5.2   | 2379, 2380       | `./data/etcd/data`        | 单节点, 自动压缩 1h |
| redis      | redis                        | 6379              | `./data/redis/data`       | AOF 持久化 |
| mysql      | mysql:5.7                    | 3306              | `./data/mysql/data`       | DB: `teamgram`, 从 `teamgramd/sql/` 初始化 |
| minio      | minio/minio                  | 9000 (API), 9001 (Console) | `./data/minio/data` | Creds: minio/miniostorage |
| minio_mc   | minio/mc                     | -                 | 挂载 `minio_init.sh`      | 初始化桶 |

### 应用容器（docker-compose.yaml）

单容器 `teamgram`（从 Dockerfile 构建），暴露所有 14 个端口。多阶段构建：`golang:1.19` → `ubuntu`。

---

## 关键外部依赖

| 包                                    | 版本     | 用途                  |
|:--------------------------------------|:---------|:---------------------|
| `github.com/teamgram/proto`           | v0.170.0 | MTProto 协议定义/生成代码 |
| `github.com/teamgram/marmota`         | v0.1.19  | 基础设施库（网络/缓存/命令/定时器）|
| `github.com/teamgram/go-zero`         | v1.6.0   | go-zero fork（替换上游）|
| `github.com/zeromicro/go-zero`        | v1.6.0   | go-zero 上游（被 fork 替换）|
| `github.com/Shopify/sarama`           |          | Kafka 客户端          |
| `github.com/minio/minio-go/v7`        |          | MinIO S3 兼容客户端   |
| `github.com/bwmarrin/snowflake`        |          | 分布式 ID 生成        |
| `github.com/nyaruka/phonenumbers`      |          | 手机号解析/验证（libphonenumber Go 版）|
| `google.golang.org/grpc`              | v1.59.0  | gRPC 框架             |
| `go.etcd.io/etcd/client/v3`           | v3.5     | etcd 服务发现          |
| `github.com/oschwald/geoip2-golang`   |          | GeoIP 数据库查询       |
| `github.com/disintegration/imaging`    |          | 图片 resize/处理       |
| `github.com/chai2010/webp`            |          | WebP 图片编解码        |
| `k8s.io/client-go`                    |          | Kubernetes 客户端（可选 k8s 部署）|

---

## 公共工具包 pkg/

### code/ — 短信验证码（策略模式）

```go
type VerifyCodeInterface interface {
    SendSmsVerifyCode(ctx, phoneNumber, code, codeHash string) (string, error)
    VerifySmsCode(ctx, codeHash, code, extraData string) error
}
```

内置实现：
- **`none`**：开发模式，只接受验证码 `"12345"`
- **`me`**：HTTP 发送 SMS（GET `SendCodeUrl?phone=X&code=Y`），对比验证码

工厂：`NewVerifyCode(c SmsVerifyCodeConfig)` 根据 `c.Name` 返回实现。

### deduplication/ — 消息去重

Redis 实现，60 秒 TTL：

```go
type MessageDeDuplicate struct { kv kv.Store }

// HasDuplicateMessage: INCR duplicate_message_id_{userId}_{randomId} + EXPIRE 60s
//   返回 true 如果 counter > 1（重复消息）
// PutDuplicateMessage: 存储原始响应 duplicate_message_data_{userId}_{randomId}
// GetDuplicateMessage: 获取缓存的原始响应
```

用途：客户端重传相同 `randomId` 的消息时，直接返回原始响应而非创建重复消息。

### 其他工具

| 包                | 用途                          |
|:-----------------|:------------------------------|
| `phonenumber/`   | 手机号标准化（`MakePhoneNumberHelper`）|
| `mention/`       | @提及解析（UTF-16 偏移感知）  |
| `hashx/`         | 哈希工具                      |
| `goffmpeg/`      | FFmpeg Go 封装                 |
| `httpx/`         | HTTP 渲染辅助                  |
| `env2/`          | 环境常量（`PredefinedUser2` 开关）|
| `conf/`          | BFFProxyClients 配置结构体     |

---

## 核心业务流程

### 登录流程（auth.sendCode → auth.signIn）

```
客户端                    BFF (authorization)              Service (biz/code)
  │                           │                               │
  │── auth.sendCode ────→     │                               │
  │                           │── 校验 API ID/Hash            │
  │                           │── 标准化手机号                 │
  │                           │── 检查是否被封禁（插件）       │
  │                           │── 频率限制检查                 │
  │                           │── UserGetImmutableUserByPhone ─→ 查用户是否存在
  │                           │                               │
  │                           │   如果用户在线:                │
  │                           │     生成 5 位验证码            │
  │                           │     CodeTypeApp（应用内推送）   │
  │                           │     从 777000 发送消息(login code: XXXXX)
  │                           │                               │
  │                           │   如果测试用户:                │
  │                           │     验证码固定 "12345"         │
  │                           │                               │
  │                           │   否则:                        │
  │                           │     调用 VerifyCodeInterface.SendSmsVerifyCode
  │                           │                               │
  │←── auth.SentCode ────── │                               │
  │                           │                               │
  │── auth.signIn ──────→    │                               │
  │                           │── 校验验证码和 hash            │
  │                           │── VerifySmsCode               │
  │                           │── 如果未注册: 返回 SignUpRequired │
  │                           │── 如果预定义用户: 自动创建      │
  │                           │── AuthsessionBindAuthKeyUser   │
  │                           │── 检查 2FA（插件）             │
  │                           │── 向其他 session 发送登录通知  │
  │←── auth.Authorization ── │                               │
```

### 发送消息流程

```
客户端                    BFF (messages)      msg 服务         sync 服务         客户端B
  │                           │                 │                │                │
  │── messages.sendMessage →  │                 │                │                │
  │                           │── 解析 peer     │                │                │
  │                           │── 校验消息内容   │                │                │
  │                           │── MsgSendMessage ──→ msg         │                │
  │                           │                 │── 去重检查     │                │
  │                           │                 │── 生成 msg_id  │                │
  │                           │                 │── 存入 MySQL   │                │
  │                           │                 │── 发布 Inbox-T │                │
  │                           │                 │── 发布 Sync-T ──→ sync          │
  │                           │                 │                │── 查在线状态   │
  │                           │                 │                │── PTS 队列     │
  │                           │                 │                │── 推送 Session │
  │                           │                 │                │──→ Gateway ──→ │
  │←── Updates ────────────── │                 │                │                │
  │                           │── 异步清除草稿  │                │                │
```

---

## 开发备忘

### 构建与运行

```bash
# 构建所有服务（-tags=jsoniter）
make

# 或单独构建
make gateway
make bff
make msg

# 产出目录
teamgramd/bin/  # 11 个二进制文件

# 启动基础设施
docker-compose -f docker-compose-env.yaml up -d

# 按依赖顺序启动所有服务
cd teamgramd/bin && ./runall2.sh

# Docker 一键部署
docker-compose -f docker-compose-env.yaml up -d  # 先启基础设施
docker-compose up -d                              # 再启应用
```

### 新增 BFF 接口（完整步骤）

1. 在 `app/bff/<module>/internal/core/` 添加 `<method>_handler.go`
2. 实现 handler 函数（遵循 Core struct 模式）
3. 在 `app/bff/<module>/internal/server/grpc/service/` 注册 gRPC 方法
4. 在 `app/bff/bff/internal/server/server.go` 中注册该模块的 helper
5. 在 `teamgramd/etc/session.yaml` 的 `IDMap` 中添加路由（如果是新 proto service）
6. 重新编译 bff 和 session 服务

### 新增 Service

1. 在 `app/service/` 创建目录，遵循代码组织规范
2. 在 `teamgramd/etc/` 添加 YAML 配置
3. 在 `Makefile` 添加构建目标
4. 在 `teamgramd/bin/runall2.sh` 添加启动命令
5. 其他服务通过 etcd key 发现并调用

### 数据库变更

在 `teamgramd/sql/` 添加迁移文件，命名格式：`migrate-YYYYMMDD-description.sql`

### 注意事项

- **自动生成代码**：标记 `WARNING! All changes made in this file will be lost!` 的文件由 mtprotoc 生成，不要手动修改
- **系统用户 777000**：发验证码消息用，不要删除
- **RSA 密钥**：`server_pkcs1.key` 必须与客户端配套，指纹 `12240908862933197005`
- **社区版限制**：Channels、超级群、音视频通话、Bot、贴纸、Reactions、GroupCalls、Payments、SecretChats 等为企业版功能
- **MySQL 密码**：开发环境配置为空密码 `root:@tcp(...)`，Docker 环境通过环境变量 `DB_ROOT_PASSWORD` 设置
- **验证码**：开发环境 `Code.Name: "none"`，只接受 `"12345"`
- **消息去重**：基于 `randomId` + 60s Redis TTL，客户端重传会收到缓存响应
- **PredefinedUser2**：`env2.PredefinedUser2` 开关控制是否启用预定义用户自动创建
