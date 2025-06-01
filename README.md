# EventStore

---

Store streams and events via grpc

### Build

```bash
make clean      # clean - remove generated sources and built binaries
make tools      # install tools (protoc-gen*) and update dependencies
make compile    # generate sources
make build      # build binary
make build-all  # build cross-platform binaries
make build-spec # build test-executable
make package    # build docker image
make test       # run unit-test
make verify     # run test-executable
make spec       # run acceptance tests
make sim        # run simulator
```

### Run

... as single-instance with in-memory key-value storage  
`./bin/eventstore`

or as distributed service with postgres key-value storage incl. batch-processing
```bash
docker run -d --name testdb \
  -e POSTGRES_USER=testuser \
  -e POSTGRES_DB=testdb \
  -e POSTGRES_PASSWORD=11111 \
  -p 5432:5432 postgres:latest

DB=pgx ./bin/eventstore
```

### Env

| Key       | Value/Example (* = default) | Description                                           |
|-----------|-----------------------------|-------------------------------------------------------|
| DB        | `in-memory`*                | In-Memory Key-Value                                   |
|           | `pg`                        | Postgres Key-Value                                    |
|           | `pgx`                       | Postgres experimental Key-Value with batch processing |
| DB_HOST   | `localhost`*                |                                                       |
| DB_PORT   | `5432`*                     |                                                       |
| DB_NAME   | `testdb`*                   |                                                       |
| DB_USER   | `testdb`*                   |                                                       |
| DB_PASS   | `11111`*                    |                                                       |
| DB_PARAMS | `?sslmode=disable`*         |                                                       |

### Benchmark

Measure runtime for appending 1000 events (100 streams with 10 events each)

| Storage Type                     | Duration                                                                  |
|----------------------------------|---------------------------------------------------------------------------|
| In Memory Key-Value              | 2025/05/23 13:40:02 [ACCESS]	 edge.GrpcTransport.Append took 32.291066ms  |
| Postgres Key-Value               | 2025/05/23 13:36:31 [ACCESS]	 edge.GrpcTransport.Append took 397.122135ms |
| Postgres Key-Value X             | 2025/05/23 19:18:53 [ACCESS]	 edge.GrpcTransport.Append took 36.433729ms  |
| Postgres Relational              | TODO                                                                      |
| File based Key-Value             | TODO                                                                      |
| File based /w io_uring Key-Value | TODO                                                                      |