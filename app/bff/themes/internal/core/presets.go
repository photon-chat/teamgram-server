package core

import (
	"github.com/gogo/protobuf/types"
	"github.com/teamgram/proto/mtproto"
)

func stringValue(v string) *types.StringValue {
	return &types.StringValue{Value: v}
}

// makeChatTheme creates a chat theme with light and dark settings
func makeChatTheme(id int64, emoticon string, lightAccent int32, lightMsgColors []int32, darkAccent int32, darkMsgColors []int32) *mtproto.Theme {
	settings := []*mtproto.ThemeSettings{
		// Light (Day) settings
		mtproto.MakeTLThemeSettings(&mtproto.ThemeSettings{
			BaseTheme:     mtproto.MakeTLBaseThemeDay(&mtproto.BaseTheme{}).To_BaseTheme(),
			AccentColor:   lightAccent,
			MessageColors: lightMsgColors,
		}).To_ThemeSettings(),
		// Dark settings
		mtproto.MakeTLThemeSettings(&mtproto.ThemeSettings{
			BaseTheme:     mtproto.MakeTLBaseThemeNight(&mtproto.BaseTheme{}).To_BaseTheme(),
			AccentColor:   darkAccent,
			MessageColors: darkMsgColors,
		}).To_ThemeSettings(),
	}

	return mtproto.MakeTLTheme(&mtproto.Theme{
		Id:         id,
		AccessHash: id * 31,
		Slug:       emoticon,
		Title:      emoticon,
		ForChat:    true,
		Emoticon:   stringValue(emoticon),
		Settings:   settings,
	}).To_Theme()
}

// defaultChatThemes returns predefined emoticon-based chat themes
func defaultChatThemes() []*mtproto.Theme {
	return []*mtproto.Theme{
		makeChatTheme(20001, "❤", 0xf5524a, []int32{0xf57f87, 0xf5d1c3}, 0xf5524a, []int32{0x8b3a3a, 0x5c2626}),
		makeChatTheme(20002, "🏠", 0x7e5836, []int32{0xdbc3a0, 0xc9b289}, 0xc69b6d, []int32{0x4a3527, 0x3a2a1f}),
		makeChatTheme(20003, "🌿", 0x5da352, []int32{0xa5d68c, 0xc6e6a9}, 0x5da352, []int32{0x2e4a2a, 0x1f3a1b}),
		makeChatTheme(20004, "☀️", 0xe29e35, []int32{0xf5d97e, 0xf5e6ab}, 0xe29e35, []int32{0x5c4a1f, 0x4a3a15}),
		makeChatTheme(20005, "🍊", 0xe07e39, []int32{0xf5b87e, 0xf5d4a8}, 0xe07e39, []int32{0x5c3a1f, 0x4a2f15}),
		makeChatTheme(20006, "🌊", 0x3d8ed1, []int32{0x7ec4ea, 0xa8d8f0}, 0x3d8ed1, []int32{0x1f3a5c, 0x152f4a}),
		makeChatTheme(20007, "🌸", 0xd45a8f, []int32{0xf5a0c0, 0xf5c6d6}, 0xd45a8f, []int32{0x5c1f3a, 0x4a152f}),
		makeChatTheme(20008, "💜", 0x7b5ebd, []int32{0xc6b1ef, 0xd5c6f5}, 0x7b5ebd, []int32{0x3a2a5c, 0x2f1f4a}),
		makeChatTheme(20009, "🎮", 0x4a8e3f, []int32{0x85c77c, 0xaad6a2}, 0x4a8e3f, []int32{0x2a4a26, 0x1f3a1b}),
		makeChatTheme(20010, "🎄", 0xc74040, []int32{0xd47676, 0x72a550}, 0xc74040, []int32{0x5c2626, 0x2e4a2a}),
	}
}
