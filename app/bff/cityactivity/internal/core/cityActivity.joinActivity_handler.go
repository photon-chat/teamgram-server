package core

import (
	"github.com/teamgram/proto/mtproto"
	chatpb "github.com/teamgram/teamgram-server/app/service/biz/chat/chat"
)

func (c *CityActivityCore) CityActivityJoinActivity(in *mtproto.TLCityActivityJoinActivity) (*mtproto.Bool, error) {
	if c.MD == nil {
		return mtproto.BoolFalse, mtproto.ErrInternelServerError
	}

	city := in.GetCity()
	if city == "" && c.MD.ClientAddr != "" {
		city = c.svcCtx.Dao.GetCityByIp(c.MD.ClientAddr)
	}

	err := c.svcCtx.Dao.JoinActivity(c.ctx, in.GetId(), c.MD.UserId, city)
	if err != nil {
		c.Logger.Errorf("cityActivity.joinActivity - error: %v", err)
		return mtproto.BoolFalse, err
	}

	// Auto-add user to activity group chat
	activity, err := c.svcCtx.Dao.GetActivityById(c.ctx, in.GetId())
	if err == nil && activity != nil && activity.ChatId > 0 {
		_, err2 := c.svcCtx.Dao.ChatAddChatUser(c.ctx, &chatpb.TLChatAddChatUser{
			ChatId:    activity.ChatId,
			InviterId: 0,
			UserId:    c.MD.UserId,
			IsBot:     false,
		})
		if err2 != nil {
			c.Logger.Errorf("cityActivity.joinActivity - add to chat error: %v", err2)
		}
	}

	return mtproto.BoolTrue, nil
}
