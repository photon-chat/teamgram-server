package core

import (
	"github.com/teamgram/proto/mtproto"
)

func (c *CityActivityCore) CityActivityDeleteActivity(in *mtproto.TLCityActivityDeleteActivity) (*mtproto.Bool, error) {
	if c.MD == nil {
		return mtproto.BoolFalse, mtproto.ErrInternelServerError
	}
	err := c.svcCtx.Dao.DeleteActivity(c.ctx, in.GetId(), c.MD.UserId)
	if err != nil {
		c.Logger.Errorf("cityActivity.deleteActivity - error: %v", err)
		return nil, err
	}
	return mtproto.BoolTrue, nil
}
