# EventStore

---

Store streams and events via grpc 


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