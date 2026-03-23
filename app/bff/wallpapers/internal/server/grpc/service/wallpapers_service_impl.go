package service

import (
	"context"

	"github.com/teamgram/proto/mtproto"
	"github.com/teamgram/teamgram-server/app/bff/wallpapers/internal/core"
)

// AccountGetWallPapers
func (s *Service) AccountGetWallPapers(ctx context.Context, request *mtproto.TLAccountGetWallPapers) (*mtproto.Account_WallPapers, error) {
	c := core.New(ctx, s.svcCtx)
	c.Logger.Debugf("account.getWallPapers - metadata: %s, request: %s", c.MD.DebugString(), request.DebugString())

	r, err := c.AccountGetWallPapers(request)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// AccountGetWallPaper
func (s *Service) AccountGetWallPaper(ctx context.Context, request *mtproto.TLAccountGetWallPaper) (*mtproto.WallPaper, error) {
	c := core.New(ctx, s.svcCtx)
	c.Logger.Debugf("account.getWallPaper - metadata: %s, request: %s", c.MD.DebugString(), request.DebugString())

	r, err := c.AccountGetWallPaper(request)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// AccountUploadWallPaper
func (s *Service) AccountUploadWallPaper(ctx context.Context, request *mtproto.TLAccountUploadWallPaper) (*mtproto.WallPaper, error) {
	c := core.New(ctx, s.svcCtx)
	c.Logger.Debugf("account.uploadWallPaper - metadata: %s, request: %s", c.MD.DebugString(), request.DebugString())

	r, err := c.AccountUploadWallPaper(request)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// AccountSaveWallPaper
func (s *Service) AccountSaveWallPaper(ctx context.Context, request *mtproto.TLAccountSaveWallPaper) (*mtproto.Bool, error) {
	c := core.New(ctx, s.svcCtx)
	c.Logger.Debugf("account.saveWallPaper - metadata: %s, request: %s", c.MD.DebugString(), request.DebugString())

	r, err := c.AccountSaveWallPaper(request)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// AccountInstallWallPaper
func (s *Service) AccountInstallWallPaper(ctx context.Context, request *mtproto.TLAccountInstallWallPaper) (*mtproto.Bool, error) {
	c := core.New(ctx, s.svcCtx)
	c.Logger.Debugf("account.installWallPaper - metadata: %s, request: %s", c.MD.DebugString(), request.DebugString())

	r, err := c.AccountInstallWallPaper(request)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// AccountResetWallPapers
func (s *Service) AccountResetWallPapers(ctx context.Context, request *mtproto.TLAccountResetWallPapers) (*mtproto.Bool, error) {
	c := core.New(ctx, s.svcCtx)
	c.Logger.Debugf("account.resetWallPapers - metadata: %s, request: %s", c.MD.DebugString(), request.DebugString())

	r, err := c.AccountResetWallPapers(request)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// AccountGetMultiWallPapers
func (s *Service) AccountGetMultiWallPapers(ctx context.Context, request *mtproto.TLAccountGetMultiWallPapers) (*mtproto.Vector_WallPaper, error) {
	c := core.New(ctx, s.svcCtx)
	c.Logger.Debugf("account.getMultiWallPapers - metadata: %s, request: %s", c.MD.DebugString(), request.DebugString())

	r, err := c.AccountGetMultiWallPapers(request)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// MessagesSetChatWallPaper
func (s *Service) MessagesSetChatWallPaper(ctx context.Context, request *mtproto.TLMessagesSetChatWallPaper) (*mtproto.Updates, error) {
	c := core.New(ctx, s.svcCtx)
	c.Logger.Debugf("messages.setChatWallPaper - metadata: %s, request: %s", c.MD.DebugString(), request.DebugString())

	r, err := c.MessagesSetChatWallPaper(request)
	if err != nil {
		return nil, err
	}

	return r, nil
}
