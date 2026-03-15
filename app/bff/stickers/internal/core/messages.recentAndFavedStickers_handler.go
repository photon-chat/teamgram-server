package core

import (
	"time"

	"github.com/teamgram/proto/mtproto"
	"github.com/teamgram/teamgram-server/app/bff/stickers/internal/dal/dataobject"
	"github.com/teamgram/teamgram-server/app/bff/stickers/internal/dao"
	mediapb "github.com/teamgram/teamgram-server/app/service/media/media"
)

// telegramCombineInt64Hash is the Telegram-standard hash algorithm used by all clients.
// It must match the client's combineInt64Hash exactly for NotModified to work.
func telegramCombineInt64Hash(acc *uint64, value uint64) {
	*acc ^= (*acc >> 21)
	*acc ^= (*acc << 35)
	*acc ^= (*acc >> 4)
	*acc += value
}

// MessagesGetRecentStickers handles the messages.getRecentStickers TL method.
func (c *StickersCore) MessagesGetRecentStickers(in *mtproto.TLMessagesGetRecentStickers) (*mtproto.Messages_RecentStickers, error) {
	userId := c.MD.UserId

	rows, err := c.svcCtx.Dao.UserRecentStickersDAO.SelectByUser(c.ctx, userId, 200)
	if err != nil {
		c.Logger.Errorf("messages.getRecentStickers - SelectByUser(%d) error: %v", userId, err)
		return nil, mtproto.ErrInternelServerError
	}

	if len(rows) == 0 {
		return mtproto.MakeTLMessagesRecentStickers(&mtproto.Messages_RecentStickers{
			Hash:     0,
			Packs:    []*mtproto.StickerPack{},
			Stickers: []*mtproto.Document{},
			Dates:    []int32{},
		}).To_Messages_RecentStickers(), nil
	}

	stickers := make([]*mtproto.Document, 0, len(rows))
	dates := make([]int32, 0, len(rows))

	for i := range rows {
		doc, err2 := dao.DeserializeStickerDoc(rows[i].DocumentData)
		if err2 != nil {
			c.Logger.Errorf("messages.getRecentStickers - deserialize document %d error: %v", rows[i].DocumentId, err2)
			continue
		}
		stickers = append(stickers, doc)
		dates = append(dates, int32(rows[i].Date2))
	}

	hash := computeRecentStickersHash(rows)

	if in.GetHash() != 0 && in.GetHash() == hash {
		return mtproto.MakeTLMessagesRecentStickersNotModified(nil).To_Messages_RecentStickers(), nil
	}

	packs := buildUserStickerPacks(rows)

	return mtproto.MakeTLMessagesRecentStickers(&mtproto.Messages_RecentStickers{
		Hash:     hash,
		Packs:    packs,
		Stickers: stickers,
		Dates:    dates,
	}).To_Messages_RecentStickers(), nil
}

// MessagesGetFavedStickers handles the messages.getFavedStickers TL method.
func (c *StickersCore) MessagesGetFavedStickers(in *mtproto.TLMessagesGetFavedStickers) (*mtproto.Messages_FavedStickers, error) {
	userId := c.MD.UserId

	rows, err := c.svcCtx.Dao.UserFavedStickersDAO.SelectByUser(c.ctx, userId, 200)
	if err != nil {
		c.Logger.Errorf("messages.getFavedStickers - SelectByUser(%d) error: %v", userId, err)
		return nil, mtproto.ErrInternelServerError
	}

	if len(rows) == 0 {
		return mtproto.MakeTLMessagesFavedStickers(&mtproto.Messages_FavedStickers{
			Hash:     0,
			Packs:    []*mtproto.StickerPack{},
			Stickers: []*mtproto.Document{},
		}).To_Messages_FavedStickers(), nil
	}

	stickers := make([]*mtproto.Document, 0, len(rows))

	for i := range rows {
		doc, err2 := dao.DeserializeStickerDoc(rows[i].DocumentData)
		if err2 != nil {
			c.Logger.Errorf("messages.getFavedStickers - deserialize document %d error: %v", rows[i].DocumentId, err2)
			continue
		}
		stickers = append(stickers, doc)
	}

	hash := computeFavedStickersHash(rows)

	if in.GetHash() != 0 && in.GetHash() == hash {
		return mtproto.MakeTLMessagesFavedStickersNotModified(nil).To_Messages_FavedStickers(), nil
	}

	packs := buildFavedStickerPacks(rows)

	return mtproto.MakeTLMessagesFavedStickers(&mtproto.Messages_FavedStickers{
		Hash:     hash,
		Packs:    packs,
		Stickers: stickers,
	}).To_Messages_FavedStickers(), nil
}

