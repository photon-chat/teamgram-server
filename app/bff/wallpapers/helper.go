package wallpapers_helper

import (
	"github.com/teamgram/teamgram-server/app/bff/wallpapers/internal/config"
	"github.com/teamgram/teamgram-server/app/bff/wallpapers/internal/server/grpc/service"
	"github.com/teamgram/teamgram-server/app/bff/wallpapers/internal/svc"
)

type (
	Config = config.Config
)

func New(c Config) *service.Service {
	return service.New(svc.NewServiceContext(c))
}
