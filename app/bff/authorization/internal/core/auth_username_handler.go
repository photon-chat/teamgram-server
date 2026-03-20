package core

import (
	"context"

	"github.com/teamgram/proto/mtproto"
	"github.com/teamgram/proto/mtproto/crypto"
	"github.com/teamgram/proto/mtproto/rpc/metadata"
	"github.com/teamgram/teamgram-server/app/bff/authorization/auth_username"
	"github.com/teamgram/teamgram-server/app/bff/authorization/internal/svc"
	"github.com/teamgram/teamgram-server/app/service/authsession/authsession"
	"github.com/teamgram/teamgram-server/app/service/biz/user/user_password"
	userpb "github.com/teamgram/teamgram-server/app/service/biz/user/user"
	usernamepb "github.com/teamgram/teamgram-server/app/service/biz/username/username"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type AuthUsernameCore struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	MD *metadata.RpcMetadata
}

func NewAuthUsernameCore(ctx context.Context, svcCtx *svc.ServiceContext) *AuthUsernameCore {
	return &AuthUsernameCore{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
		MD:     metadata.RpcMetadataFromIncoming(ctx),
	}
}

// createAuthKeyAndBind 创建 MTProto auth_key 并绑定到 user_id（公共逻辑）
func (c *AuthUsernameCore) createAuthKeyAndBind(userId int64) *crypto.AuthKey {
	key := crypto.CreateAuthKey()
	c.svcCtx.Dao.AuthsessionClient.AuthsessionSetAuthKey(c.ctx, &authsession.TLAuthsessionSetAuthKey{
		AuthKey: &mtproto.AuthKeyInfo{
			AuthKeyId:          key.AuthKeyId(),
			AuthKey:            key.AuthKey(),
			AuthKeyType:        mtproto.AuthKeyTypePerm,
			PermAuthKeyId:      key.AuthKeyId(),
			TempAuthKeyId:      0,
			MediaTempAuthKeyId: 0,
		},
	})
	c.svcCtx.Dao.AuthsessionClient.AuthsessionBindAuthKeyUser(c.ctx, &authsession.TLAuthsessionBindAuthKeyUser{
		AuthKeyId: key.AuthKeyId(),
		UserId:    userId,
	})
	return key
}

// CheckUsernameAvailable 检查用户名是否可用
func (c *AuthUsernameCore) CheckUsernameAvailable(in *auth_username.CheckUsernameAvailableReq) (*auth_username.CheckUsernameAvailableResp, error) {
	if len(in.Username) < 3 || len(in.Username) > 32 {
		return &auth_username.CheckUsernameAvailableResp{Available: false}, nil
	}

	checkResult, err := c.svcCtx.Dao.UsernameClient.UsernameCheckUsername(c.ctx, &usernamepb.TLUsernameCheckUsername{
		Username: in.Username,
	})
	if err != nil {
		c.Logger.Errorf("check username error: %v", err)
		return nil, err
	}

	available := checkResult.PredicateName == "usernameNotExisted"
	return &auth_username.CheckUsernameAvailableResp{Available: available}, nil
}

