package utils

import (
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"
)

var idCount atomic.Uint32

func GenerateID() string {
	t := time.Now().Unix()
	r1 := rand.Uint32()

	count := idCount.Add(1) % 0x10000

	id := fmt.Sprintf("%08x%04x%08x", t, count, r1)
	return id
}

func GenerateRand64() string {
	r1 := rand.Uint64()
	id := fmt.Sprintf("%016x", r1)
	return id
}

func GenerateRandInt(length ...int) string {
	r1 := rand.Uint64()
	id := fmt.Sprintf("%d", r1)
	if len(length) != 0 {
		return id[:length[0]]
	}
	return id
}
