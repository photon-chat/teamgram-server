package core

import (
	"github.com/teamgram/proto/mtproto"
)

// MessagesReportSpam
// messages.reportSpam#cf1592db peer:InputPeer = Bool;
func (c *ReportsCore) MessagesReportSpam(in *mtproto.TLMessagesReportSpam) (*mtproto.Bool, error) {
	c.Logger.Infof("messages.reportSpam - user_id: %d, peer: %v",
		c.MD.UserId,
		in.GetPeer())

	return mtproto.BoolTrue, nil
}