// UsernameRegister 用户名+密码注册
func (c *AuthUsernameCore) UsernameRegister(in *auth_username.UsernameRegisterReq) (*auth_username.AuthResp, error) {
	// 1. 验证参数
	if len(in.Username) < 3 || len(in.Username) > 32 {
		c.Logger.Errorf("invalid username: %s", in.Username)
		return nil, mtproto.ErrUsernameInvalid
	}
	if len(in.Password) < 6 || len(in.Password) > 128 {
		c.Logger.Errorf("invalid password length")
		return nil, mtproto.ErrPasswordHashInvalid
	}

	// 2. 检查用户名是否已存在
	checkResult, err := c.svcCtx.Dao.UsernameClient.UsernameCheckUsername(c.ctx, &usernamepb.TLUsernameCheckUsername{
		Username: in.Username,
	})
	if err != nil {
		c.Logger.Errorf("check username error: %v", err)
		return nil, err
	}
	if checkResult.PredicateName != "usernameNotExisted" {
		c.Logger.Errorf("username already occupied: %s", in.Username)
		return nil, mtproto.ErrUsernameOccupied
	}

	// 3. 生成密码哈希
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		c.Logger.Errorf("generate password hash error: %v", err)
		return nil, mtproto.ErrInternelServerError
	}

	// 4. 创建用户
	user, err := c.svcCtx.Dao.UserClient.UserCreateNewUser(c.ctx, &userpb.TLUserCreateNewUser{
		Phone:     "",
		FirstName: in.FirstName,
		LastName:  in.LastName,
	})
	if err != nil {
		c.Logger.Errorf("create user error: %v", err)
		return nil, err
	}

	// 5. 保存密码
	_, err = c.svcCtx.Dao.UserPasswordClient.SaveUserPassword(c.ctx, &user_password.SaveUserPasswordReq{
		UserId:       user.Id(),
		PasswordHash: string(passwordHash),
	})
	if err != nil {
		c.Logger.Errorf("save password error: %v", err)
		return nil, err
	}

	// 6. 设置用户名
	c.svcCtx.Dao.UserClient.UserUpdateUsername(c.ctx, &userpb.TLUserUpdateUsername{
		UserId:   user.Id(),
		Username: in.Username,
	})
	c.svcCtx.Dao.UsernameClient.UsernameUpdateUsername(c.ctx, &usernamepb.TLUsernameUpdateUsername{
		PeerType: mtproto.PEER_USER,
		PeerId:   user.Id(),
		Username: in.Username,
	})

	// 7. 创建 auth_key 并绑定
	key := c.createAuthKeyAndBind(user.Id())

	c.Logger.Infof("user registered: id=%d, username=%s", user.Id(), in.Username)

	return &auth_username.AuthResp{
		UserId:    user.Id(),
		Username:  in.Username,
		FirstName: in.FirstName,
		LastName:  in.LastName,
		AuthKey:   key.AuthKey(),
		AuthKeyId: key.AuthKeyId(),
	}, nil
}

// UsernameSignIn 用户名+密码登录
func (c *AuthUsernameCore) UsernameSignIn(in *auth_username.UsernameSignInReq) (*auth_username.AuthResp, error) {
	if in.Username == "" || in.Password == "" {
		return nil, mtproto.ErrInputRequestInvalid
	}

	// 1. 通过用户名获取 user_id
	peer, err := c.svcCtx.Dao.UsernameClient.UsernameResolveUsername(c.ctx, &usernamepb.TLUsernameResolveUsername{
		Username: in.Username,
	})
	if err != nil || peer == nil {
		c.Logger.Errorf("username not found: %s", in.Username)
		return nil, mtproto.ErrUsernameNotOccupied
	}

	userId := peer.UserId
	if userId == 0 {
		c.Logger.Errorf("invalid user_id for username: %s", in.Username)
		return nil, mtproto.ErrUsernameNotOccupied
	}

	// 2. 获取密码哈希
	passwordResp, err := c.svcCtx.Dao.UserPasswordClient.GetUserPassword(c.ctx, &user_password.GetUserPasswordReq{
		UserId: userId,
	})
	if err != nil || passwordResp.PasswordHash == "" {
		c.Logger.Errorf("password not found for user: %d", userId)
		return nil, mtproto.ErrPasswordHashInvalid
	}

	// 3. 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(passwordResp.PasswordHash), []byte(in.Password))
	if err != nil {
		c.Logger.Errorf("password mismatch for user: %d", userId)
		return nil, mtproto.ErrPasswordHashInvalid
	}

	// 4. 获取用户信息
	user, err := c.svcCtx.Dao.UserClient.UserGetImmutableUser(c.ctx, &userpb.TLUserGetImmutableUser{
		Id: userId,
	})
	if err != nil || user == nil {
		c.Logger.Errorf("get user info error: %v", err)
		return nil, mtproto.ErrInternelServerError
	}

	// 5. 创建 auth_key 并绑定
	key := c.createAuthKeyAndBind(userId)

	c.Logger.Infof("user signed in: id=%d, username=%s", userId, in.Username)

	return &auth_username.AuthResp{
		UserId:    userId,
		Username:  in.Username,
		FirstName: user.User.FirstName,
		LastName:  user.User.LastName,
		AuthKey:   key.AuthKey(),
		AuthKeyId: key.AuthKeyId(),
	}, nil
}

