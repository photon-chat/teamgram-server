package core

import (
	"github.com/teamgram/proto/mtproto"
)

// AccountReportPeer
// account.reportPeer#c5ba3d86 peer:InputPeer reason:ReportReason message:string = Bool;
func (c *ReportsCore) AccountReportPeer(in *mtproto.TLAccountReportPeer) (*mtproto.Bool, error) {
	c.Logger.Infof("account.reportPeer - user_id: %d, peer: %v, reason: %s, message: %s",
		c.MD.UserId,
		in.GetPeer(),
		in.GetReason().GetPredicateName(),
		in.GetMessage())

	return mtproto.BoolTrue, nil
}
