package dataobject

type UserInstalledStickerSetsDO struct {
	Id            int64 `db:"id"`
	UserId        int64 `db:"user_id"`
	SetId         int64 `db:"set_id"`
	SetType       int32 `db:"set_type"`
	OrderNum      int32 `db:"order_num"`
	InstalledDate int64 `db:"installed_date"`
	Archived      bool  `db:"archived"`
	Deleted       bool  `db:"deleted"`
}
