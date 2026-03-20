package core

import (
	"github.com/teamgram/proto/mtproto"
)

// AuthGetAuthMethods
// auth.getAuthMethods = Auth_AuthMethods;
func (c *AuthorizationCore) AuthGetAuthMethods(in *mtproto.TLAuthGetAuthMethods) (*mtproto.Auth_AuthMethods, error) {
	methods := c.svcCtx.Config.GetAuthMethods()
	return &mtproto.Auth_AuthMethods{
		Constructor: mtproto.CRC32_auth_authMethods,
		Methods:     methods,
	}, nil
}
