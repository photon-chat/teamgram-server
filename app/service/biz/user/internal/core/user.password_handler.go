package core

import (
	"github.com/teamgram/teamgram-server/app/service/biz/user/user_password"
)

// UserSavePassword 保存用户密码哈希
func (c *UserCore) UserSavePassword(in *user_password.SaveUserPasswordReq) (*user_password.SaveUserPasswordResp, error) {
	err := c.svcCtx.Dao.SaveUserPassword(c.ctx, in.UserId, in.PasswordHash)
	if err != nil {
		c.Logger.Errorf("user.saveUserPassword - error: %v", err)
		return nil, err
	}
	return &user_password.SaveUserPasswordResp{}, nil
}

// UserGetPassword 获取用户密码哈希
func (c *UserCore) UserGetPassword(in *user_password.GetUserPasswordReq) (*user_password.GetUserPasswordResp, error) {
	hash, err := c.svcCtx.Dao.GetUserPassword(c.ctx, in.UserId)
	if err != nil {
		c.Logger.Errorf("user.getUserPassword - error: %v", err)
		return nil, err
	}
	return &user_password.GetUserPasswordResp{PasswordHash: hash}, nil
}

// UserGetPasswordByPhone 通过手机号获取 user_id + 密码哈希
func (c *UserCore) UserGetPasswordByPhone(in *user_password.GetUserPasswordByPhoneReq) (*user_password.GetUserPasswordByPhoneResp, error) {
	userId, hash, err := c.svcCtx.Dao.GetUserPasswordByPhone(c.ctx, in.Phone)
	if err != nil {
		c.Logger.Errorf("user.getUserPasswordByPhone - error: %v", err)
		return nil, err
	}
	return &user_password.GetUserPasswordByPhoneResp{UserId: userId, PasswordHash: hash}, nil
}
