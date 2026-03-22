# 自动加群功能

## 概述

用户注册后自动加入两个群：
1. **总群** — 所有新注册用户统一加入
2. **城市群** — 根据用户注册 IP 的 GeoIP 地理位置，加入对应城市群

所有群由系统管理员（小助手 `group_assistant`，ID: 777001）创建并持有群主身份，真实用户作为成员加入。

## 核心流程

```
用户注册
  → autoJoinGroups（同步，在返回 auth 响应前执行）
    → joinAutoGroup（总群，groupType=1）
    → GetCityAndLocaleByIp（通过 GeoIP2 查城市）
    → joinAutoGroup（城市群，groupType=2，groupKey=城市名）
  → 返回 auth 响应
  → 客户端加载 dialogs → 看到群和欢迎消息
```

### 建群逻辑（joinAutoGroup）

1. **查找活跃群**：`SELECT ... FROM auto_groups WHERE group_type=? AND group_key=? AND is_full=0 FOR UPDATE`
2. **无活跃群** → 小助手（777001）创建新群，注册用户作为成员加入
3. **有活跃群** → `ChatAddChatUser` 将用户加入
4. **群满（199人+1系统管理员=200上限）** → 标记当前群已满，小助手创建新群（序号+1），如"总群 2"
5. 每次加入后由小助手发送欢迎消息

### 群满溢出

- `auto_groups.participant_count` 追踪真实用户数（不含系统管理员）
- 达到 199 时标记 `is_full=1`
- 同时有 `ChatAddChatUser` 错误检测作为二级保护
- 新群序号自增：总群 → 总群 2 → 总群 3 ...

## 系统管理员（小助手）

- 用户 ID: 777001
- 用户名: `group_assistant`
- 密码: `assistant@2026`（可登录查看群消息）
- 职责：所有自动群的群主 + 欢迎消息发送者

## 城市群 & 多语言

城市名通过 GeoIP2 MaxMind 数据库获取，优先使用当地语言：

| 国家 | locale | 群名示例 | 欢迎消息语言 |
|------|--------|----------|------------|
| CN/TW/HK | zh-CN | 北京群 | 中文 |
| JP | ja | 東京グループ | 日文 |
| DE/AT/CH | de | Berlin-Gruppe | 德文 |
| ES/MX/AR | es | Grupo Madrid | 西班牙文 |
| FR/BE | fr | Groupe Paris | 法文 |
| BR/PT | pt-BR | Grupo São Paulo | 葡萄牙文 |
| RU/BY/KZ | ru | Группа Москва | 俄文 |
| 其他 | en | London group | 英文 |

总群名称和欢迎消息固定为中文。

## 并发安全

- MySQL 事务 + `SELECT ... FOR UPDATE` 行锁防止并发注册创建重复群
- 错误不阻塞注册流程（仅记录日志）

## 配置

在 `bff.yaml`（或 `etc2/bff.yaml`）中添加：

```yaml
AutoGroupMySQL:
  DSN: root:密码@tcp(mysql:3306)/teamgram?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai
  Active: 64
  Idle: 64
  IdleTimeout: 4h

SystemAdminUserId: 777001
```

注意：`etc/authorization.yaml` 已废弃，生产环境使用合并 BFF 模式（`bff` 二进制 + `bff.yaml`）。

## 数据库

### auto_groups 表（在 1_teamgram.sql 中）

```sql
CREATE TABLE `auto_groups` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `group_type` tinyint(4) NOT NULL COMMENT '1=总群, 2=地区群',
  `group_key` varchar(128) NOT NULL DEFAULT '' COMMENT '总群为空, 地区群为城市名',
  `sequence_num` int(11) NOT NULL DEFAULT '1',
  `chat_id` bigint(20) NOT NULL,
  `creator_user_id` bigint(20) NOT NULL COMMENT '群主（始终为系统管理员）',
  `participant_count` int(11) NOT NULL DEFAULT '1',
  `is_full` tinyint(1) NOT NULL DEFAULT '0',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_chat_id` (`chat_id`),
  KEY `idx_type_key_full` (`group_type`, `group_key`, `is_full`)
);
```

### 系统管理员（在 z_init.sql 中）

自动插入 777001 用户及其登录密码。

## 涉及文件

| 文件 | 说明 |
|------|------|
| `app/bff/authorization/internal/core/auto_join_groups.go` | 核心逻辑 |
| `app/bff/authorization/internal/dao/auto_groups.go` | DAO 层（事务操作） |
| `app/bff/authorization/internal/dao/auto_groups_do.go` | DO 结构体 |
| `app/bff/authorization/internal/dao/api.go` | GeoIP 城市查询 |
| `app/bff/authorization/internal/dao/dao.go` | AutoGroupDB 初始化 |
| `app/bff/authorization/internal/config/config.go` | 配置字段 |
| `app/bff/bff/internal/config/config.go` | 合并 BFF 配置 |
| `app/bff/bff/internal/server/server.go` | 配置传递 |
| `app/bff/authorization/internal/core/auth.signUp_handler.go` | 注册入口（手机+验证码） |
| `app/bff/authorization/internal/core/auth.usernameRegister_handler.go` | 注册入口（用户名+密码） |
| `app/bff/authorization/internal/core/auth.phonePasswordRegister_handler.go` | 注册入口（手机+密码） |
| `teamgramd/sql/1_teamgram.sql` | auto_groups 建表 DDL |
| `teamgramd/sql/z_init.sql` | 777001 用户 + 密码种子数据 |

## 部署注意

1. GeoIP2 数据库文件 `GeoLite2-City.mmdb` 需要存在于工作目录中
2. 城市群功能依赖用户真实公网 IP，内网 IP 无法解析城市
3. MySQL `docker-entrypoint-initdb.d` 仅在首次初始化时执行，已有数据需手动建表
