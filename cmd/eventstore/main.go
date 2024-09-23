package main

import (
	"log"
	"net"

	"github.com/openyard/eventstore/internal/app/eventstore/domain"
	"github.com/openyard/eventstore/internal/app/eventstore/edge"
	"github.com/openyard/eventstore/internal/app/kvstore"
	"github.com/openyard/eventstore/pkg/genproto/grpcapi"
	"github.com/openyard/eventstore/pkg/kvstore/memkv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var buckets = []string{kvstore.KvsBucketIndex, kvstore.KvsBucketContent}

func main() {
	s := domain.NewService(domain.WithKeyValueStore(memkv.NewMemoryKVS(buckets...)))
	t := edge.NewGrpcTransport(edge.WithHandleFunc(s.HandleFunc), edge.WithQueryFunc(s.QueryFunc))

	grpcSrv := grpc.NewServer()
	grpcapi.RegisterEventStoreServer(grpcSrv, t)
	grpcapi.RegisterTransportServer(grpcSrv, t)
	// Register reflection service on gRPC server.
	reflection.Register(grpcSrv)

	lis, err := net.Listen("tcp", ":2006")
	if err != nil {
		log.Fatalf("[ERROR] couldn't start listener: %v", err)
	}
	log.Panic(grpcSrv.Serve(lis))
}
