package dataobject

type StickerSetDocumentsDO struct {
	Id              int64  `db:"id"`
	SetId           int64  `db:"set_id"`
	DocumentId      int64  `db:"document_id"`
	StickerIndex    int32  `db:"sticker_index"`
	Emoji           string `db:"emoji"`
	BotFileId       string `db:"bot_file_id"`
	BotFileUniqueId string `db:"bot_file_unique_id"`
	BotThumbFileId  string `db:"bot_thumb_file_id"`
	DocumentData    string `db:"document_data"`
	FileDownloaded  bool   `db:"file_downloaded"`
}
