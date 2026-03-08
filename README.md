# sharded-lru-cache

A high-performance, distributed-ready in-memory cache written in Go, featuring AOF persistence, binary-safe transparency, and automated log compaction.

## Context & Usage
- The Users: Other backend services (e.g., a microservice needing to cache database results or API responses).
- The Use Case: High-throughput systems where a single global lock would become a bottleneck (e.g., a session store or a metadata cache).
- Scalability: 
  - Throughput: Designed for 100k+ requests per second.
  - Data Volume: Typically gigabytes of RAM.
  - Payloads: Small to medium objects (JSON blobs, protobufs).

## Key Features
- Sharded Architecture: Uses a custom Hash Ring to distribute keys across multiple LRU shards, minimizing mutex contention for high-concurrency workloads.
- Binary-Safe Persistence: Implements an Append-Only File (AOF) that stores raw bytes, avoiding the common JSON float64 precision loss.
- Log Compaction: Background process to rewrite the AOF, keeping the disk footprint minimal by removing expired or overwritten keys.
- Production-Ready Client: A Go SDK supporting generic GetAs[T] types for seamless struct unmarshaling.
- Dockerized: Multi-stage builds for a tiny, portable footprint.

## Tech Stack
- Language: Go 1.25+
- Storage: In-memory LRU with `sync.RWMutex` sharding.
- API: RESTful JSON/Base64.
- Deployment: Docker / Docker Hub.

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

### Run with Docker
```
docker pull hiroki111/sharded-lru-cache:v1.0.0
docker run -p 8080:8080 -v $(pwd)/data:/app/data hiroki111/sharded-lru-cache:v1.0.0
```

### Update Docker image
```
# Change "v1.0.0" to the actual version you use
docker build -t hiroki111/sharded-lru-cache:v1.0.0 .
docker push hiroki111/sharded-lru-cache:v1.0.0
```

### Use the Go Client
```
c := client.NewClient("http://localhost:8080")

// Store complex structs
user := User{ID: 1, Name: "Alice"}
c.Set("user:1", user, 10*time.Minute)

// Retrieve with full type safety
val, _ := client.GetAs[User](c, "user:1")
fmt.Println(val.Name) // Alice
```

### Run locally
```
# Run the server
go run cmd/cache-server/main.go

# Open another terminal

# Set a value
# QmF0bWFu is "Batman" in Base64
curl -s -X POST http://localhost:8080/set -d '{"key": "hero", "value": "QmF0bWFu", "ttl": 3600}'

# Get the value
curl "http://localhost:8080/get?key=hero"

# Get stats
curl "http://localhost:8080/stats"

# Run all tests
go test ./...

# Run benchmarks to see sharding performance
go test -bench=. ./pkg/shard/
```

## Design Decisions & Trade-offs
- Why []byte over interface{}? To ensure the server remains "type-blind," allowing it to store any data format without losing numerical precision during JSON unmarshaling, while maximizing CPU usage in the server side.
- Why HTTP over gRPC? For maximum compatibility with web-based microservices while keeping the implementation simple and debuggable via curl.
- AOF Recovery: On startup, the server re-scans the AOF to rebuild the memory state, ensuring data durability against process crashes.

## Future Enhancement Ideas
- Raft/Paxos: To make it a truly distributed cluster across multiple machines.
- Consistent Hashing with Multiple Nodes (Client-Side Sharding)
- Custom Serialization: Try protobuf to reduce the size of payload.
- Prometheus Metrics: For professional-grade monitoring.