package grpc

import (
	"github.com/teamgram/proto/mtproto"
	"github.com/teamgram/teamgram-server/app/bff/reports/internal/server/grpc/service"
	"github.com/teamgram/teamgram-server/app/bff/reports/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

// New new a grpc server.
func New(ctx *svc.ServiceContext, c zrpc.RpcServerConf) *zrpc.RpcServer {
	s, err := zrpc.NewServer(c, func(grpcServer *grpc.Server) {
		mtproto.RegisterRPCReportsServer(grpcServer, service.New(ctx))
	})
	logx.Must(err)
	return s
}
