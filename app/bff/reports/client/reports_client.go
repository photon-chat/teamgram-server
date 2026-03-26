package reports_client

import (
	"context"

	"github.com/teamgram/proto/mtproto"

	"github.com/zeromicro/go-zero/zrpc"
)

var _ *mtproto.Bool

type ReportsClient interface {
	AccountReportPeer(ctx context.Context, in *mtproto.TLAccountReportPeer) (*mtproto.Bool, error)
	AccountReportProfilePhoto(ctx context.Context, in *mtproto.TLAccountReportProfilePhoto) (*mtproto.Bool, error)
	MessagesReportSpam(ctx context.Context, in *mtproto.TLMessagesReportSpam) (*mtproto.Bool, error)
	MessagesReport(ctx context.Context, in *mtproto.TLMessagesReport) (*mtproto.Bool, error)
	MessagesReportEncryptedSpam(ctx context.Context, in *mtproto.TLMessagesReportEncryptedSpam) (*mtproto.Bool, error)
	ChannelsReportSpam(ctx context.Context, in *mtproto.TLChannelsReportSpam) (*mtproto.Bool, error)
}

type defaultReportsClient struct {
	cli zrpc.Client
}

func NewReportsClient(cli zrpc.Client) ReportsClient {
	return &defaultReportsClient{
		cli: cli,
	}
}

func (m *defaultReportsClient) AccountReportPeer(ctx context.Context, in *mtproto.TLAccountReportPeer) (*mtproto.Bool, error) {
	client := mtproto.NewRPCReportsClient(m.cli.Conn())
	return client.AccountReportPeer(ctx, in)
}

func (m *defaultReportsClient) AccountReportProfilePhoto(ctx context.Context, in *mtproto.TLAccountReportProfilePhoto) (*mtproto.Bool, error) {
	client := mtproto.NewRPCReportsClient(m.cli.Conn())
	return client.AccountReportProfilePhoto(ctx, in)
}

func (m *defaultReportsClient) MessagesReportSpam(ctx context.Context, in *mtproto.TLMessagesReportSpam) (*mtproto.Bool, error) {
	client := mtproto.NewRPCReportsClient(m.cli.Conn())
	return client.MessagesReportSpam(ctx, in)
}

func (m *defaultReportsClient) MessagesReport(ctx context.Context, in *mtproto.TLMessagesReport) (*mtproto.Bool, error) {
	client := mtproto.NewRPCReportsClient(m.cli.Conn())
	return client.MessagesReport(ctx, in)
}

func (m *defaultReportsClient) MessagesReportEncryptedSpam(ctx context.Context, in *mtproto.TLMessagesReportEncryptedSpam) (*mtproto.Bool, error) {
	client := mtproto.NewRPCReportsClient(m.cli.Conn())
	return client.MessagesReportEncryptedSpam(ctx, in)
}

func (m *defaultReportsClient) ChannelsReportSpam(ctx context.Context, in *mtproto.TLChannelsReportSpam) (*mtproto.Bool, error) {
	client := mtproto.NewRPCReportsClient(m.cli.Conn())
	return client.ChannelsReportSpam(ctx, in)
}
