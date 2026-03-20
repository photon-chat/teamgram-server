package service

import (
	"context"

	"github.com/teamgram/teamgram-server/app/service/biz/user/internal/core"
	"github.com/teamgram/teamgram-server/app/service/biz/user/internal/svc"
	"github.com/teamgram/teamgram-server/app/service/biz/user/user_password"
)

type UserPasswordService struct {
	user_password.UnimplementedUserPasswordServiceServer
	svcCtx *svc.ServiceContext
}

func NewUserPasswordService(svcCtx *svc.ServiceContext) *UserPasswordService {
	return &UserPasswordService{svcCtx: svcCtx}
}

func (s *UserPasswordService) SaveUserPassword(ctx context.Context, req *user_password.SaveUserPasswordReq) (*user_password.SaveUserPasswordResp, error) {
	c := core.New(ctx, s.svcCtx)
	c.Logger.Debugf("user_password.saveUserPassword - request: %v", req)
	r, err := c.UserSavePassword(req)
	c.Logger.Debugf("user_password.saveUserPassword - reply: %v", r)
	return r, err
}

func (s *UserPasswordService) GetUserPassword(ctx context.Context, req *user_password.GetUserPasswordReq) (*user_password.GetUserPasswordResp, error) {
	c := core.New(ctx, s.svcCtx)
	c.Logger.Debugf("user_password.getUserPassword - request: %v", req)
	r, err := c.UserGetPassword(req)
	c.Logger.Debugf("user_password.getUserPassword - reply: %v", r)
	return r, err
}

func (s *UserPasswordService) GetUserPasswordByPhone(ctx context.Context, req *user_password.GetUserPasswordByPhoneReq) (*user_password.GetUserPasswordByPhoneResp, error) {
	c := core.New(ctx, s.svcCtx)
	c.Logger.Debugf("user_password.getUserPasswordByPhone - request: %v", req)
	r, err := c.UserGetPasswordByPhone(req)
	c.Logger.Debugf("user_password.getUserPasswordByPhone - reply: %v", r)
	return r, err
}
