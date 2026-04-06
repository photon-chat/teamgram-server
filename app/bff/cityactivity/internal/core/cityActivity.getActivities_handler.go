package core

import (
	"github.com/teamgram/proto/mtproto"
	"github.com/teamgram/teamgram-server/app/bff/cityactivity/internal/dao"
	userpb "github.com/teamgram/teamgram-server/app/service/biz/user/user"
	media "github.com/teamgram/teamgram-server/app/service/media/media"
)

func (c *CityActivityCore) CityActivityGetActivities(in *mtproto.TLCityActivityGetActivities) (*mtproto.CityActivity_Activities, error) {
	// 客户端传的城市直接用于过滤，空城市 = 查询全部活动
	// IP检测城市只在创建活动时使用，不在列表查询时使用
	city := in.GetCity()

	offset := in.GetOffset()
	limit := in.GetLimit()
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	filter := in.GetFilter()
	q := in.GetQ()
	c.Logger.Infof("cityActivity.getActivities - city: %s, offset: %d, limit: %d, filter: %d, q: %s, constructor: %d, full: %+v", city, offset, limit, filter, q, in.GetConstructor(), in)

	activities, count, err := c.svcCtx.Dao.GetActivitiesByCity(c.ctx, city, offset, limit, filter, q)
	if err != nil {
		c.Logger.Errorf("cityActivity.getActivities - error: %v", err)
		return nil, err
	}

	var userId int64
	if c.MD != nil {
		userId = c.MD.UserId
	}

	// Batch get first photo for each activity
	activityIds := make([]int64, 0, len(activities))
	for _, a := range activities {
		activityIds = append(activityIds, a.Id)
	}
	firstPhotoIds, _ := c.svcCtx.Dao.GetActivitiesFirstPhotoIds(c.ctx, activityIds)

	// Batch resolve creator names
	creatorIds := make([]int64, 0, len(activities))
	for _, a := range activities {
		creatorIds = append(creatorIds, a.UserId)
	}
	creatorNames := make(map[int64]string)
	if len(creatorIds) > 0 {
		if userData, err2 := c.svcCtx.Dao.UserClient.UserGetUserDataListByIdList(c.ctx, &userpb.TLUserGetUserDataListByIdList{
			UserIdList: creatorIds,
		}); err2 == nil {
			for _, ud := range userData.GetDatas() {
				name := ud.GetFirstName()
				if ln := ud.GetLastName(); ln != "" {
					name += " " + ln
				}
				creatorNames[ud.GetId()] = name
			}
		}
	}
	for _, a := range activities {
		if name, ok := creatorNames[a.UserId]; ok {
			a.CreatorName = name
		}
	}

	// Resolve photos via MediaClient
	photoCache := make(map[int64]*mtproto.Photo)
	for _, pid := range firstPhotoIds {
		if _, ok := photoCache[pid]; !ok {
			photo, err2 := c.svcCtx.Dao.MediaGetPhoto(c.ctx, &media.TLMediaGetPhoto{PhotoId: pid})
			if err2 == nil && photo != nil {
				photoCache[pid] = photo
			}
		}
	}

	result := mtproto.MakeTLCityActivityActivities(&mtproto.CityActivity_Activities{
		Count:      count,
		Activities: make([]*mtproto.CityActivity, 0, len(activities)),
	})

	for _, a := range activities {
		isJoined := false
		if userId > 0 {
			isJoined = c.svcCtx.Dao.IsUserJoined(c.ctx, a.Id, userId)
		}
		proto := activityToProto(a, isJoined)
		// Attach first photo if available
		if pid, ok := firstPhotoIds[a.Id]; ok {
			if photo, ok2 := photoCache[pid]; ok2 {
				proto.Photos = []*mtproto.Photo{photo}
			}
		}
		result.Data2.Activities = append(result.Data2.Activities, proto)
	}

	return result.To_CityActivity_Activities(), nil
}

func activityToProto(a *dao.Activity, isJoined bool) *mtproto.CityActivity {
	return mtproto.MakeTLCityActivity(&mtproto.CityActivity{
		Id:               a.Id,
		UserId:           a.UserId,
		Title:            a.Title,
		Description:      a.Description,
		PhotoId:          a.PhotoId,
		City:             a.City,
		StartTime:        a.StartTime,
		EndTime:          a.EndTime,
		MaxParticipants:  a.MaxParticipants,
		Status:           a.Status,
		IsGlobal:         mtproto.ToBool(a.IsGlobal == 1),
		ParticipantCount: a.ParticipantCount,
		IsJoined:         mtproto.ToBool(isJoined),
		CreatorName:      a.CreatorName,
		CreatedAt:        a.CreatedAt,
		ChatId:           a.ChatId,
	}).To_CityActivity()
}
