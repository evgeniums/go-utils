package crypt_utils

import (
	cryptorand "crypto/rand"

	"github.com/dchest/uniuri"
)

func GenerateCryptoRand(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := cryptorand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func GenerateString(length ...int) string {
	if len(length) == 0 {
		return uniuri.New()
	}
	return uniuri.NewLen(length[0])
}
