package core

import (
	"github.com/teamgram/proto/mtproto"
)

// MessagesReportEncryptedSpam
// messages.reportEncryptedSpam#4b0c8c0f peer:InputEncryptedChat = Bool;
func (c *ReportsCore) MessagesReportEncryptedSpam(in *mtproto.TLMessagesReportEncryptedSpam) (*mtproto.Bool, error) {
	c.Logger.Infof("messages.reportEncryptedSpam - user_id: %d, peer: %v",
		c.MD.UserId,
		in.GetPeer())

	return mtproto.BoolTrue, nil
}