// MessagesSaveRecentSticker handles the messages.saveRecentSticker TL method.
func (c *StickersCore) MessagesSaveRecentSticker(in *mtproto.TLMessagesSaveRecentSticker) (*mtproto.Bool, error) {
	userId := c.MD.UserId
	inputDoc := in.GetId()
	if inputDoc == nil {
		return mtproto.BoolFalse, nil
	}
	docId := inputDoc.GetId()

	if mtproto.FromBool(in.GetUnsave()) {
		_, err := c.svcCtx.Dao.UserRecentStickersDAO.SoftDelete(c.ctx, userId, docId)
		if err != nil {
			c.Logger.Errorf("messages.saveRecentSticker - SoftDelete(%d, %d) error: %v", userId, docId, err)
			return nil, mtproto.ErrInternelServerError
		}
		return mtproto.BoolTrue, nil
	}

	doc, err := c.svcCtx.Dao.MediaClient.MediaGetDocument(c.ctx, &mediapb.TLMediaGetDocument{Id: docId})
	if err != nil {
		c.Logger.Errorf("messages.saveRecentSticker - MediaGetDocument(%d) error: %v", docId, err)
		return nil, mtproto.ErrInternelServerError
	}

	emoji := extractStickerEmoji(doc)
	docData, err := dao.SerializeStickerDoc(doc)
	if err != nil {
		c.Logger.Errorf("messages.saveRecentSticker - SerializeStickerDoc error: %v", err)
		return nil, mtproto.ErrInternelServerError
	}

	err = c.svcCtx.Dao.UserRecentStickersDAO.InsertOrUpdate(c.ctx, &dataobject.UserRecentStickersDO{
		UserId:       userId,
		DocumentId:   docId,
		Emoji:        emoji,
		DocumentData: docData,
		Date2:        time.Now().Unix(),
	})
	if err != nil {
		c.Logger.Errorf("messages.saveRecentSticker - InsertOrUpdate error: %v", err)
		return nil, mtproto.ErrInternelServerError
	}

	return mtproto.BoolTrue, nil
}

// MessagesClearRecentStickers handles the messages.clearRecentStickers TL method.
func (c *StickersCore) MessagesClearRecentStickers(in *mtproto.TLMessagesClearRecentStickers) (*mtproto.Bool, error) {
	userId := c.MD.UserId

	_, err := c.svcCtx.Dao.UserRecentStickersDAO.ClearByUser(c.ctx, userId)
	if err != nil {
		c.Logger.Errorf("messages.clearRecentStickers - ClearByUser(%d) error: %v", userId, err)
		return nil, mtproto.ErrInternelServerError
	}

	return mtproto.BoolTrue, nil
}

