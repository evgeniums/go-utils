package crypt_utils

import (
	cryptorand "crypto/rand"
)

func GenerateCryptoRand(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := cryptorand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}
