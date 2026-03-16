# Sticker 模块日志对比分析

> 对比 Telegram iOS 客户端 vs OwnPod iOS 客户端启动后点击 Sticker 菜单的完整流程

## 1. OwnPod 错误汇总

### 1.1 STICKER_ID_INVALID 错误 (7 个)

客户端启动时请求 5 种系统内置 sticker set，服务端不认识这些 predicate 类型：

| 请求 | 说明 | 重要性 |
|------|------|--------|
| `getStickerSet(inputStickerSetAnimatedEmoji)` | 内置动画 emoji（聊天表情动画） | **高** — 影响聊天体验 |
| `getStickerSet(inputStickerSetAnimatedEmojiAnimations)` | 内置 emoji 播放动画 | **高** |
| `getStickerSet(inputStickerSetEmojiGenericAnimations)` | 通用 emoji 动画效果 | 中 |
| `getStickerSet(inputStickerSetEmojiDefaultStatuses)` | 默认 emoji 状态列表 | 中 |
| `getStickerSet(inputStickerSetEmojiDefaultTopicIcons)` | 默认话题图标 | 低 |

**Telegram 返回**: 完整的 `messages.StickerSet`（含 set, packs, documents），每个 document 有 thumbs、fileReference 等完整信息。

**我们返回**: `STICKER_ID_INVALID` 错误。

### 1.2 METHOD_NOT_IMPL 错误 (20+ 个)

客户端调用但服务端完全未实现的 API：

| 方法 | 说明 | 重要性 |
|------|------|--------|
| `messages.getSuggestedDialogFilters` | 推荐对话文件夹 | 低 |
| `help.getPremiumPromo` | Premium 推广 | 低 |
| `messages.getDefaultHistoryTTL` | 默认消息自毁时间 | 低 |
| `messages.getAvailableReactions` | 可用的表情回复列表 | **高** — 影响 emoji 回复 |
| `messages.getEmojiGroups` | emoji 分组 | 中 |
| `messages.getEmojiStatusGroups` | emoji 状态分组 | 低 |
| `messages.getEmojiProfilePhotoGroups` | 头像 emoji 分组 | 低 |
| `messages.getAttachMenuBots` | 附件菜单 bot | 低 |
| `account.getSavedRingtones` | 保存的铃声 | 低 |
| `account.getRecentEmojiStatuses` | 最近使用的 emoji 状态 | 低 |
| `account.getDefaultEmojiStatuses` | 默认 emoji 状态 | 低 |
| `account.getDefaultProfilePhotoEmojis` | 默认头像 emoji | 低 |
| `account.getDefaultGroupPhotoEmojis` | 默认群头像 emoji | 低 |
| `messages.getRecentReactions` | 最近使用的 reaction | **高** — 影响 emoji 回复 |
| `messages.getTopReactions` | 热门 reaction | **高** — 影响 emoji 回复 |
| `messages.getEmojiStickers` | emoji sticker 集合 | 中 |
| `messages.getFeaturedEmojiStickers` | 热门 emoji sticker | 低 |
| `messages.getWebPagePreview` | 链接预览 | **已实现** (可能路由未配) |
| `messages.getRecentLocations` | 最近位置 | 低 |
| `messages.getScheduledHistory` | 定时消息 | 低 |
| `help.test` | 测试端点 | 低 |

### 1.3 正常返回 (空数据)

这些 API 已实现，返回了空结果（符合预期 — 新用户没有数据）：

| 方法 | 返回 |
|------|------|
| `messages.getStickers(emoticon: 👋⭐️)` | `stickers(hash: 0, stickers: [])` ✓ |
| `messages.getRecentStickers(hash: 0)` | `recentStickers(hash: 0, packs: [], stickers: [], dates: [])` ✓ |
| `messages.getSavedGifs(hash: 0)` | `savedGifs(hash: 0, gifs: [])` ✓ |
| `messages.getFavedStickers(hash: 0)` | `favedStickers(hash: 0, packs: [], stickers: [])` ✓ |
| `messages.getAllStickers(hash: 0)` | `allStickers(hash: 0, sets: [])` ✓ |
| `messages.getMaskStickers(hash: 0)` | `allStickers(hash: 0, sets: [])` ✓ |
| `messages.getArchivedStickers(...)` | `archivedStickers(count: 0, sets: [])` ✓ |
| `messages.getFeaturedStickers(hash: 0)` | `featuredStickers(...)` ✓ (有数据) |
| `messages.getDialogFilters` | 正常返回 ✓ |

---

## 2. 数据格式差异对比

### 2.1 StickerSet flags 差异

**Telegram `getAllStickers` 响应**:
```
StickerSet.stickerSet(
  flags: 161,                              ← flags 正确设置
  installedDate: Optional(1692798722),     ← 真实时间戳
  id: 891885078961979410,
  thumbs: nil,
  thumbDcId: nil,
  thumbVersion: nil,
  thumbDocumentId: nil,
  count: 60,
  hash: 349290923                          ← 非零 hash
)
```

