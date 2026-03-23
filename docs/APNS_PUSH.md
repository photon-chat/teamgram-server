# APNs 推送系统文档

## 概述

TeamGram 的 Apple Push Notification (APNs) 推送系统，用于向离线 iOS 设备发送实时推送通知。

## 配置参数

| 参数 | 值 | 说明 |
|-----|----------|------|
| KeyFile | `../etc/AuthKey_JH5C27A29G.p8` | APNs 认证密钥文件路径 |
| KeyID | `JH5C27A29G` | Apple Key ID |
| TeamID | `3WA4Q9D2GD` | Apple Team ID |
| BundleID | `org.delta.pchat` | iOS 应用 Bundle ID |
| Production | dev: `false` / prod: `true` | 由 docker-compose 自动控制 |

## 部署（零配置）

APNs 配置通过 Docker 环境变量自动注入，新机器直接启动即可：

```bash
# 开发环境
docker-compose build && docker-compose up -d

# 生产环境
docker-compose -f docker-compose.prod.yaml build && docker-compose -f docker-compose.prod.yaml up -d
```

不需要运行 `configure-apns.sh`，不需要编辑任何配置文件。`entrypoint.sh` 自动完成一切。

## 推送决策逻辑：什么时候推、什么时候不推

### 核心原则

**TCP 连接在且用户在前台 → 消息走 TCP → 不推送。用户报告离线(后台)或 TCP 断开 → 推 APNs。**

### 详细决策流程

```
消息到达 Sync 服务
  │
  ▼
SyncPushUpdates (Kafka Sync-T)
  │
  ├─ 1. 推送到所有在线 session（TCP 直连）
  │
  ├─ 2. 收集在线 session 的 AuthKeyId 列表 → excludeList
  │
  ├─ 3. 查询 UserGetLastSeen(userId)
  │     └─ 如果 expires == 0（用户已上报 offline）→ 清空 excludeList
  │        （即使 TCP 连接还在也推 APNs，解决后台30秒间隙问题）
  │
  └─ 4. 调用 SyncPushUpdatesIfNot（异步）
       │
       ├─ 查询 devices 表：该用户所有有效设备 token
       │
       ├─ 按 token 去重（同设备多次登录只推一次）
       │
       ├─ 排除 excludeList 中在线的设备
       │
       └─ 对每个离线设备 → SendAPNsPush → Apple 推送服务器
```

### 后台间隙问题（已解决）

**问题**: iOS 切到后台时，socket 连接还会保持约30秒，这段时间消息通过 TCP 送达但用户看不到。

**解决方案**: iOS 切后台时会主动调用 `account.updateStatus(offline: true)`，服务端将 `user_presences.expires` 设为 0。推送决策时，除了检查 socket 连接状态，还会检查 `UserGetLastSeen` 的 `expires` 字段。如果用户已上报 offline，即使 socket 还在也会发送 APNs 推送。

### iOS 客户端后台行为

| 阶段 | MTProto 连接 | 服务端认为 | 推送行为 |
|------|------------|----------|---------|
| app 在前台 | `resume()` 保持连接 | 在线 (expires=300) | 不推送（TCP 送达，app 显示 in-app 横幅） |
| 刚切后台 (~30秒) | 后台保活期间连接仍在 | 在线但已报 offline (expires=0) | **推送**（检测到 expires=0，清空 excludeList） |
| 后台保活结束 | `pause()` → TCP 断开 | 离线 | **推送**（APNs 是唯一通道） |
| 收到 APNs 推送 | 临时 `resume()` ~29秒 | 短暂在线 | iOS 唤醒 app，执行 `pollState` 拉取新消息 |
| app 被系统杀掉 | 无连接 | 离线 | **推送**（APNs 是唯一通道） |

### iOS 客户端通知显示

| 场景 | 通知方式 | 说明 |
|------|---------|------|
| app 在前台 | 自定义 in-app 横幅 | 顶部滑入/5秒消失/可点击跳转/可上滑关闭 |
| app 在后台 | APNs 系统推送 | iOS 系统通知栏显示 |
| app 被杀掉 | APNs 系统推送 | iOS 系统通知栏显示 |
| 正在查看该对话 | 不通知 | 消息直接显示在聊天界面 |

**注意**: iOS 客户端对消息**不使用本地通知**。本地通知仅用于来电（CallKit 不可用时）。所有离线消息通知完全依赖服务端 APNs 推送。

## 完整推送流程

