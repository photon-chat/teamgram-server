package core

import (
	"github.com/gogo/protobuf/types"
	"github.com/teamgram/proto/mtproto"
)

func int32Value(v int32) *types.Int32Value {
	return &types.Int32Value{Value: v}
}

// makeColorWallpaper creates a solid color wallPaperNoFile
func makeColorWallpaper(id int64, color int32, isDefault bool) *mtproto.WallPaper {
	return mtproto.MakeTLWallPaperNoFile(&mtproto.WallPaper{
		Id:      id,
		Default: isDefault,
		Settings: mtproto.MakeTLWallPaperSettings(&mtproto.WallPaperSettings{
			BackgroundColor: int32Value(color),
		}).To_WallPaperSettings(),
	}).To_WallPaper()
}

// makeGradientWallpaper creates a gradient wallPaperNoFile
func makeGradientWallpaper(id int64, colors []int32, rotation int32, isDefault bool) *mtproto.WallPaper {
	settings := &mtproto.WallPaperSettings{
		BackgroundColor: int32Value(colors[0]),
		Rotation:        int32Value(rotation),
	}
	if len(colors) > 1 {
		settings.SecondBackgroundColor = int32Value(colors[1])
	}
	if len(colors) > 2 {
		settings.ThirdBackgroundColor = int32Value(colors[2])
	}
	if len(colors) > 3 {
		settings.FourthBackgroundColor = int32Value(colors[3])
	}
	return mtproto.MakeTLWallPaperNoFile(&mtproto.WallPaper{
		Id:       id,
		Default:  isDefault,
		Settings: mtproto.MakeTLWallPaperSettings(settings).To_WallPaperSettings(),
	}).To_WallPaper()
}

// defaultWallpapers returns a set of built-in color and gradient wallpapers
func defaultWallpapers() []*mtproto.WallPaper {
	return []*mtproto.WallPaper{
		// Solid colors
		makeColorWallpaper(10001, 0xffffff, true), // White (default)
		makeColorWallpaper(10002, 0xd4e7ed, false),
		makeColorWallpaper(10003, 0xd6e2ea, false),
		makeColorWallpaper(10004, 0xe8d0b8, false),
		makeColorWallpaper(10005, 0xd5e1c5, false),
		makeColorWallpaper(10006, 0xc2d6e6, false),
		makeColorWallpaper(10007, 0xe7d5c0, false),
		makeColorWallpaper(10008, 0xdadce0, false),
		makeColorWallpaper(10009, 0x1b2836, false), // Dark

		// Two-color gradients
		makeGradientWallpaper(10101, []int32{0xdbddbb, 0x6ba587}, 0, false),
		makeGradientWallpaper(10102, []int32{0x7fa381, 0xfff5c3}, 0, false),
		makeGradientWallpaper(10103, []int32{0xfec496, 0xdd6cb9}, 0, false),
		makeGradientWallpaper(10104, []int32{0x7ec4ea, 0xc6b1ef}, 0, false),
		makeGradientWallpaper(10105, []int32{0x22191c, 0x392d2e}, 0, false), // Dark gradient

		// Four-color gradients
		makeGradientWallpaper(10201, []int32{0xdbddbb, 0x6ba587, 0xd5d88d, 0x88b884}, 0, false),
		makeGradientWallpaper(10202, []int32{0xeaa36e, 0xf0e486, 0x7ec4ea, 0xc6b1ef}, 0, false),
		makeGradientWallpaper(10203, []int32{0xffc6a0, 0xffdf9a, 0xb2e5a2, 0xa2c1e5}, 0, false),
		makeGradientWallpaper(10204, []int32{0x598bf6, 0x7a5eef, 0xd67cff, 0xf38b58}, 0, false), // Dark
	}
}
