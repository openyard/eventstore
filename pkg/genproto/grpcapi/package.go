//go:generate protoc --go_out=. --go-grpc_out=. --proto_path=../../../api/proto --proto_path=../../../api/grpc --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative eventstore.proto service.proto
package grpcapi
