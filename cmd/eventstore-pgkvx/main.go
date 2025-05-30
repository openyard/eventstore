package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/openyard/eventstore/internal/app/eventstore/domain"
	"github.com/openyard/eventstore/internal/app/eventstore/edge"
	"github.com/openyard/eventstore/pkg/genproto/grpcapi"
	"github.com/openyard/eventstore/pkg/persistence/pgkv/x"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	s := domain.NewKVSXService(domain.WithKeyValueStoreX(x.NewPostgresKVSX(openDB())))
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

func openDB() *sql.DB {
	dbHost := GetEnv("DB_HOST", "localhost")
	dbPort := GetEnv("DB_PORT", "5432")
	dbName := GetEnv("DB_NAME", "testdb")
	dbUser := GetEnv("DB_USER", "testdb")
	dbPass := GetEnv("DB_PASS", "11111")
	dbParams := GetEnv("DB_PARAMS", "?sslmode=disable")

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s%s", dbUser, url.QueryEscape(dbPass), dbHost, dbPort, dbName, dbParams)
	db, err := sql.Open("pgx", connString)
	if err != nil {
		log.Printf("[ERROR] couldn't open database connection: %s", err)
	}

	for i := 0; i < 10; i++ {
		if err = db.Ping(); err != nil {
			log.Printf("[WARN] database ping failed: %s", err)
			time.Sleep(10 * time.Second)
		} else {
			break
		}
	}
	if err != nil {
		log.Fatalf("[ERROR] couldn't connect to database: %s", err)
	}
	return db
}

func GetEnv(key, defaultValue string) string {
	value, ok := os.LookupEnv(key)
	if ok && value != "" {
		return value
	}
	return defaultValue
}
