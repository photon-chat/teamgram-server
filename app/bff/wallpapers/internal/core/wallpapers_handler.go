package core

import (
	"github.com/teamgram/proto/mtproto"
)

// AccountGetWallPapers - returns default color/gradient wallpapers
func (c *WallpapersCore) AccountGetWallPapers(in *mtproto.TLAccountGetWallPapers) (*mtproto.Account_WallPapers, error) {
	wallpapers := defaultWallpapers()

	return mtproto.MakeTLAccountWallPapers(&mtproto.Account_WallPapers{
		Hash:       int64(len(wallpapers)),
		Wallpapers: wallpapers,
	}).To_Account_WallPapers(), nil
}

// AccountGetWallPaper
func (c *WallpapersCore) AccountGetWallPaper(in *mtproto.TLAccountGetWallPaper) (*mtproto.WallPaper, error) {
	return nil, mtproto.ErrWallpaperInvalid
}

// AccountUploadWallPaper
func (c *WallpapersCore) AccountUploadWallPaper(in *mtproto.TLAccountUploadWallPaper) (*mtproto.WallPaper, error) {
	return nil, mtproto.ErrWallpaperInvalid
}

// AccountSaveWallPaper
func (c *WallpapersCore) AccountSaveWallPaper(in *mtproto.TLAccountSaveWallPaper) (*mtproto.Bool, error) {
	return mtproto.BoolTrue, nil
}

// AccountInstallWallPaper
func (c *WallpapersCore) AccountInstallWallPaper(in *mtproto.TLAccountInstallWallPaper) (*mtproto.Bool, error) {
	return mtproto.BoolTrue, nil
}

// AccountResetWallPapers
func (c *WallpapersCore) AccountResetWallPapers(in *mtproto.TLAccountResetWallPapers) (*mtproto.Bool, error) {
	return mtproto.BoolTrue, nil
}

// AccountGetMultiWallPapers
func (c *WallpapersCore) AccountGetMultiWallPapers(in *mtproto.TLAccountGetMultiWallPapers) (*mtproto.Vector_WallPaper, error) {
	return &mtproto.Vector_WallPaper{
		Datas: []*mtproto.WallPaper{},
	}, nil
}
