package core

import (
	"github.com/teamgram/proto/mtproto"
)

func (c *CityActivityCore) CityActivityEditActivity(in *mtproto.TLCityActivityEditActivity) (*mtproto.CityActivity, error) {
	if c.MD == nil {
		return nil, mtproto.ErrInternelServerError
	}
	err := c.svcCtx.Dao.EditActivity(c.ctx, in.GetId(), c.MD.UserId,
		in.GetTitle(), in.GetDescription(), in.GetPhotoId(),
		in.GetStartTime(), in.GetEndTime(), in.GetStatus())
	if err != nil {
		c.Logger.Errorf("cityActivity.editActivity - error: %v", err)
		return nil, err
	}

	activity, err := c.svcCtx.Dao.GetActivityById(c.ctx, in.GetId())
	if err != nil {
		return nil, err
	}
	if activity == nil {
		return nil, mtproto.ErrInternelServerError
	}

	return activityToProto(activity, c.svcCtx.Dao.IsUserJoined(c.ctx, activity.Id, c.MD.UserId)), nil
}
