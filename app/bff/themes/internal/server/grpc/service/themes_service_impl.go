package service

import (
	"context"

	"github.com/teamgram/proto/mtproto"
	"github.com/teamgram/teamgram-server/app/bff/themes/internal/core"
)

// AccountUploadTheme
func (s *Service) AccountUploadTheme(ctx context.Context, request *mtproto.TLAccountUploadTheme) (*mtproto.Document, error) {
	c := core.New(ctx, s.svcCtx)
	c.Logger.Debugf("account.uploadTheme - metadata: %s, request: %s", c.MD.DebugString(), request.DebugString())

	r, err := c.AccountUploadTheme(request)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// AccountCreateTheme
func (s *Service) AccountCreateTheme(ctx context.Context, request *mtproto.TLAccountCreateTheme) (*mtproto.Theme, error) {
	c := core.New(ctx, s.svcCtx)
	c.Logger.Debugf("account.createTheme - metadata: %s, request: %s", c.MD.DebugString(), request.DebugString())

	r, err := c.AccountCreateTheme(request)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// AccountUpdateTheme
func (s *Service) AccountUpdateTheme(ctx context.Context, request *mtproto.TLAccountUpdateTheme) (*mtproto.Theme, error) {
	c := core.New(ctx, s.svcCtx)
	c.Logger.Debugf("account.updateTheme - metadata: %s, request: %s", c.MD.DebugString(), request.DebugString())

	r, err := c.AccountUpdateTheme(request)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// AccountSaveTheme
func (s *Service) AccountSaveTheme(ctx context.Context, request *mtproto.TLAccountSaveTheme) (*mtproto.Bool, error) {
	c := core.New(ctx, s.svcCtx)
	c.Logger.Debugf("account.saveTheme - metadata: %s, request: %s", c.MD.DebugString(), request.DebugString())

	r, err := c.AccountSaveTheme(request)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// AccountInstallTheme
func (s *Service) AccountInstallTheme(ctx context.Context, request *mtproto.TLAccountInstallTheme) (*mtproto.Bool, error) {
	c := core.New(ctx, s.svcCtx)
	c.Logger.Debugf("account.installTheme - metadata: %s, request: %s", c.MD.DebugString(), request.DebugString())

	r, err := c.AccountInstallTheme(request)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// AccountGetTheme
func (s *Service) AccountGetTheme(ctx context.Context, request *mtproto.TLAccountGetTheme) (*mtproto.Theme, error) {
	c := core.New(ctx, s.svcCtx)
	c.Logger.Debugf("account.getTheme - metadata: %s, request: %s", c.MD.DebugString(), request.DebugString())

	r, err := c.AccountGetTheme(request)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// AccountGetThemes
func (s *Service) AccountGetThemes(ctx context.Context, request *mtproto.TLAccountGetThemes) (*mtproto.Account_Themes, error) {
	c := core.New(ctx, s.svcCtx)
	c.Logger.Debugf("account.getThemes - metadata: %s, request: %s", c.MD.DebugString(), request.DebugString())

	r, err := c.AccountGetThemes(request)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// AccountGetChatThemes
func (s *Service) AccountGetChatThemes(ctx context.Context, request *mtproto.TLAccountGetChatThemes) (*mtproto.Account_Themes, error) {
	c := core.New(ctx, s.svcCtx)
	c.Logger.Debugf("account.getChatThemes - metadata: %s, request: %s", c.MD.DebugString(), request.DebugString())

	r, err := c.AccountGetChatThemes(request)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// MessagesSetChatTheme
func (s *Service) MessagesSetChatTheme(ctx context.Context, request *mtproto.TLMessagesSetChatTheme) (*mtproto.Updates, error) {
	c := core.New(ctx, s.svcCtx)
	c.Logger.Debugf("messages.setChatTheme - metadata: %s, request: %s", c.MD.DebugString(), request.DebugString())

	r, err := c.MessagesSetChatTheme(request)
	if err != nil {
		return nil, err
	}

	return r, nil
}
