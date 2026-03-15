package core

import (
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/teamgram/proto/mtproto"
	"github.com/teamgram/teamgram-server/app/bff/stickers/internal/dal/dataobject"
)

// resolveInputStickerSetId resolves an InputStickerSet to a set_id and set_type.
func (c *StickersCore) resolveInputStickerSetId(inputSet *mtproto.InputStickerSet) (int64, int32, error) {
	if inputSet == nil {
		return 0, 0, mtproto.ErrStickerIdInvalid
	}

	switch inputSet.GetPredicateName() {
	case mtproto.Predicate_inputStickerSetID:
		setDO, err := c.svcCtx.Dao.StickerSetsDAO.SelectBySetId(c.ctx, inputSet.GetId())
		if err != nil {
			return 0, 0, err
		}
		if setDO == nil {
			return 0, 0, mtproto.ErrStickerIdInvalid
		}
		return setDO.SetId, stickerSetType(setDO), nil

	case mtproto.Predicate_inputStickerSetShortName:
		setDO, err := c.svcCtx.Dao.StickerSetsDAO.SelectByShortName(c.ctx, inputSet.GetShortName())
		if err != nil {
			return 0, 0, err
		}
		if setDO == nil {
			return 0, 0, mtproto.ErrStickerIdInvalid
		}
		return setDO.SetId, stickerSetType(setDO), nil

	default:
		return 0, 0, mtproto.ErrStickerIdInvalid
	}
}

// stickerSetType determines the set_type from a StickerSetsDO.
func stickerSetType(setDO *dataobject.StickerSetsDO) int32 {
	if setDO.IsMasks {
		return 1
	}
	if setDO.IsEmojis {
		return 2
	}
	return 0
}

// setTypeFromFlags determines the set_type from Masks/Emojis flags.
func setTypeFromFlags(masks, emojis bool) int32 {
	if masks {
		return 1
	}
	if emojis {
		return 2
	}
	return 0
}

// MessagesInstallStickerSet handles the messages.installStickerSet TL method.
func (c *StickersCore) MessagesInstallStickerSet(in *mtproto.TLMessagesInstallStickerSet) (*mtproto.Messages_StickerSetInstallResult, error) {
	userId := c.MD.UserId

	setId, setType, err := c.resolveInputStickerSetId(in.GetStickerset())
	if err != nil {
		c.Logger.Errorf("messages.installStickerSet - resolveInputStickerSetId error: %v", err)
		return nil, err
	}

	archived := mtproto.FromBool(in.GetArchived())

	if archived {
		err = c.svcCtx.Dao.UserInstalledStickerSetsDAO.InsertOrUpdate(c.ctx, &dataobject.UserInstalledStickerSetsDO{
			UserId:        userId,
			SetId:         setId,
			SetType:       setType,
			OrderNum:      0,
			InstalledDate: time.Now().Unix(),
			Archived:      true,
		})
		if err != nil {
			c.Logger.Errorf("messages.installStickerSet - InsertOrUpdate(archived) error: %v", err)
			return nil, mtproto.ErrInternelServerError
		}
		return mtproto.MakeTLMessagesStickerSetInstallResultSuccess(nil).To_Messages_StickerSetInstallResult(), nil
	}

	// Shift existing sets' order_num +1 to make room at position 0
	err = c.svcCtx.Dao.UserInstalledStickerSetsDAO.IncrementOrderNum(c.ctx, userId, setType)
	if err != nil {
		c.Logger.Errorf("messages.installStickerSet - IncrementOrderNum error: %v", err)
		return nil, mtproto.ErrInternelServerError
	}

	// Upsert the new set at position 0
	err = c.svcCtx.Dao.UserInstalledStickerSetsDAO.InsertOrUpdate(c.ctx, &dataobject.UserInstalledStickerSetsDO{
		UserId:        userId,
		SetId:         setId,
		SetType:       setType,
		OrderNum:      0,
		InstalledDate: time.Now().Unix(),
		Archived:      false,
	})
	if err != nil {
		c.Logger.Errorf("messages.installStickerSet - InsertOrUpdate error: %v", err)
		return nil, mtproto.ErrInternelServerError
	}

	return mtproto.MakeTLMessagesStickerSetInstallResultSuccess(nil).To_Messages_StickerSetInstallResult(), nil
}

