package shard

import (
	"hash/fnv"
	"sort"
	"strconv"
)

type HashRing struct {
	shards              []uint32
	mapHashToShardIndex map[uint32]int
}

func NewHashRing(shardCount int, replicas int) *HashRing {
	hr := &HashRing{
		shards:              make([]uint32, 0),
		mapHashToShardIndex: make(map[uint32]int),
	}
	for i := 0; i < shardCount; i++ {
		for r := 0; r < replicas; r++ {
			nodeName := "shard-" + strconv.Itoa(i) + "-v" + strconv.Itoa(r)
			h := hr.hash(nodeName)

			hr.shards = append(hr.shards, h)
			hr.mapHashToShardIndex[h] = i
		}
	}
	sort.Slice(hr.shards, func(i, j int) bool {
		return hr.shards[i] < hr.shards[j]
	})
	return hr
}

func (hr *HashRing) GetShardIndex(key string, shardCount int) int {
	if len(hr.shards) == 0 {
		return 0
	}

	hash := hr.hash(key)
	// Find the first shard hash >= key hash (binary search)
	shardIndex := sort.Search(len(hr.shards), func(i int) bool {
		return hr.shards[i] >= hash
	})

	if shardIndex == len(hr.shards) {
		shardIndex = 0
	}

	return hr.mapHashToShardIndex[hr.shards[shardIndex]]
}

func (hr *HashRing) hash(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}