```
┌─────────────────────────────────────────────────────────────┐
│                    iOS 客户端启动                             │
│  1. 请求通知权限                                             │
│  2. registerForRemoteNotifications() → 获取 APNs token       │
│  3. account.registerDevice(tokenType:1, token, appSandbox)   │
└──────────────────────────┬──────────────────────────────────┘
                           │ MTProto RPC
                           ▼
┌─────────────────────────────────────────────────────────────┐
│              BFF Notification Service                        │
│  account.registerDevice_handler.go                           │
│  → INSERT INTO devices (ON DUPLICATE KEY UPDATE)             │
│  → secret 字段 hex 编码存储                                   │
└──────────────────────────┬──────────────────────────────────┘
                           │ 写入 MySQL
                           ▼
┌─────────────────────────────────────────────────────────────┐
│              devices 表 (MySQL)                              │
│  user_id | auth_key_id | token | token_type | state          │
│  777004  | 260387...   | da95..| 1          | 0 (有效)       │
└──────────────────────────────────────────────────────────────┘

                    === 消息发送时 ===

┌─────────────────────────────────────────────────────────────┐
│  用户A 发送消息给 用户B                                       │
│  Msg Service → Kafka Sync-T topic                            │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│  Sync Service 消费 Sync-T                                    │
│  SyncPushUpdates handler (core.go)                           │
│                                                              │
│  ① StatusGetUserOnlineSessions(userId=B)                     │
│     → 获取 B 的所有在线 session                               │
│                                                              │
│  ② 对每个在线 session: PushUpdatesToSession (TCP 直连)        │
│     → 在线设备立即收到消息                                     │
│                                                              │
│  ③ 收集在线 session 的 PermAuthKeyId → excludeList           │
│                                                              │
│  ④ UserGetLastSeen(userId=B) 检查用户是否报告 offline         │
│     → 如果 expires == 0 → 清空 excludeList（后台间隙修复）    │
│                                                              │
│  ⑤ 异步调用 SyncPushUpdatesIfNot(userId, excludeList, data)  │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│  SyncPushUpdatesIfNot handler                                │
│  (sync.pushUpdatesIfNot_handler.go)                          │
│                                                              │
│  ① GetUserAPNsDevices(userId=B) → 查 devices 表              │
│                                                              │
│  ② 排除在线设备 (excludeMap[dev.AuthKeyId])                   │
│                                                              │
│  ③ 按 token 去重 (pushedTokens map) — 防止重复推送            │
│                                                              │
│  ④ extractPushPayload(updates) — 提取消息内容                 │
│     支持: updateShortMessage / updateShortChatMessage         │
│           updates(updateNewMessage) / updateShort             │
│                                                              │
│  ⑤ 对每个离线设备: SendAPNsPush(token, payload)               │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│  SendAPNsPush (push.go)                                      │
│  ① 构建 APNs payload (alert/sound/badge/custom data)         │
│  ② apns2.Client.PushWithContext()                            │
│  ③ 处理响应:                                                  │
│     - 成功 → 日志记录                                         │
│     - BadDeviceToken/Unregistered/ExpiredToken → state=1      │
│     - 其他错误 → 日志记录                                     │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│  Apple APNs 服务器 → iOS 设备                                │
│                                                              │
│  iOS 收到推送后:                                              │
│  ① 显示系统通知 (banner/sound/badge)                          │
│  ② 如果 app 未被杀掉: 临时唤醒 ~29秒                          │
│     → mtProto.resume() → pollState() → 拉取完整消息数据       │
│  ③ 用户点击通知 → 打开 app → 进入对应对话                      │
└─────────────────────────────────────────────────────────────┘
```

## devices 表

### 结构

```sql
CREATE TABLE `devices` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `auth_key_id` bigint(20) NOT NULL,
  `user_id` bigint(20) NOT NULL,
  `token_type` int(11) NOT NULL,        -- 1: iOS APNs
  `token` varchar(512) NOT NULL,         -- APNs device token
  `no_muted` tinyint(1) NOT NULL DEFAULT '0',
  `app_sandbox` tinyint(1) NOT NULL DEFAULT '0',
  `secret` varchar(1024) NOT NULL DEFAULT '',  -- hex 编码存储
  `other_uids` varchar(1024) NOT NULL DEFAULT '',
  `state` tinyint(1) NOT NULL DEFAULT '0',  -- 0: 有效, 1: 无效
  PRIMARY KEY (`id`),
  UNIQUE KEY (`auth_key_id`, `user_id`, `token_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

已包含在 `teamgramd/sql/1_teamgram.sql` 中，随数据库初始化自动创建。

### 多设备/多账号场景

| 场景 | devices 表记录 | 推送行为 |
|------|--------------|---------|
| 1 个用户 1 台设备 | 1 条 (user_id=A, token=X) | 离线时推到设备 X |
| 1 个用户 2 台设备 | 2 条 (token=X, token=Y) | 每台离线设备各推一次 |
| 2 个用户 1 台设备 | 2 条 (user_id=A token=X, user_id=B token=X) | 各自独立推送 |
| 同设备重新登录 | 新增一条 (新 auth_key_id, 同 token) | 按 token 去重，只推一次 |

### 数据清理

- 无效 token 自动标记: APNs 返回 BadDeviceToken/Unregistered/ExpiredToken 时，`state` 自动设为 1
- 可定期清理过期数据：`DELETE FROM devices WHERE state = 1 AND updated_at < DATE_SUB(NOW(), INTERVAL 30 DAY)`

## 代码文件

