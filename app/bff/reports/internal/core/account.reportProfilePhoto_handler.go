package core

import (
	"github.com/teamgram/proto/mtproto"
)

// AccountReportProfilePhoto
// account.reportProfilePhoto#fa8cc6f5 peer:InputPeer photo_id:InputPhoto reason:ReportReason message:string = Bool;
func (c *ReportsCore) AccountReportProfilePhoto(in *mtproto.TLAccountReportProfilePhoto) (*mtproto.Bool, error) {
	c.Logger.Infof("account.reportProfilePhoto - user_id: %d, peer: %v, photo_id: %v, reason: %s, message: %s",
		c.MD.UserId,
		in.GetPeer(),
		in.GetPhotoId(),
		in.GetReason().GetPredicateName(),
		in.GetMessage())

	return mtproto.BoolTrue, nil
}
