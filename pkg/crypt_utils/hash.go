package crypt_utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"

	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type Hash struct {
	Hash         hash.Hash
	StringCoding StringCoding
}

func (h *Hash) Sum() []byte {
	return h.Hash.Sum(nil)
}

func (h *Hash) Add(data []byte) {
	h.Hash.Write(data)
}

func (h *Hash) Calc(data ...[]byte) []byte {
	for _, block := range data {
		h.Hash.Write(block)
	}
	return h.Hash.Sum(nil)
}

func (h *Hash) CalcStr(data ...[]byte) string {
	return h.StringCoding.Encode(h.Calc(data...))
}

func (h *Hash) CalcStrStr(data ...string) string {
	for _, d := range data {
		h.Add([]byte(d))
	}
	return h.CalcStr()
}

func (h *Hash) Check(sum []byte) bool {
	return hmac.Equal(h.Sum(), sum)
}

func (h *Hash) CheckStr(sum string) error {

	sumB, err := h.StringCoding.Decode(sum)
	if err != nil {
		return fmt.Errorf("failed to decode sum: %s", err)
	}

	if !hmac.Equal(h.Sum(), sumB) {
		return errors.New("hash sums mismatch")
	}

	return nil
}

func NewHash(digestBuilder ...DigestBuilder) *Hash {
	var builder = utils.OptionalArg(sha256.New, digestBuilder...)
	h := &Hash{Hash: builder(), StringCoding: &Base64StringCoding{}}
	return h
}