// MessagesFaveSticker handles the messages.faveSticker TL method.
func (c *StickersCore) MessagesFaveSticker(in *mtproto.TLMessagesFaveSticker) (*mtproto.Bool, error) {
	userId := c.MD.UserId
	inputDoc := in.GetId()
	if inputDoc == nil {
		return mtproto.BoolFalse, nil
	}
	docId := inputDoc.GetId()

	if mtproto.FromBool(in.GetUnfave()) {
		_, err := c.svcCtx.Dao.UserFavedStickersDAO.SoftDelete(c.ctx, userId, docId)
		if err != nil {
			c.Logger.Errorf("messages.faveSticker - SoftDelete(%d, %d) error: %v", userId, docId, err)
			return nil, mtproto.ErrInternelServerError
		}
		return mtproto.BoolTrue, nil
	}

	doc, err := c.svcCtx.Dao.MediaClient.MediaGetDocument(c.ctx, &mediapb.TLMediaGetDocument{Id: docId})
	if err != nil {
		c.Logger.Errorf("messages.faveSticker - MediaGetDocument(%d) error: %v", docId, err)
		return nil, mtproto.ErrInternelServerError
	}

	emoji := extractStickerEmoji(doc)
	docData, err := dao.SerializeStickerDoc(doc)
	if err != nil {
		c.Logger.Errorf("messages.faveSticker - SerializeStickerDoc error: %v", err)
		return nil, mtproto.ErrInternelServerError
	}

	err = c.svcCtx.Dao.UserFavedStickersDAO.InsertOrUpdate(c.ctx, &dataobject.UserFavedStickersDO{
		UserId:       userId,
		DocumentId:   docId,
		Emoji:        emoji,
		DocumentData: docData,
		Date2:        time.Now().Unix(),
	})
	if err != nil {
		c.Logger.Errorf("messages.faveSticker - InsertOrUpdate error: %v", err)
		return nil, mtproto.ErrInternelServerError
	}

	return mtproto.BoolTrue, nil
}

// extractStickerEmoji extracts the emoji from the documentAttributeSticker attribute.
func extractStickerEmoji(doc *mtproto.Document) string {
	for _, attr := range doc.GetAttributes() {
		if attr.GetPredicateName() == mtproto.Predicate_documentAttributeSticker {
			return attr.GetAlt()
		}
	}
	return ""
}

// computeRecentStickersHash computes the Telegram-standard hash over recent sticker document IDs.
// Recent stickers use normal order (not reversed).
func computeRecentStickersHash(rows []dataobject.UserRecentStickersDO) int64 {
	if len(rows) == 0 {
		return 0
	}
	var acc uint64
	for _, r := range rows {
		telegramCombineInt64Hash(&acc, uint64(r.DocumentId))
	}
	return int64(acc)
}

// computeFavedStickersHash computes the Telegram-standard hash over faved sticker document IDs.
// Faved stickers use reverse order (reverseHashOrder: true on iOS client).
func computeFavedStickersHash(rows []dataobject.UserFavedStickersDO) int64 {
	if len(rows) == 0 {
		return 0
	}
	var acc uint64
	for i := len(rows) - 1; i >= 0; i-- {
		telegramCombineInt64Hash(&acc, uint64(rows[i].DocumentId))
	}
	return int64(acc)
}

// buildUserStickerPacks groups recent stickers by emoji into StickerPack objects.
func buildUserStickerPacks(rows []dataobject.UserRecentStickersDO) []*mtproto.StickerPack {
	emojiMap := make(map[string][]int64)
	for _, r := range rows {
		if r.Emoji != "" {
			emojiMap[r.Emoji] = append(emojiMap[r.Emoji], r.DocumentId)
		}
	}

	packs := make([]*mtproto.StickerPack, 0, len(emojiMap))
	for emoji, docIds := range emojiMap {
		packs = append(packs, mtproto.MakeTLStickerPack(&mtproto.StickerPack{
			Emoticon:  emoji,
			Documents: docIds,
		}).To_StickerPack())
	}
	return packs
}

// buildFavedStickerPacks groups faved stickers by emoji into StickerPack objects.
func buildFavedStickerPacks(rows []dataobject.UserFavedStickersDO) []*mtproto.StickerPack {
	emojiMap := make(map[string][]int64)
	for _, r := range rows {
		if r.Emoji != "" {
			emojiMap[r.Emoji] = append(emojiMap[r.Emoji], r.DocumentId)
		}
	}

	packs := make([]*mtproto.StickerPack, 0, len(emojiMap))
	for emoji, docIds := range emojiMap {
		packs = append(packs, mtproto.MakeTLStickerPack(&mtproto.StickerPack{
			Emoticon:  emoji,
			Documents: docIds,
		}).To_StickerPack())
	}
	return packs
}
