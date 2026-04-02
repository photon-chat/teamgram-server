package core

import (
	"github.com/teamgram/proto/mtproto"
)

func (c *CityActivityCore) CityActivityLeaveActivity(in *mtproto.TLCityActivityLeaveActivity) (*mtproto.Bool, error) {
	if c.MD == nil {
		return mtproto.BoolFalse, mtproto.ErrInternelServerError
	}
	err := c.svcCtx.Dao.LeaveActivity(c.ctx, in.GetId(), c.MD.UserId)
	if err != nil {
		c.Logger.Errorf("cityActivity.leaveActivity - error: %v", err)
		return mtproto.BoolFalse, err
	}
	return mtproto.BoolTrue, nil
}
