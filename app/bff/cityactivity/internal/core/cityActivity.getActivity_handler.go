package core

import (
	"github.com/teamgram/proto/mtproto"
)

func (c *CityActivityCore) CityActivityGetActivity(in *mtproto.TLCityActivityGetActivity) (*mtproto.CityActivity, error) {
	activity, err := c.svcCtx.Dao.GetActivityById(c.ctx, in.GetId())
	if err != nil {
		c.Logger.Errorf("cityActivity.getActivity - error: %v", err)
		return nil, mtproto.ErrInternelServerError
	}
	if activity == nil {
		return nil, mtproto.ErrInternelServerError
	}

	var userId int64
	if c.MD != nil {
		userId = c.MD.UserId
	}
	isJoined := false
	if userId > 0 {
		isJoined = c.svcCtx.Dao.IsUserJoined(c.ctx, activity.Id, userId)
	}

	return activityToProto(activity, isJoined), nil
}