**OwnPod `getAllStickers` 响应**:
```
allStickers(hash: 0, sets: [])             ← 空（因为没安装）
```

**OwnPod `getFeaturedStickers` 中的 StickerSet**:
```
StickerSet.stickerSet(
  flags: 0,                                ← ❌ flags=0（应包含 animated/video 等标志）
  installedDate: nil,
  id: 2033387742250930176,
  thumbs: nil,                             ← ❌ 缺少 thumbs
  thumbDcId: nil,
  thumbVersion: nil,
  thumbDocumentId: nil,
  count: 3,
  hash: 0                                  ← ❌ hash=0
)
```

### 2.2 Document 格式差异

**Telegram Document**:
```
Document.document(
  flags: 1,                                ← flags=1
  id: 773947703670341874,
  fileReference: 0069b79d42afd180...21b,   ← ✓ 非空 fileReference
  date: 1567439429,
  mimeType: application/x-tgsticker,
  size: 5272,
  thumbs: Optional([                        ← ✓ 有缩略图
    PhotoSize.photoPathSize(type: "j", bytes: ...),
    PhotoSize.photoSize(type: "m", w: 128, h: 128, size: 5038)
  ]),
  dcId: 2,
  attributes: [
    documentAttributeImageSize(w: 512, h: 512),
    documentAttributeSticker(alt: 😂, stickerset: inputStickerSetID(...)),
    documentAttributeFilename(fileName: AnimatedSticker.tgs)
  ]
)
```

**OwnPod Document**:
```
Document.document(
  flags: 0,                                ← ❌ flags=0（应为 1 如果有 thumbs）
  id: 2033387183066320896,
  fileReference: <null>,                   ← ❌ fileReference 为 null！
  date: 1773632257,
  mimeType: application/x-tgsticker,
  size: 5272,
  thumbs: nil,                             ← ❌ 缺少缩略图（客户端无法显示预览）
  dcId: 1,
  attributes: [
    documentAttributeSticker(alt: 😂, stickerset: inputStickerSetID(...)),   ← ✓
    documentAttributeImageSize(w: 512, h: 512),                              ← ✓
    documentAttributeFilename(fileName: AgAD8gADVp29Cg.tgs)                  ← ✓
  ]
)
```

### 2.3 关键差异总结

| 字段 | Telegram | OwnPod | 影响 |
|------|----------|--------|------|
| `StickerSet.flags` | 正确计算（如 49=animated+installed+thumbs） | 始终 0 或仅 animated | 客户端无法判断 set 特性 |
| `StickerSet.hash` | 非零真实 hash | 0 | 客户端每次都要重新请求 |
| `StickerSet.thumbs` | 有 photoPathSize + photoSize | nil | 列表页无缩略图 |
| `StickerSet.thumbDocumentId` | 部分有值 | nil | 部分场景缺图 |
| `Document.flags` | 1（有 thumbs 时） | 0 | 客户端可能不解析 thumbs |
| `Document.fileReference` | 非空 byte array | **null** | **严重** — 可能导致下载失败 |
| `Document.thumbs` | 有 photoPathSize + photoSize | nil | **严重** — 无法显示贴纸预览 |
| `FeaturedStickers` cover 类型 | `stickerSetCovered` / `stickerSetFullCovered` | `stickerSetCovered` | emoji stickers 用 fullCovered |

---

## 3. 需要实现/修复的工作项

### P0 — 必须修复（影响基本使用）

#### 3.1 `getStickerSet` 支持系统内置集合
客户端启动时请求 5 种系统内置 sticker set：
- `inputStickerSetAnimatedEmoji`
- `inputStickerSetAnimatedEmojiAnimations`
- `inputStickerSetEmojiGenericAnimations`
- `inputStickerSetEmojiDefaultStatuses`
- `inputStickerSetEmojiDefaultTopicIcons`

**方案**: 用 Telegram 日志中的数据（set metadata + document list）预置到数据库，`getStickerSet` 识别这些 predicate 时返回预置数据。或者返回空的 `stickerSetNotModified` / 空 set 让客户端静默处理（而不是报错）。

#### 3.2 `Document.fileReference` 不能为 null
当前返回 `<null>`，应返回 `[]byte{}`（空字节数组）。

**方案**: 检查 `makeStickerSetFromDO` 和 Document 序列化/反序列化逻辑，确保 `FileReference = []byte{}` 而非 nil。

#### 3.3 Document 缩略图 (thumbs)
Telegram 的 Document 包含 `photoPathSize`（SVG 路径压缩）和 `photoSize`（小图），客户端用来在列表中显示预览。

