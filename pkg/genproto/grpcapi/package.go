//go:generate protoc --go_out=. --go-grpc_out=. --proto_path=../../../api/grpc -I=../../../api/proto -I=../../../api/third_party --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative eventstore.proto service.proto
package grpcapi
