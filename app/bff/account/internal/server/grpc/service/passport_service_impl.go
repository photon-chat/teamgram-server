// Copyright 2024 Teamgram Authors
//  All rights reserved.
//
// Author: teamgramio (teamgram.io@gmail.com)
//

package service

import (
	"context"

	"github.com/teamgram/proto/mtproto"
	"github.com/teamgram/teamgram-server/app/bff/account/internal/core"
)

// AccountGetAuthorizations
// account.getAuthorizations#e320c158 = account.Authorizations;
func (s *Service) AccountGetAuthorizations(ctx context.Context, request *mtproto.TLAccountGetAuthorizations) (*mtproto.Account_Authorizations, error) {
	c := core.New(ctx, s.svcCtx)
	c.Logger.Debugf("account.getAuthorizations - metadata: %s, request: %s", c.MD.DebugString(), request.DebugString())

	r, err := c.AccountGetAuthorizations(request)
	if err != nil {
		return nil, err
	}

	c.Logger.Debugf("account.getAuthorizations - reply: %s", r.DebugString())
	return r, err
}

// AccountGetAllSecureValues
// account.getAllSecureValues#b288bc7d = Vector<SecureValue>;
func (s *Service) AccountGetAllSecureValues(ctx context.Context, request *mtproto.TLAccountGetAllSecureValues) (*mtproto.Vector_SecureValue, error) {
	return nil, mtproto.ErrMethodNotImpl
}

// AccountGetSecureValue
// account.getSecureValue#73665bc2 types:Vector<SecureValueType> = Vector<SecureValue>;
func (s *Service) AccountGetSecureValue(ctx context.Context, request *mtproto.TLAccountGetSecureValue) (*mtproto.Vector_SecureValue, error) {
	return nil, mtproto.ErrMethodNotImpl
}

// AccountSaveSecureValue
// account.saveSecureValue#899fe31d value:InputSecureValue secure_secret_id:long = SecureValue;
func (s *Service) AccountSaveSecureValue(ctx context.Context, request *mtproto.TLAccountSaveSecureValue) (*mtproto.SecureValue, error) {
	return nil, mtproto.ErrMethodNotImpl
}

// AccountDeleteSecureValue
// account.deleteSecureValue#b880bc4b types:Vector<SecureValueType> = Bool;
func (s *Service) AccountDeleteSecureValue(ctx context.Context, request *mtproto.TLAccountDeleteSecureValue) (*mtproto.Bool, error) {
	return nil, mtproto.ErrMethodNotImpl
}

// AccountGetAuthorizationForm
// account.getAuthorizationForm#a929597a bot_id:long scope:string public_key:string = account.AuthorizationForm;
func (s *Service) AccountGetAuthorizationForm(ctx context.Context, request *mtproto.TLAccountGetAuthorizationForm) (*mtproto.Account_AuthorizationForm, error) {
	return nil, mtproto.ErrMethodNotImpl
}

// AccountAcceptAuthorization
// account.acceptAuthorization#f3ed4c73 bot_id:long scope:string public_key:string value_hashes:Vector<SecureValueHash> credentials:SecureCredentialsEncrypted = Bool;
func (s *Service) AccountAcceptAuthorization(ctx context.Context, request *mtproto.TLAccountAcceptAuthorization) (*mtproto.Bool, error) {
	return nil, mtproto.ErrMethodNotImpl
}

// AccountSendVerifyPhoneCode
// account.sendVerifyPhoneCode#a5a356f9 phone_number:string settings:CodeSettings = auth.SentCode;
func (s *Service) AccountSendVerifyPhoneCode(ctx context.Context, request *mtproto.TLAccountSendVerifyPhoneCode) (*mtproto.Auth_SentCode, error) {
	return nil, mtproto.ErrMethodNotImpl
}

// AccountVerifyPhone
// account.verifyPhone#4dd3a7f6 phone_number:string phone_code_hash:string phone_code:string = Bool;
func (s *Service) AccountVerifyPhone(ctx context.Context, request *mtproto.TLAccountVerifyPhone) (*mtproto.Bool, error) {
	return nil, mtproto.ErrMethodNotImpl
}

// UsersSetSecureValueErrors
// users.setSecureValueErrors#90c894b5 id:InputUser errors:Vector<SecureValueError> = Bool;
func (s *Service) UsersSetSecureValueErrors(ctx context.Context, request *mtproto.TLUsersSetSecureValueErrors) (*mtproto.Bool, error) {
	return nil, mtproto.ErrMethodNotImpl
}

// HelpGetPassportConfig
// help.getPassportConfig#c661ad08 hash:int = help.PassportConfig;
func (s *Service) HelpGetPassportConfig(ctx context.Context, request *mtproto.TLHelpGetPassportConfig) (*mtproto.Help_PassportConfig, error) {
	return nil, mtproto.ErrMethodNotImpl
}