// MessagesUninstallStickerSet handles the messages.uninstallStickerSet TL method.
func (c *StickersCore) MessagesUninstallStickerSet(in *mtproto.TLMessagesUninstallStickerSet) (*mtproto.Bool, error) {
	userId := c.MD.UserId

	setId, _, err := c.resolveInputStickerSetId(in.GetStickerset())
	if err != nil {
		c.Logger.Errorf("messages.uninstallStickerSet - resolveInputStickerSetId error: %v", err)
		return nil, err
	}

	_, err = c.svcCtx.Dao.UserInstalledStickerSetsDAO.SoftDelete(c.ctx, userId, setId)
	if err != nil {
		c.Logger.Errorf("messages.uninstallStickerSet - SoftDelete(%d, %d) error: %v", userId, setId, err)
		return nil, mtproto.ErrInternelServerError
	}

	return mtproto.BoolTrue, nil
}

// MessagesReorderStickerSets handles the messages.reorderStickerSets TL method.
func (c *StickersCore) MessagesReorderStickerSets(in *mtproto.TLMessagesReorderStickerSets) (*mtproto.Bool, error) {
	userId := c.MD.UserId
	setType := setTypeFromFlags(in.GetMasks(), in.GetEmojis())

	for idx, setId := range in.GetOrder() {
		_, err := c.svcCtx.Dao.UserInstalledStickerSetsDAO.UpdateOrder(c.ctx, userId, setId, int32(idx))
		if err != nil {
			c.Logger.Errorf("messages.reorderStickerSets - UpdateOrder(%d, %d, %d) error: %v", userId, setId, idx, err)
			// Continue with remaining sets even if one fails
		}
	}

	_ = setType // setType used to filter, but we trust the client's Order list
	return mtproto.BoolTrue, nil
}

// MessagesGetAllStickers handles the messages.getAllStickers TL method.
func (c *StickersCore) MessagesGetAllStickers(in *mtproto.TLMessagesGetAllStickers) (*mtproto.Messages_AllStickers, error) {
	userId := c.MD.UserId

	// Query installed sets for regular type (0)
	rows, err := c.svcCtx.Dao.UserInstalledStickerSetsDAO.SelectByUserAndType(c.ctx, userId, 0)
	if err != nil {
		c.Logger.Errorf("messages.getAllStickers - SelectByUserAndType(%d, 0) error: %v", userId, err)
		return nil, mtproto.ErrInternelServerError
	}

	if len(rows) == 0 {
		return mtproto.MakeTLMessagesAllStickers(&mtproto.Messages_AllStickers{
			Hash: 0,
			Sets: []*mtproto.StickerSet{},
		}).To_Messages_AllStickers(), nil
	}

	hash := computeInstalledSetsHash(rows)

	if in.GetHash() != 0 && in.GetHash() == hash {
		return mtproto.MakeTLMessagesAllStickersNotModified(nil).To_Messages_AllStickers(), nil
	}

	// Look up full StickerSet metadata for each installed set
	sets := make([]*mtproto.StickerSet, 0, len(rows))
	for _, row := range rows {
		setDO, err2 := c.svcCtx.Dao.StickerSetsDAO.SelectBySetId(c.ctx, row.SetId)
		if err2 != nil || setDO == nil {
			c.Logger.Errorf("messages.getAllStickers - SelectBySetId(%d) error: %v", row.SetId, err2)
			continue
		}

		ss := makeStickerSetFromDO(setDO)
		// Set InstalledDate to mark it as installed
		ss.InstalledDate = &types.Int32Value{Value: int32(row.InstalledDate)}
		sets = append(sets, ss)
	}

	return mtproto.MakeTLMessagesAllStickers(&mtproto.Messages_AllStickers{
		Hash: hash,
		Sets: sets,
	}).To_Messages_AllStickers(), nil
}

// computeInstalledSetsHash computes the Telegram-standard hash over installed set IDs.
func computeInstalledSetsHash(rows []dataobject.UserInstalledStickerSetsDO) int64 {
	if len(rows) == 0 {
		return 0
	}
	var acc uint64
	for _, r := range rows {
		telegramCombineInt64Hash(&acc, uint64(r.SetId))
	}
	return int64(acc)
}
