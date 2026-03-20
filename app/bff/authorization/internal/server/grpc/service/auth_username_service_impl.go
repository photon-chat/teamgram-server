package service

import (
	"context"

	"github.com/teamgram/teamgram-server/app/bff/authorization/auth_username"
	"github.com/teamgram/teamgram-server/app/bff/authorization/internal/core"
	"github.com/teamgram/teamgram-server/app/bff/authorization/internal/svc"
)

type AuthUsernameService struct {
	auth_username.UnimplementedAuthUsernameServiceServer
	svcCtx *svc.ServiceContext
}

func NewAuthUsernameService(svcCtx *svc.ServiceContext) *AuthUsernameService {
	return &AuthUsernameService{svcCtx: svcCtx}
}

func (s *AuthUsernameService) GetAuthMethods(ctx context.Context, req *auth_username.GetAuthMethodsReq) (*auth_username.GetAuthMethodsResp, error) {
	return &auth_username.GetAuthMethodsResp{
		AuthMethods: s.svcCtx.Config.GetAuthMethods(),
	}, nil
}

func (s *AuthUsernameService) CheckUsernameAvailable(ctx context.Context, req *auth_username.CheckUsernameAvailableReq) (*auth_username.CheckUsernameAvailableResp, error) {
	c := core.NewAuthUsernameCore(ctx, s.svcCtx)
	c.Logger.Debugf("auth_username.checkUsernameAvailable - request: %s", req)
	r, err := c.CheckUsernameAvailable(req)
	c.Logger.Debugf("auth_username.checkUsernameAvailable - reply: %s", r)
	return r, err
}

func (s *AuthUsernameService) UsernameRegister(ctx context.Context, req *auth_username.UsernameRegisterReq) (*auth_username.AuthResp, error) {
	c := core.NewAuthUsernameCore(ctx, s.svcCtx)
	c.Logger.Debugf("auth_username.usernameRegister - request: %s", req)
	r, err := c.UsernameRegister(req)
	c.Logger.Debugf("auth_username.usernameRegister - reply: %s", r)
	return r, err
}

func (s *AuthUsernameService) UsernameSignIn(ctx context.Context, req *auth_username.UsernameSignInReq) (*auth_username.AuthResp, error) {
	c := core.NewAuthUsernameCore(ctx, s.svcCtx)
	c.Logger.Debugf("auth_username.usernameSignIn - request: %s", req)
	r, err := c.UsernameSignIn(req)
	c.Logger.Debugf("auth_username.usernameSignIn - reply: %s", r)
	return r, err
}

func (s *AuthUsernameService) PhonePasswordRegister(ctx context.Context, req *auth_username.PhonePasswordRegisterReq) (*auth_username.AuthResp, error) {
	c := core.NewAuthUsernameCore(ctx, s.svcCtx)
	c.Logger.Debugf("auth_username.phonePasswordRegister - request: %s", req)
	r, err := c.PhonePasswordRegister(req)
	c.Logger.Debugf("auth_username.phonePasswordRegister - reply: %s", r)
	return r, err
}

func (s *AuthUsernameService) PhonePasswordSignIn(ctx context.Context, req *auth_username.PhonePasswordSignInReq) (*auth_username.AuthResp, error) {
	c := core.NewAuthUsernameCore(ctx, s.svcCtx)
	c.Logger.Debugf("auth_username.phonePasswordSignIn - request: %s", req)
	r, err := c.PhonePasswordSignIn(req)
	c.Logger.Debugf("auth_username.phonePasswordSignIn - reply: %s", r)
	return r, err
}
