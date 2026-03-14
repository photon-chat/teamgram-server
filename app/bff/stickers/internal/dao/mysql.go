package dao

import (
	"github.com/teamgram/marmota/pkg/stores/sqlx"
	"github.com/teamgram/teamgram-server/app/bff/stickers/internal/dal/dao/mysql_dao"
)

type Mysql struct {
	*sqlx.DB
	*mysql_dao.StickerSetsDAO
	*mysql_dao.StickerSetDocumentsDAO
}

func newMysqlDao(db *sqlx.DB) *Mysql {
	return &Mysql{
		DB:                     db,
		StickerSetsDAO:         mysql_dao.NewStickerSetsDAO(db),
		StickerSetDocumentsDAO: mysql_dao.NewStickerSetDocumentsDAO(db),
	}
}