| 层级 | 文件 | 功能 |
|------|------|------|
| Config | `app/messenger/sync/internal/config/config.go` | APNsConfig 结构定义 |
| DAO Init | `app/messenger/sync/internal/dao/dao.go` | APNs 客户端 + UserClient 初始化 |
| Push | `app/messenger/sync/internal/dao/push.go` | GetUserAPNsDevices / SendAPNsPush |
| 推送触发 | `app/messenger/sync/internal/core/core.go:231-258` | pushUpdatesToSession 中触发 APNs + 后台间隙检测 |
| 推送逻辑 | `app/messenger/sync/internal/core/sync.pushUpdatesIfNot_handler.go` | 排除在线/去重/extractPushPayload |
| 设备注册 | `app/bff/notification/internal/core/account.registerDevice_handler.go` | 处理 iOS 客户端注册请求 |
| 设备注销 | `app/bff/notification/internal/core/account.unregisterDevice_handler.go` | 处理设备注销 |
| 设备 DB | `app/bff/notification/internal/dao/devices.go` | RegisterDevice / UnregisterDevice |
| 状态上报 | `app/bff/account/internal/core/account.updateStatus_handler.go` | iOS 后台上报 offline |
| 状态存储 | `app/service/biz/user/internal/dao/user_status.go` | PutLastSeenAt / GetLastSeenAt |

## 推送负载格式

### 标准推送

```json
{
  "aps": {
    "alert": { "title": "Alice", "body": "Hello!" },
    "sound": "default", "badge": 1, "mutable-content": 1
  },
  "custom": { "from_id": 123, "msg_id": 1, "peer_type": "user", "peer_id": 456 }
}
```

### 群组消息

```json
{
  "aps": {
    "alert": { "title": "New Message", "body": "Let's meet" },
    "sound": "default", "badge": 1, "mutable-content": 1
  },
  "custom": { "from_id": 123, "msg_id": 2, "peer_type": "chat", "peer_id": 999, "chat_id": 999 }
}
```

### 静音推送

```json
{
  "aps": { "content-available": 1, "mutable-content": 1 },
  "custom": { "from_id": 123, "msg_id": 1, "peer_type": "user", "peer_id": 456 }
}
```

## 错误处理

| APNs 响应 | 服务端处理 |
|-----------|----------|
| 200 Success | 日志记录 ApnsID |
| 400 BadDeviceToken | 标记 state=1，停止推送到该 token |
| 410 Unregistered | 标记 state=1 |
| 410 ExpiredToken | 标记 state=1 |
| 其他错误 | 日志记录，不标记 |

## 故障排查

```bash
# 1. APNs 客户端是否初始化
docker exec <container> grep -i "apns" /app/logs/sync/access.log

# 2. 推送有没有发出
docker exec <container> grep "SendAPNsPush\|pushUpdatesIfNot" /app/logs/sync/access.log | tail -20

# 3. 设备注册有没有成功
docker exec <container> grep -i "registerDevice" /app/logs/bff/access.log | tail -5

# 4. 有没有注册错误
docker exec <container> grep -i "device\|register" /app/logs/bff/error.log | tail -10

# 5. 数据库里有没有设备
docker exec <mysql-container> mysql -uroot -p<password> teamgram -e "SELECT user_id, LEFT(token,20), state FROM devices;"
```

## 测试

```bash
# Go 单元测试（含 p8 真实连接测试）
go test -vet=off ./app/messenger/sync/internal/dao/ -run "TestP8|TestAPNs|TestPushPayload|TestDeviceInfo" -v

# 配置测试脚本 (15 项检查)
./teamgramd/scripts/test-apns-push.sh
```

## 文件清单

```
teamgram-server/
├── docker-compose.yaml              # 开发 APNs 环境变量 (Production=false)
├── docker-compose.prod.yaml         # 生产 APNs 环境变量 (Production=true)
├── teamgramd/
│   ├── etc/
│   │   ├── sync.yaml                # APNs 配置模板 (Docker 自动注入)
│   │   └── AuthKey_JH5C27A29G.p8    # 认证密钥 (.gitignore)
│   ├── docker/entrypoint.sh         # 自动注入 APNs 配置
│   └── scripts/
│       ├── configure-apns.sh        # 仅非 Docker 裸机部署
│       └── test-apns-push.sh        # 配置测试脚本
├── app/messenger/sync/internal/
│   ├── config/config.go
│   ├── dao/{dao,push,push_test}.go
│   └── core/{core,sync.pushUpdatesIfNot_handler}.go
├── app/bff/notification/internal/
│   ├── core/{registerDevice,unregisterDevice}_handler.go
│   └── dao/devices.go
└── docs/
    ├── APNS_PUSH.md                 # 本文档
    └── APNS_PUSH_QUICKSTART.md
```

---

**更新日志:**

| 日期 | 说明 |
|------|------|
| 2026-03-23 | 初始版本: 推送系统实现 |
| 2026-03-23 | Docker 零配置自动化; 修复 secret hex 编码; 修复 Push-T 无消费者; 按 token 去重防重复推送; 完整推送决策逻辑文档 |
| 2026-03-23 | 修复后台间隙问题: 检查 UserGetLastSeen expires=0 时清空 excludeList，iOS 切后台即推APNs |
