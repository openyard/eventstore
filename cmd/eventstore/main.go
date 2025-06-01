package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/openyard/eventstore/internal/app/eventstore/domain"
	"github.com/openyard/eventstore/internal/app/eventstore/edge"
	"github.com/openyard/eventstore/internal/app/persistance"
	"github.com/openyard/eventstore/pkg/genproto/grpcapi"
	"github.com/openyard/eventstore/pkg/persistence/memkv"
	"github.com/openyard/eventstore/pkg/persistence/pgkv"
	"github.com/openyard/eventstore/pkg/persistence/pgkv/x"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	name     = "eventstore"
	vendor   = "Openyard ES³"
	version  = "TRUNK"
	revision = "HEAD"
)

var (
	Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		_, _ = fmt.Fprintf(os.Stderr, "  start eventstore\n")
		flag.PrintDefaults()
	}
	buckets = []string{persistance.KvsBucketIndex, persistance.KvsBucketContent}
)

func main() {
	log.SetPrefix(fmt.Sprintf("[%s/%s] ", name, version))
	log.SetFlags(log.Ldate | log.Lmicroseconds)
	log.SetOutput(os.Stdout)

	help := flag.Bool("h", false, "print usage")
	port := flag.String("l", ":2006", "tcp port to listen on")
	printv := flag.Bool("v", false, "print version")

	flag.Parse()

	if *help {
		Usage()
		os.Exit(0)
	}

	if *printv {
		fmt.Printf("%s %s - Rev. %s (by %s)\n", name, version, revision, vendor)
		os.Exit(0)
	}

	hf, qf := initSvc(GetEnv("DB", "in-memory"))
	t := edge.NewGrpcTransport(edge.WithHandleFunc(hf), edge.WithQueryFunc(qf))

	grpcSrv := grpc.NewServer()
	grpcapi.RegisterEventStoreServer(grpcSrv, t)
	grpcapi.RegisterTransportServer(grpcSrv, t)
	// Register reflection service on gRPC server.
	reflection.Register(grpcSrv)

	lis, err := net.Listen("tcp", *port)
	if err != nil {
		log.Fatalf("[ERROR] couldn't start listener: %v", err)
	}
	log.Panic(grpcSrv.Serve(lis))
}

func initSvc(db string) (domain.HandleFunc, domain.QueryFunc) {
	switch db {
	case "pg":
		s := domain.NewKVSService(domain.WithKeyValueStore(pgkv.NewPostgresKVS(openDB())))
		return s.HandleFunc, s.QueryFunc
	case "pgx":
		s := domain.NewKVSXService(domain.WithKeyValueStoreX(x.NewPostgresKVSX(openDB())))
		return s.HandleFunc, s.QueryFunc
	case "in-memory":
		fallthrough
	default:
		s := domain.NewKVSService(domain.WithKeyValueStore(memkv.NewMemoryKVS(buckets...)))
		return s.HandleFunc, s.QueryFunc
	}
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