// PhonePasswordRegister 手机号+密码注册
func (c *AuthUsernameCore) PhonePasswordRegister(in *auth_username.PhonePasswordRegisterReq) (*auth_username.AuthResp, error) {
	// 1. 验证参数
	if in.Phone == "" {
		c.Logger.Errorf("invalid phone")
		return nil, mtproto.ErrPhoneNumberInvalid
	}
	if len(in.Password) < 6 || len(in.Password) > 128 {
		c.Logger.Errorf("invalid password length")
		return nil, mtproto.ErrPasswordHashInvalid
	}

	// 2. 生成密码哈希
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		c.Logger.Errorf("generate password hash error: %v", err)
		return nil, mtproto.ErrInternelServerError
	}

	// 3. 创建用户（带手机号）
	user, err := c.svcCtx.Dao.UserClient.UserCreateNewUser(c.ctx, &userpb.TLUserCreateNewUser{
		Phone:     in.Phone,
		FirstName: in.FirstName,
		LastName:  in.LastName,
	})
	if err != nil {
		c.Logger.Errorf("create user error: %v", err)
		return nil, err
	}

	// 4. 保存密码
	_, err = c.svcCtx.Dao.UserPasswordClient.SaveUserPassword(c.ctx, &user_password.SaveUserPasswordReq{
		UserId:       user.Id(),
		PasswordHash: string(passwordHash),
	})
	if err != nil {
		c.Logger.Errorf("save password error: %v", err)
		return nil, err
	}

	// 5. 创建 auth_key 并绑定
	key := c.createAuthKeyAndBind(user.Id())

	c.Logger.Infof("user registered with phone: id=%d, phone=%s", user.Id(), in.Phone)

	return &auth_username.AuthResp{
		UserId:    user.Id(),
		Phone:     in.Phone,
		FirstName: in.FirstName,
		LastName:  in.LastName,
		AuthKey:   key.AuthKey(),
		AuthKeyId: key.AuthKeyId(),
	}, nil
}

// PhonePasswordSignIn 手机号+密码登录
func (c *AuthUsernameCore) PhonePasswordSignIn(in *auth_username.PhonePasswordSignInReq) (*auth_username.AuthResp, error) {
	if in.Phone == "" || in.Password == "" {
		return nil, mtproto.ErrInputRequestInvalid
	}

	// 1. 通过手机号获取 user_id + 密码哈希
	passwordResp, err := c.svcCtx.Dao.UserPasswordClient.GetUserPasswordByPhone(c.ctx, &user_password.GetUserPasswordByPhoneReq{
		Phone: in.Phone,
	})
	if err != nil || passwordResp.UserId == 0 {
		c.Logger.Errorf("phone not found or no password: %s", in.Phone)
		return nil, mtproto.ErrPhoneNumberUnoccupied
	}

	// 2. 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(passwordResp.PasswordHash), []byte(in.Password))
	if err != nil {
		c.Logger.Errorf("password mismatch for phone: %s", in.Phone)
		return nil, mtproto.ErrPasswordHashInvalid
	}

	userId := passwordResp.UserId

	// 3. 获取用户信息
	user, err := c.svcCtx.Dao.UserClient.UserGetImmutableUser(c.ctx, &userpb.TLUserGetImmutableUser{
		Id: userId,
	})
	if err != nil || user == nil {
		c.Logger.Errorf("get user info error: %v", err)
		return nil, mtproto.ErrInternelServerError
	}

	// 4. 创建 auth_key 并绑定
	key := c.createAuthKeyAndBind(userId)

	c.Logger.Infof("user signed in with phone: id=%d, phone=%s", userId, in.Phone)

	return &auth_username.AuthResp{
		UserId:    userId,
		Phone:     in.Phone,
		FirstName: user.User.FirstName,
		LastName:  user.User.LastName,
		AuthKey:   key.AuthKey(),
		AuthKeyId: key.AuthKeyId(),
	}, nil
}
