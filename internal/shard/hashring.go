package shard

import (
	"hash/fnv"
	"sort"
	"strconv"
)

type HashRing struct {
	hashes              []uint32
	mapHashToShardIndex map[uint32]int
}

func NewHashRing(shardCount int, replicas int) *HashRing {
	hr := &HashRing{
		hashes:              make([]uint32, 0),
		mapHashToShardIndex: make(map[uint32]int),
	}
	for i := 0; i < shardCount; i++ {
		for r := 0; r < replicas; r++ {
			nodeName := "shard-" + strconv.Itoa(i) + "-v" + strconv.Itoa(r)
			h := hr.hash(nodeName)

			hr.hashes = append(hr.hashes, h)
			hr.mapHashToShardIndex[h] = i
		}
	}
	sort.Slice(hr.hashes, func(i, j int) bool {
		return hr.hashes[i] < hr.hashes[j]
	})
	return hr
}

func (hr *HashRing) GetShardIndex(key string) int {
	if len(hr.hashes) == 0 {
		return 0
	}

	hash := hr.hash(key)
	// Find the first shard hash >= key hash (binary search)
	nearestHasheIndex := sort.Search(len(hr.hashes), func(i int) bool {
		return hr.hashes[i] >= hash
	})

	if nearestHasheIndex == len(hr.hashes) {
		nearestHasheIndex = 0
	}

	hashInRing := hr.hashes[nearestHasheIndex]
	return hr.mapHashToShardIndex[hashInRing]
}

func (hr *HashRing) hash(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}
