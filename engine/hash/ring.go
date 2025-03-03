package hash

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"log/slog"
	"os"
	"sort"
	"sync"

	"github.com/ohnomail00/super-duper-s3/engine"
)

type Ring struct {
	nodes   []uint32
	nodeMap map[uint32]engine.Server
	sync.RWMutex
}

// GeneratePartKey generates a unique key for a part.
func GeneratePartKey(index int, offset int) string {
	var randomPart uint32
	err := binary.Read(rand.Reader, binary.LittleEndian, &randomPart)
	if err != nil {
		slog.Error("failed to generate random part", "error", err)
		os.Exit(2)
	}
	return fmt.Sprintf("part-%d-%d-%d", index, offset, randomPart)
}

func NewRing() *Ring {
	return &Ring{nodeMap: make(map[uint32]engine.Server)}
}

func (r *Ring) AddNode(server engine.Server, replicas int) {
	r.Lock()
	defer r.Unlock()
	for i := 0; i < replicas; i++ {
		nodeKey := fmt.Sprintf("%s#%d", server.Address, i)
		hash := r.hashKeyFNV(nodeKey)
		r.nodes = append(r.nodes, hash)
		r.nodeMap[hash] = server
	}
	sort.Slice(r.nodes, func(i, j int) bool { return r.nodes[i] < r.nodes[j] })
}

func (r *Ring) GetNode(key string) engine.Server {

	if len(r.nodes) == 0 {
		panic("No nodes in the ring")
	}
	hash := r.hashKeyFNV(key)
	idx := sort.Search(len(r.nodes), func(i int) bool { return r.nodes[i] >= hash })
	if idx == len(r.nodes) {
		idx = 0
	}
	return r.nodeMap[r.nodes[idx]]
}

func (r *Ring) hashKeyFNV(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}
