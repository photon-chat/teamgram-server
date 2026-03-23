package dao

import (
	"github.com/teamgram/teamgram-server/app/bff/wallpapers/internal/config"
)

type Dao struct {
}

func New(c config.Config) *Dao {
	return &Dao{}
}
