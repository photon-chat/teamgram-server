# WebPage 预览 — iOS 客户端元素 vs 后端字段映射

> iOS class: `ChatMessageAttachedContentNode`
> 文件: `Telegram-iOS/submodules/TelegramUI/Components/Chat/ChatMessageAttachedContentNode/Sources/ChatMessageAttachedContentNode.swift`

---

## 1. UI 元素一览

| # | UI 节点 | 类型 | 说明 |
|---|---------|------|------|
| 1 | `backgroundView` | `MessageInlineBlockBackgroundView` | 左侧竖线 + 背景色块（引用样式） |
| 2 | `title` | `TextNodeWithEntities` | 站名标题（加粗，带颜色），如 "YouTube"、"GitHub" |
| 3 | `subtitle` | `TextNodeWithEntities` | 页面标题（加粗），最多 5 行 |
| 4 | `text` | `TextNodeWithEntities` | 页面描述，最多 12 行（有媒体时 4-6 行） |
| 5 | `inlineMedia` | `TransformImageNode` | 小缩略图 54×54，显示在文本右侧 |
| 6 | `contentMedia` | `ChatMessageInteractiveMediaNode` | 全宽媒体（图片、视频、贴纸、Story 等） |
| 7 | `contentFile` | `ChatMessageInteractiveFileNode` | 文件附件行（文档、音频） |
| 8 | `actionButton` | `ChatMessageAttachedContentButtonNode` | 底部按钮（"Instant View"、"View Channel" 等） |
| 9 | `actionButtonSeparator` | `SimpleLayer` | 按钮上方的细分隔线 |
| 10 | `statusNode` | `ChatMessageDateAndStatusNode` | 时间戳、已读状态、reactions |
| 11 | `titleBadgeLabel` | `TextNode` | 标题旁小标签（广告用 "What is this?"） |
| 12 | `closeButton` | `ComponentView` | 关闭按钮（广告用） |

---

## 2. WebPage TL 字段 → UI 元素映射

### 2.1 文字类

| WebPage TL 字段 | UI 元素 | 说明 |
|------------------|---------|------|
| `site_name` | `title` | 站名，如 "YouTube"、"Twitter"，加粗显示 |
| `title` | `subtitle` | 页面标题，加粗，最多 5 行 |
| `description` | `text` | 页面描述，最多 12 行 |
| `author` | — | 不直接渲染为独立元素，但影响 `backgroundView` 的颜色 |
| `type` | — | 决定媒体展示方式和按钮文案（见下方） |

### 2.2 媒体类

| WebPage TL 字段 | UI 元素 | 条件 |
|------------------|---------|------|
| `photo` | `inlineMedia` (54×54) | type 为 "article" 或无 instantPage 的普通链接 |
| `photo` | `contentMedia` (全宽) | type 为 "photo"/"video"/"embed"/"gif"/"document"/"telegram_album" |
| `document` | `contentMedia` (全宽) | 视频文件（`.isVideo`）、贴纸 |
| `document` | `contentFile` (文件行) | 其他文件类型（文档、音频） |
| `embed_url` | `contentMedia` | 嵌入式视频/内容（YouTube embed 等） |

### 2.3 嵌入式内容 (Embed)

| WebPage TL 字段 | 用途 |
|------------------|------|
| `embed_url` | 嵌入内容的 URL，决定是否有 embed 播放器 |
| `embed_type` | 嵌入类型（"text/html"、"video/mp4" 等） |
| `embed_width` | embed iframe 的宽度 |
| `embed_height` | embed iframe 的高度 |

**Embed 展示逻辑**:
- 有 `embed_url` + autoplay → 媒体区在文字**上方**（`preferMediaBeforeText`）
- 有 `embed_url` 无 autoplay → 显示 `photo` 或 `document` 作为封面

### 2.4 行为控制

