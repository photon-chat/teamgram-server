package dataobject

type StickerSetsDO struct {
	Id           int64  `db:"id"`
	SetId        int64  `db:"set_id"`
	AccessHash   int64  `db:"access_hash"`
	ShortName    string `db:"short_name"`
	Title        string `db:"title"`
	StickerType  string `db:"sticker_type"`
	IsAnimated   bool   `db:"is_animated"`
	IsVideo      bool   `db:"is_video"`
	IsMasks      bool   `db:"is_masks"`
	IsEmojis     bool   `db:"is_emojis"`
	IsOfficial   bool   `db:"is_official"`
	StickerCount int32  `db:"sticker_count"`
	Hash         int32  `db:"hash"`
	ThumbDocId   int64  `db:"thumb_doc_id"`
	DataJson     string `db:"data_json"`
	FetchedAt    int64  `db:"fetched_at"`
}
