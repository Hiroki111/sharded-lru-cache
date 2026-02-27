# sharded-lru-cache

## Context & Usage
- The Users: Other backend services (e.g., a microservice needing to cache database results or API responses).
- The Use Case: High-throughput systems where a single global lock would become a bottleneck (e.g., a session store or a metadata cache).
- Scalability: 
  - Throughput: Designed for 100k+ requests per second.
  - Data Volume: Typically gigabytes of RAM.
  - Payloads: Small to medium objects (JSON blobs, protobufs).

## Key Features
- Go Generics: Type-agnostic implementation supporting any comparable key and any value type.
- Sharded Architecture: Uses a Manager pattern to divide the cache into $N$ independent shards, reducing lock contention and improving performance on multi-core systems.
- Consistent Hashing: Implements a Hash Ring with Virtual Nodes to ensure uniform data distribution across shards.
- Durability (AOF): Append-Only File persistence with bufio buffering and background synchronization to ensure data survives server restarts.
- LRU Eviction: O(1) Least Recently Used eviction policy using a doubly linked list and a hash map.
- Active Janitor: A background goroutine that periodically sweeps and reaps expired items to prevent memory bloat.

## Architecture

### The Hash Ring
To avoid "Cache Stampedes" and ensure high availability, the system uses consistent hashing. By mapping shards to a circular hash space, we minimize the amount of data remapping required if the shard count changes.

### Persistence Layer
The AOF (Append-Only File) utilizes JSON + Base64 serialization. This allows complex structs to be stored as single, safe strings on disk, preventing file corruption from special characters or newlines in the data.

### Memory Management
The cache employs a dual-eviction strategy:
1. Lazy Eviction: Items are checked for expiration during access (Get).
2. Active Eviction: A "Janitor" goroutine runs at configurable intervals to clean up "zombie" data that hasn't been accessed.


## Usage

```
# Run the server
go run cmd/cache-server/main.go

# Open another terminal

# Set a value
curl -X POST http://localhost:8080/set \
     -d '{"key": "golang", "value": "is awesome", "ttl": 10}'

# Get the value
curl "http://localhost:8080/get?key=golang"

# Get stats
curl "http://localhost:8080/stats"

# Run all tests
go test ./internal/...

# Run benchmarks to see sharding performance
go test -bench=. ./internal/shard/
```

## Benchmark

LRU Cache
```
go test -bench=. ./internal/lru/
goos: linux
goarch: amd64
pkg: github.com/Hiroki111/sharded-lru-cache/internal/lru
cpu: 11th Gen Intel(R) Core(TM) i7-1165G7 @ 2.80GHz
BenchmarkLRU_Set-4      13053570                92.61 ns/op
```

Sharded LRU Cache - 1 shard
```
goos: linux
goarch: amd64
pkg: github.com/Hiroki111/sharded-lru-cache/internal/shard
cpu: 11th Gen Intel(R) Core(TM) i7-1165G7 @ 2.80GHz
BenchmarkShardedCache_Parallel-4         2758726               459.9 ns/op
```

Sharded LRU Cache - 32 shards
```
goos: linux
goarch: amd64
pkg: github.com/Hiroki111/sharded-lru-cache/internal/shard
cpu: 11th Gen Intel(R) Core(TM) i7-1165G7 @ 2.80GHz
BenchmarkShardedCache_Parallel-4         5563995               219.2 ns/op
```

Sharded LRU Cache - 1024 shards
```
goos: linux
goarch: amd64
pkg: github.com/Hiroki111/sharded-lru-cache/internal/shard
cpu: 11th Gen Intel(R) Core(TM) i7-1165G7 @ 2.80GHz
BenchmarkShardedCache_Parallel-4         5569131               239.1 ns/op
```

## Engineering Trade-offs
- AOF vs Snapshots: I chose AOF for higher durability. While it results in larger files, it ensures that every write is captured.
- Buffered Writing: Used bufio.Writer to turn expensive Disk I/O into cheap RAM-to-RAM copies, flushing to disk asynchronously to maintain high throughput.
- Base64 Encoding: Implemented to guarantee that the line-delimited AOF format remains robust even when storing binary data or complex JSON objects.

## Future Enhancement Ideas
- Custom Serialization: Supporting gob or protobuf for cross-network compatibility.
- Distributed Layer: Adding a gRPC or HTTP interface to turn it into a standalone service.
- Log Compaction: To keep the AOF file from growing infinitely.
- Raft/Paxos: To make it a truly distributed cluster across multiple machines.
- Prometheus Metrics: For professional-grade monitoring.
