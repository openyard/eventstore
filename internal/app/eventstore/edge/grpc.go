package edge

import (
	"context"
	"log"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/openyard/eventstore/internal/app/eventstore/domain"
	"github.com/openyard/eventstore/pkg/genproto/grpcapi"
)

var (
	_ grpcapi.EventStoreServer = (*GrpcTransport)(nil)
	_ grpcapi.TransportServer  = (*GrpcTransport)(nil)
)

type GrpcTransportOption func(transport *GrpcTransport)

type GrpcTransport struct {
	grpcapi.UnimplementedEventStoreServer
	grpcapi.UnimplementedTransportServer
	node   *snowflake.Node
	handle func(cmd domain.Command) error
	query  func(cmd domain.Command) ([]domain.Stream, error)
}

func NewGrpcTransport(opts ...GrpcTransportOption) *GrpcTransport {
	node, err := snowflake.NewNode(1)
	assertNoError(err)
	t := &GrpcTransport{
		node: node,
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

func WithHandleFunc(hf func(cmd domain.Command) error) GrpcTransportOption {
	return func(g *GrpcTransport) {
		g.handle = hf
	}
}

func WithQueryFunc(qf func(cmd domain.Command) ([]domain.Stream, error)) GrpcTransportOption {
	return func(g *GrpcTransport) {
		g.query = qf
	}
}

func (g GrpcTransport) Append(ctx context.Context, request *grpcapi.AppendRequest) (*grpcapi.Empty, error) {
	start := time.Now()
	defer func() { log.Printf("[ACCESS]\t %T.Append took %s", g, time.Since(start)) }()
	var streamData []domain.StreamData
	for _, s := range request.StreamData {
		streamData = append(streamData, domain.NewStreamData(s.Name, s.ExpectedVersion, api2domain(s.Events)...))

	}
	cmd := domain.Append(streamData...)
	return &grpcapi.Empty{}, g.handle(domain.NewCommand(ctx, domain.AppendCmd, cmd))
}

func (g GrpcTransport) Read(ctx context.Context, request *grpcapi.ReadRequest) (*grpcapi.Streams, error) {
	start := time.Now()
	defer func() { log.Printf("[ACCESS]\t %T.Read took %s", g, time.Since(start)) }()
	cmd := domain.Read(request.Streams...)
	streams, err := g.query(domain.NewCommand(ctx, domain.ReadCmd, cmd))
	return domainStream2ApiStream(streams, err)
}

func (g GrpcTransport) ReadAt(ctx context.Context, request *grpcapi.ReadAtRequest) (*grpcapi.Streams, error) {
	start := time.Now()
	defer func() { log.Printf("[ACCESS]\t %T.ReadAt took %s", g, time.Since(start)) }()
	cmd := domain.ReadAt(request.At.AsTime(), request.Streams...)
	streams, err := g.query(domain.NewCommand(ctx, domain.ReadAtCmd, cmd))
	return domainStream2ApiStream(streams, err)
}

func (g GrpcTransport) Subscribe(request *grpcapi.SubscriptionRequest, server grpcapi.Transport_SubscribeServer) error {
	//TODO implement me
	panic("implement me")
}

func (g GrpcTransport) SubscribeWithID(request *grpcapi.SubscriptionWithIDRequest, server grpcapi.Transport_SubscribeWithIDServer) error {
	//TODO implement me
	panic("implement me")
}

func (g GrpcTransport) SubscribeWithOffset(request *grpcapi.SubscriptionWithOffsetRequest, server grpcapi.Transport_SubscribeWithOffsetServer) error {
	//TODO implement me
	panic("implement me")
}

func domainStream2ApiStream(streams []domain.Stream, err error) (*grpcapi.Streams, error) {
	result := &grpcapi.Streams{Streams: make([]*grpcapi.Stream, 0)}
	for _, stream := range streams {
		result.Streams = append(result.Streams, domain2api(stream.Name(), stream.Events()))
	}
	return result, err
}

func assertNoError(err error) {
	if err != nil {
		log.Panic(err)
	}
}
