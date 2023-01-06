package utils

import (
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"
)

var idCount atomic.Uint32

func GenerateID() string {
	t := time.Now().UTC().Unix()
	r1 := rand.Int31n(0x7fffffff)

	count := idCount.Add(1) % 0x10000

	id := fmt.Sprintf("%08x%04x%04x", t, count, r1)
	return id
}

func GenerateRand64() string {
	r1 := rand.Int63()
	id := fmt.Sprintf("%016x", r1)
	return id
}
