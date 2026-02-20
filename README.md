# sharded-lru-cache

## Context & Usage
- The Users: Other backend services (e.g., a microservice needing to cache database results or API responses).
- The Use Case: High-throughput systems where a single global lock would become a bottleneck (e.g., a session store or a metadata cache).
- Scalability: 
  - Throughput: Designed for 100k+ requests per second.
  - Data Volume: Typically gigabytes of RAM.
  - Payloads: Small to medium objects (JSON blobs, protobufs).

## Minimum Viable Product (MVP)
1. LRU Logic: A doubly linked list 🔗 combined with a hash map 🗺️ for $O(1)$ access and eviction.
2. Sharding Strategy: A hashing function (like fnv64a) to map keys to specific shards.
3. Concurrency Control: Using sync.RWMutex per shard to allow concurrent reads.
4. Basic API: Get(key), Set(key, value, TTL), and Delete(key).

## Roadmap
1. Implement the Doubly Linked List and Hash Map manually (don't just use container/list). This is where you conquer your fear of pointers and memory allocation.
2. Benchmarking and Profiling. Use go test -bench and pprof to generate flame graphs. See exactly how much time the Garbage Collector spends cleaning up your evicted nodes.
3. Add the Sharding logic. Implement a hashing algorithm (like FNV-1a) to distribute keys.
4. Benchmarking and Profilin again.

## Optional Enhancements
- TTL (Time-to-Live): Automatically expiring keys after a duration.
- Prometheus Metrics: Tracking hit/miss ratios and eviction counts 📊.
- Custom Serialization: Supporting gob or protobuf for cross-network compatibility.
- Distributed Layer: Adding a gRPC or HTTP interface to turn it into a standalone service.

```
/sharded-cache
├── cmd/
│   └── cache-server/    # The main application entry point
│       └── main.go      # Compiles to 'cache-server' binary
├── internal/            # Private code (not importable by other projects)
│   ├── lru/             # Core eviction logic
│   └── shard/           # Concurrency and sharding management
├── pkg/                 # Public library code (if you want others to use your cache)
├── go.mod
└── README.md
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
