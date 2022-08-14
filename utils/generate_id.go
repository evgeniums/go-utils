package utils

import (
	"fmt"
	"math/rand"
	"time"
)

var idCount = 0

func GenerateID() string {
	t := time.Now().UTC().Unix()
	r1 := rand.Int31n(0xffff)
	r2 := rand.Int31n(0xff)

	count := idCount % 0x100
	idCount++

	id := fmt.Sprintf("%08x%02x%04x%02x", t, count, r1, r2)
	return id
}

func GenerateRand64() string {
	r1 := rand.Int63()
	id := fmt.Sprintf("%016x", r1)
	return id
}