| WebPage TL 字段 | 用途 |
|------------------|------|
| `has_large_media` | `true` 时图片显示为全宽大图，否则可能显示为 54×54 小缩略图 |
| `type` | 决定 actionButton 文案和媒体布局方式 |

---

## 3. 图片尺寸决策逻辑（小图 vs 大图）

```
defaultWebpageImageSizeIsSmall(webpage):
    if type in ["photo", "video", "embed", "gif", "document", "telegram_album"]:
        → 大图 (fullwidth)
    else if type == "article":
        → 小图 (54×54 inline)
    else if 无 instantPage:
        → 小图 (54×54 inline)
    else:
        → 大图 (fullwidth)
```

客户端额外受 `WebpagePreviewMessageAttribute.forceLargeMedia` / `forceSmallMedia` 控制（用户手动切换），以及服务端 `has_large_media` 字段。

---

## 4. Action Button 文案映射

| `type` 值 | 按钮文案 | 图标 |
|-----------|---------|------|
| 有 `instantPage` | "Instant View" | ⚡ |
| "telegram_channel" | "View Channel" | — |
| "telegram_chat" / "telegram_megagroup" | "View Group" | — |
| "telegram_message" | "View Message" | — |
| "telegram_user" | "Send Message" / "Open Profile" | — |
| "telegram_background" | "View Background" | — |
| "telegram_theme" | "View Theme" | — |
| "telegram_botapp" | "Open Bot App" | — |
| "telegram_story" | "Open Story" | — |
| "telegram_stickerset" | "View Stickers" / "View Emojis" | — |
| 无特殊 type | 无按钮 | — |

---

## 5. 后端已支持 vs 未支持

### ✅ 已支持

| # | WebPage 字段 | 后端来源 |
|---|-------------|---------|
| 1 | `id` | fnv64a(URL) |
| 2 | `url` | 原始 URL |
| 3 | `display_url` | host + path 截取 |
| 4 | `hash` | Unix 时间戳 |
| 5 | `type` | og:type，默认 "article" |
| 6 | `site_name` | og:site_name |
| 7 | `title` | og:title / `<title>` |
| 8 | `description` | og:description / `<meta name="description">` |
| 9 | `photo` | og:image → 下载 → DFS 上传 |
| 10 | `embed_url` | og:video / og:video:url / og:video:secure_url |
| 11 | `embed_type` | og:video:type |
| 12 | `embed_width` | og:video:width |
| 13 | `embed_height` | og:video:height |
| 14 | `author` | article:author / `<meta name="author">` |
| 15 | `has_large_media` | 有 photo 时设为 true |
| 16 | `date` | 当前 Unix 时间戳 |

直接图片 URL（.jpg/.png 等）：type="photo"，photo 字段填充。

### ❌ 未支持（需要额外实现）

| # | WebPage 字段 | 说明 | 复杂度 |
|---|-------------|------|--------|
| 1 | `document` | 嵌入视频的实际文件（需下载 + 上传为 Document） | 高 |
| 2 | `cached_page` (Instant View) | Telegram 独有的 IV 页面缓存 | 极高（需 IV 引擎） |
| 3 | `duration` | 视频时长（og:video 没有标准 meta，需其他方式获取） | 中 |

---

## 5.1 telegram_* 内部链接 type — 架构说明

### 关键结论

**`type` 字段完全由服务端设置**，客户端只读取它来决定 UI 布局和按钮文案。

iOS 代码确认链路：
1. 客户端发消息时提取 URL → 调用 `messages.getWebPagePreview(message: url)`
2. 服务端返回 `webPage` TL 对象，其中 `type` 字段直接传递（`TelegramMediaWebpage.swift:47`）
3. 客户端读取 type → `ChatMessageWebpageBubbleContentNode` 决定按钮文案和布局

### t.me URL → type 映射表

