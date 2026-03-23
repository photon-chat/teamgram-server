package themes_helper

import (
	"github.com/teamgram/teamgram-server/app/bff/themes/internal/config"
	"github.com/teamgram/teamgram-server/app/bff/themes/internal/server/grpc/service"
	"github.com/teamgram/teamgram-server/app/bff/themes/internal/svc"
)

type (
	Config = config.Config
)

func New(c Config) *service.Service {
	return service.New(svc.NewServiceContext(c))
}
