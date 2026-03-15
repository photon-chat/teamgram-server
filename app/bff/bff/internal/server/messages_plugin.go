package server

import (
	"context"
	"encoding/base64"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/teamgram/marmota/pkg/stores/sqlx"
	"github.com/teamgram/proto/mtproto"

	"github.com/zeromicro/go-zero/core/logx"
)

const recentStickersLimit = 20

type messagesPluginImpl struct {
	db *sqlx.DB
}

func newMessagesPlugin(c sqlx.Config) *messagesPluginImpl {
	return &messagesPluginImpl{
		db: sqlx.NewMySQL(&c),
	}
}

func (p *messagesPluginImpl) GetWebpagePreview(ctx context.Context, url string) (*mtproto.WebPage, error) {
	return nil, nil
}

func (p *messagesPluginImpl) GetMessageMedia(ctx context.Context, ownerId int64, media *mtproto.InputMedia) (*mtproto.MessageMedia, error) {
	return nil, nil
}

func (p *messagesPluginImpl) SaveRecentSticker(ctx context.Context, userId int64, doc *mtproto.Document) {
	if doc == nil {
		return
	}
	log := logx.WithContext(ctx)

	// 1. Extract emoji from documentAttributeSticker
	emoji := ""
	for _, attr := range doc.GetAttributes() {
		if attr.GetPredicateName() == mtproto.Predicate_documentAttributeSticker {
			emoji = attr.GetAlt()
			break
		}
	}

	// 2. Serialize Document to base64
	data, err := proto.Marshal(doc)
	if err != nil {
		log.Errorf("SaveRecentSticker - proto.Marshal error: %v", err)
		return
	}
	docData := base64.StdEncoding.EncodeToString(data)

	// 3. Upsert (UNIQUE KEY (user_id, document_id) prevents duplicates)
	upsertQuery := "INSERT INTO user_recent_stickers(user_id, document_id, emoji, document_data, date2) " +
		"VALUES (?, ?, ?, ?, ?) " +
		"ON DUPLICATE KEY UPDATE document_data = VALUES(document_data), emoji = VALUES(emoji), date2 = VALUES(date2), deleted = 0"
	_, err = p.db.Exec(ctx, upsertQuery, userId, doc.GetId(), emoji, docData, time.Now().Unix())
	if err != nil {
		log.Errorf("SaveRecentSticker - upsert error: %v", err)
		return
	}

	// 4. Trim to 20 entries — soft-delete oldest beyond limit
	trimQuery := "UPDATE user_recent_stickers SET deleted = 1 " +
		"WHERE user_id = ? AND deleted = 0 " +
		"AND id NOT IN (" +
		"  SELECT id FROM (SELECT id FROM user_recent_stickers WHERE user_id = ? AND deleted = 0 ORDER BY date2 DESC LIMIT ?" +
		"  ) AS keep)"
	_, err = p.db.Exec(ctx, trimQuery, userId, userId, recentStickersLimit)
	if err != nil {
		log.Errorf("SaveRecentSticker - trim error: %v", err)
	}
}
