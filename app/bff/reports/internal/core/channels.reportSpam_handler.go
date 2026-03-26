package core

import (
	"github.com/teamgram/proto/mtproto"
)

// ChannelsReportSpam
// channels.reportSpam#f44a8315 channel:InputChannel participant:InputPeer id:Vector<int> = Bool;
func (c *ReportsCore) ChannelsReportSpam(in *mtproto.TLChannelsReportSpam) (*mtproto.Bool, error) {
	c.Logger.Infof("channels.reportSpam - user_id: %d, channel: %v, participant: %v, ids: %v",
		c.MD.UserId,
		in.GetChannel(),
		in.GetParticipant(),
		in.GetId())

	return mtproto.BoolTrue, nil
}
