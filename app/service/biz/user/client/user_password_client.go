package user_client

import (
	"context"

	"github.com/teamgram/teamgram-server/app/service/biz/user/user_password"

	"github.com/zeromicro/go-zero/zrpc"
)

type UserPasswordClient interface {
	SaveUserPassword(ctx context.Context, in *user_password.SaveUserPasswordReq) (*user_password.SaveUserPasswordResp, error)
	GetUserPassword(ctx context.Context, in *user_password.GetUserPasswordReq) (*user_password.GetUserPasswordResp, error)
	GetUserPasswordByPhone(ctx context.Context, in *user_password.GetUserPasswordByPhoneReq) (*user_password.GetUserPasswordByPhoneResp, error)
}

type defaultUserPasswordClient struct {
	cli zrpc.Client
}

func NewUserPasswordClient(cli zrpc.Client) UserPasswordClient {
	return &defaultUserPasswordClient{cli: cli}
}

func (m *defaultUserPasswordClient) SaveUserPassword(ctx context.Context, in *user_password.SaveUserPasswordReq) (*user_password.SaveUserPasswordResp, error) {
	client := user_password.NewUserPasswordServiceClient(m.cli.Conn())
	return client.SaveUserPassword(ctx, in)
}

func (m *defaultUserPasswordClient) GetUserPassword(ctx context.Context, in *user_password.GetUserPasswordReq) (*user_password.GetUserPasswordResp, error) {
	client := user_password.NewUserPasswordServiceClient(m.cli.Conn())
	return client.GetUserPassword(ctx, in)
}

func (m *defaultUserPasswordClient) GetUserPasswordByPhone(ctx context.Context, in *user_password.GetUserPasswordByPhoneReq) (*user_password.GetUserPasswordByPhoneResp, error) {
	client := user_password.NewUserPasswordServiceClient(m.cli.Conn())
	return client.GetUserPasswordByPhone(ctx, in)
}