**方案**:
- 下载贴纸文件后，生成缩略图（128x128 PNG）
- 或者从 Bot API 的 `thumbnail` 字段获取缩略图并上传
- 将 PhotoSize 列表序列化到 Document 中

### P1 — 改进体验

#### 3.4 `StickerSet.flags` 正确计算
Telegram flags 含义：
- bit 0 (1): archived
- bit 1 (2): official
- bit 2 (4): masks
- bit 3 (8): installed（有 installedDate 时）
- bit 4 (16): has thumbs
- bit 5 (32): animated
- bit 6 (64): videos
- bit 7 (128): emojis

**方案**: `makeStickerSetFromDO` 中根据 set 属性正确设置 flags。

#### 3.5 `StickerSet.hash` 非零
每次返回 hash=0 导致客户端永远不会用 NotModified 缓存。

**方案**: 计算真实 hash（基于 documents 的 id 列表）。

#### 3.6 `messages.getAvailableReactions`
**Telegram 返回**: `availableReactions(hash: 0, reactions: [])` — 即使为空也是合法响应。
**我们返回**: `METHOD_NOT_IMPL`

**方案**: 返回空列表 `messages.AvailableReactions{Hash: 0, Reactions: []}` 即可。

#### 3.7 `messages.getRecentReactions` / `messages.getTopReactions`
**方案**: 返回空列表即可。

#### 3.8 `messages.getEmojiStickers`
等同于 `getAllStickers` 但只返回 `set_type=2 (emojis)` 的集合。

**方案**: 复用 `getAllStickers` 逻辑，过滤 set_type。

### P2 — 可延后

| 方法 | 说明 |
|------|------|
| `messages.getFeaturedEmojiStickers` | Telegram 返回 `stickerSetFullCovered` 类型（含完整 packs + documents） |
| `messages.getEmojiGroups` | emoji 分组 |
| `messages.getEmojiStatusGroups` | emoji 状态分组 |
| `messages.getEmojiProfilePhotoGroups` | 头像 emoji |
| `account.getRecentEmojiStatuses` | |
| `account.getDefaultEmojiStatuses` | |
| `messages.getAttachMenuBots` | |
| `account.getSavedRingtones` | |
| `messages.getScheduledHistory` | |

---

## 4. 利用 Telegram 日志数据的方案

Telegram 日志中包含完整的系统内置 sticker set 数据（AnimatedEmoji, EmojiAnimations, EmojiGenericAnimations, EmojiDefaultStatuses, EmojiDefaultTopicIcons）。可以：

1. **提取 Telegram 返回的 set metadata**: id, accessHash, title, shortName, count, hash
2. **提取 Document 列表**: id, accessHash, fileReference, mimeType, size, attributes, thumbs
3. **提取 StickerPack 列表**: emoticon → document_id mappings

但有一个问题：**Telegram 的 document id 和 accessHash 是 Telegram 服务器的，我们不能直接用**。客户端下载文件时会用这些 id 请求我们的 DFS，而我们的 DFS 没有对应文件。

**可行方案**:
- 不预置完整的 sticker set 数据
- 对这 5 个系统内置请求，返回**空的 stickerSet**（而非报错），让客户端静默处理
- 或者用 Bot API 下载这些系统 set 的文件，上传到我们的 DFS，生成我们自己的 document id

---

## 5. 实施计划（按优先级）

### Phase 1: 消除报错（1-2 天）

1. **`getStickerSet` 处理系统内置 predicate**: 对 `inputStickerSetAnimatedEmoji` 等 5 种类型返回空 set（而非 STICKER_ID_INVALID）
2. **`Document.fileReference` 修复**: 确保始终为 `[]byte{}` 非 nil
3. **返回空列表替代 METHOD_NOT_IMPL**:
   - `messages.getAvailableReactions` → 空 reactions
   - `messages.getRecentReactions` → 空 reactions
   - `messages.getTopReactions` → 空 reactions
   - `messages.getEmojiGroups` → 空 groups
   - `messages.getEmojiStickers` → 空 allStickers
   - `messages.getFeaturedEmojiStickers` → 空 featuredStickers

### Phase 2: 数据完善（3-5 天）

4. **StickerSet.flags 正确计算**: 根据 animated/video/masks/emojis/installed/thumbs 设置 flags
5. **StickerSet.hash 非零**: 按 Telegram 算法计算
6. **Document 缩略图**: 从 Bot API thumbnail 下载并上传，填充 Document.thumbs

### Phase 3: 系统内置 Sticker Set（可选，5+ 天）

7. **用 Bot API 下载 5 个系统内置 set**:
   - AnimatedEmoji (~120 个)
   - AnimatedEmojiAnimations
   - EmojiGenericAnimations (~6 个)
   - EmojiDefaultStatuses
   - EmojiDefaultTopicIcons