| URL 模式 | type | Action Button |
|----------|------|---------------|
| `t.me/{username}` (频道) | `telegram_channel` | "View Channel" |
| `t.me/{username}` (超级群) | `telegram_megagroup` | "View Group" |
| `t.me/{username}` (普通群) | `telegram_chat` | "View Group" |
| `t.me/{username}` (用户) | `telegram_user` | "Send Message" |
| `t.me/{username}?profile` | `telegram_user` | "Open Profile" |
| `t.me/{channel}/{msgId}` | `telegram_message` | "View Message" |
| `t.me/c/{channelId}/{msgId}` | `telegram_message` | "View Message" |
| `t.me/addstickers/{setname}` | `telegram_stickerset` | "View Stickers" |
| `t.me/addemoji/{setname}` | `telegram_stickerset` | "View Emojis" |
| `t.me/bg/{slug}` | `telegram_background` | "View Background" |
| `t.me/addtheme/{name}` | `telegram_theme` | "View Theme" |
| `t.me/{bot}/{app}` | `telegram_botapp` | "Open Bot App" |
| `t.me/{username}/s/{id}` | `telegram_story` | "Open Story" |
| `t.me/addlist/{slug}` | `telegram_chatlist` | "Open Chat Folder" |
| `t.me/boost/{channel}` | `telegram_channel_boost` | "Boost Channel" |
| `t.me/{channel}?voicechat` | `telegram_voicechat` | 语音聊天 |
| `t.me/+{hash}` (频道邀请) | `telegram_channel_request` | "Request to Join" |
| `t.me/+{hash}` (群邀请) | `telegram_chat_request` | "Request to Join" |

### 实现方案

在 `GetWebpagePreview` 中先检查是否为 t.me URL，如果是则走内部路由：
1. 解析 URL path 确定链接类型
2. 查询内部服务（username 解析、频道/用户/贴纸包信息）
3. 构造带正确 type 的 WebPage 返回

需要的内部服务：
- `UsernameClient` — 解析 username → peer (user/channel/chat)
- `UserClient` — 获取用户信息（名称、头像）
- `ChatClient` — 获取群/频道信息（标题、头像、成员数）
- `StickerSetsDAO` — 获取贴纸包信息

---

## 6. 典型场景示例

### 6.1 普通文章（如新闻链接）
```
┌─────────────────────────┐
│▎CNN                     │  ← title (site_name)
│▎Breaking: ...           │  ← subtitle (title)
│▎The event happened... ┌───┐
│▎yesterday in downtown │img│ ← inlineMedia (54×54, article 类型小图)
│▎...                   └───┘
│▎                            ← text (description)
└─────────────────────────┘
```

### 6.2 YouTube 视频链接
```
┌─────────────────────────┐
│▎YouTube                 │  ← title (site_name)
│▎┌─────────────────────┐ │
│▎│ ▶  Video Thumbnail  │ │  ← contentMedia (全宽, embed type)
│▎└─────────────────────┘ │
│▎Video Title             │  ← subtitle (title)
│▎Video description...    │  ← text (description)
└─────────────────────────┘
```
embed_url、embed_type、embed_width、embed_height 控制播放器参数。

### 6.3 直接图片 URL
```
┌─────────────────────────┐
│▎example.com             │  ← title (host)
│▎photo.jpg               │  ← subtitle (filename)
│▎┌─────────────────────┐ │
│▎│                     │ │  ← contentMedia (全宽, type="photo")
│▎│    Photo Preview    │ │
│▎│                     │ │
│▎└─────────────────────┘ │
└─────────────────────────┘
```

### 6.4 有大图的文章（has_large_media=true）
```
┌─────────────────────────┐
│▎House Beautiful         │  ← title (site_name)
│▎50 Best Kitchen Ideas   │  ← subtitle (title)
│▎Transform your kitchen..│  ← text (description)
│▎┌─────────────────────┐ │
│▎│                     │ │  ← contentMedia (全宽, has_large_media)
│▎│    Article Photo    │ │
│▎│                     │ │
│▎└─────────────────────┘ │
└─────────────────────────┘
```
