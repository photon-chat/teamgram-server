package core

import (
	"github.com/teamgram/proto/mtproto"
)

// MessagesReport
// messages.report#8953ab4e peer:InputPeer id:Vector<int> reason:ReportReason message:string = Bool;
func (c *ReportsCore) MessagesReport(in *mtproto.TLMessagesReport) (*mtproto.Bool, error) {
	c.Logger.Infof("messages.report - user_id: %d, peer: %v, ids: %v, reason: %s, message: %s",
		c.MD.UserId,
		in.GetPeer(),
		in.GetId(),
		in.GetReason().GetPredicateName(),
		in.GetMessage())

	return mtproto.BoolTrue, nil
}
