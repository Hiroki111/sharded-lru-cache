#!/bin/bash

# Configuration
AOF_PATH="./cache.aof"
SERVER_BIN="./cache-server"
PORT=8080

# 1. Clean up old state
rm -f $AOF_PATH
go build -o $SERVER_BIN cmd/cache-server/main.go

# 2. Start the server in the background
echo "🚀 Starting cache server..."
$SERVER_BIN &
SERVER_PID=$!

# Wait for server to boot
sleep 2

# 3. Inject data via API
echo "📥 Injecting data (Batman and Joker)..."
curl -s -X POST http://localhost:$PORT/set -d '{"key": "hero", "value": "Batman", "ttl": 3600}'
curl -s -X POST http://localhost:$PORT/set -d '{"key": "villain", "value": "Joker", "ttl": 3600}'

# Give the AOF Syncer a moment to flush the buffer to disk
echo "💾 Waiting for AOF Sync..."
sleep 2

# 4. Simulate the CRASH
echo "💥 CRASHING server now (SIGKILL)..."
kill -9 $SERVER_PID

# 5. Restart the server
echo "♻️ Restarting server (Recovery phase)..."
$SERVER_BIN &
NEW_SERVER_PID=$!
sleep 2

# 6. Verify data exists
echo "🔍 Querying recovered data..."
RESPONSE=$(curl -s "http://localhost:$PORT/get?key=hero")

if [[ $RESPONSE == *"Batman"* ]]; then
    echo "✅ SUCCESS: Batman recovered from AOF!"
else
    echo "❌ FAILURE: Data lost! Response was: $RESPONSE"
fi

# Cleanup
kill $NEW_SERVER_PID