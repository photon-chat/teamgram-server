package core

import (
	"github.com/teamgram/proto/mtproto"
	"github.com/teamgram/teamgram-server/app/service/authsession/authsession"
	userpb "github.com/teamgram/teamgram-server/app/service/biz/user/user"
	"github.com/teamgram/teamgram-server/app/service/biz/user/user_password"
	"golang.org/x/crypto/bcrypt"
)

// AuthPhonePasswordSignIn
// auth.phonePasswordSignIn phone:string password:string = Auth_Authorization;
func (c *AuthorizationCore) AuthPhonePasswordSignIn(in *mtproto.TLAuthPhonePasswordSignIn) (*mtproto.Auth_Authorization, error) {
	if in.Phone == "" || in.Password == "" {
		return nil, mtproto.ErrInputRequestInvalid
	}

	// 1. get user_id + password hash by phone
	passwordResp, err := c.svcCtx.Dao.UserPasswordClient.GetUserPasswordByPhone(c.ctx, &user_password.GetUserPasswordByPhoneReq{
		Phone: in.Phone,
	})
	if err != nil || passwordResp.UserId == 0 {
		c.Logger.Errorf("phone not found or no password: %s", in.Phone)
		return nil, mtproto.ErrPhoneNumberUnoccupied
	}

	// 2. verify password
	err = bcrypt.CompareHashAndPassword([]byte(passwordResp.PasswordHash), []byte(in.Password))
	if err != nil {
		c.Logger.Errorf("password mismatch for phone: %s", in.Phone)
		return nil, mtproto.ErrPasswordHashInvalid
	}

	userId := passwordResp.UserId

	// 3. get user info
	user, err := c.svcCtx.Dao.UserClient.UserGetImmutableUser(c.ctx, &userpb.TLUserGetImmutableUser{
		Id: userId,
	})
	if err != nil || user == nil {
		c.Logger.Errorf("get user info error: %v", err)
		return nil, mtproto.ErrInternelServerError
	}

	// 4. bind existing auth_key to user
	c.svcCtx.Dao.AuthsessionClient.AuthsessionBindAuthKeyUser(c.ctx, &authsession.TLAuthsessionBindAuthKeyUser{
		AuthKeyId: c.MD.AuthId,
		UserId:    userId,
	})

	c.Logger.Infof("user signed in with phone via MTProto: id=%d, phone=%s", userId, in.Phone)

	return mtproto.MakeTLAuthAuthorization(&mtproto.Auth_Authorization{
		User: user.ToSelfUser(),
	}).To_Auth_Authorization(), nil
}
