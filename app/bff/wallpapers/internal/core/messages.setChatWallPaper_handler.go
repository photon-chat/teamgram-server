package core

import (
	"github.com/teamgram/proto/mtproto"
)

// MessagesSetChatWallPaper
func (c *WallpapersCore) MessagesSetChatWallPaper(in *mtproto.TLMessagesSetChatWallPaper) (*mtproto.Updates, error) {
	return mtproto.MakeTLUpdates(&mtproto.Updates{
		Updates: []*mtproto.Update{},
		Users:   []*mtproto.User{},
		Chats:   []*mtproto.Chat{},
		Date:    int32(0),
		Seq:     0,
	}).To_Updates(), nil
}
