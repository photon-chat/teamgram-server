package core

import (
	"github.com/teamgram/proto/mtproto"
	"github.com/teamgram/teamgram-server/app/bff/cityactivity/internal/dao"
	media "github.com/teamgram/teamgram-server/app/service/media/media"
)

func (c *CityActivityCore) CityActivityCreateActivity(in *mtproto.TLCityActivityCreateActivity) (*mtproto.CityActivity, error) {
	if c.MD == nil {
		return nil, mtproto.ErrInternelServerError
	}
	city := in.GetCity()
	if city == "" && c.MD.ClientAddr != "" {
		city = c.svcCtx.Dao.GetCityByIp(c.MD.ClientAddr)
	}

	a := &dao.Activity{
		UserId:          c.MD.UserId,
		Title:           in.GetTitle(),
		Description:     in.GetDescription(),
		PhotoId:         in.GetPhotoId(),
		City:            city,
		StartTime:       in.GetStartTime(),
		EndTime:         in.GetEndTime(),
		MaxParticipants: in.GetMaxParticipants(),
	}

	id, err := c.svcCtx.Dao.CreateActivity(c.ctx, a)
	if err != nil {
		c.Logger.Errorf("cityActivity.createActivity - error: %v", err)
		return nil, err
	}
	a.Id = id

	// Save activity media (photo_ids)
	photoIds := in.GetPhotoIds()
	if len(photoIds) > 5 {
		photoIds = photoIds[:5]
	}
	if len(photoIds) > 0 {
		if err2 := c.svcCtx.Dao.SaveActivityMedia(c.ctx, id, photoIds); err2 != nil {
			c.Logger.Errorf("cityActivity.createActivity - save media error: %v", err2)
		}
	}

	// Resolve photos for response
	var photos []*mtproto.Photo
	for _, pid := range photoIds {
		photo, err3 := c.svcCtx.Dao.MediaGetPhoto(c.ctx, &media.TLMediaGetPhoto{PhotoId: pid})
		if err3 == nil && photo != nil {
			photos = append(photos, photo)
		}
	}

	result := activityToProto(a, false)
	result.Photos = photos
	return result, nil
}
