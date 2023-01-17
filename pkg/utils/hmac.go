package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"hash"
)

type StringCoding interface {
	Encode(data []byte) string
	Decode(data string) ([]byte, error)
}

type Base64StringCoding struct {
}

func (b *Base64StringCoding) Encode(data []byte) string {
	return base64.RawStdEncoding.WithPadding(base64.StdPadding).EncodeToString(data)
}

func (b *Base64StringCoding) Decode(data string) ([]byte, error) {
	return base64.RawStdEncoding.WithPadding(base64.StdPadding).DecodeString(data)
}

type Hmac struct {
	Hash         hash.Hash
	StringCoding StringCoding
}

func (h *Hmac) Sum() []byte {
	return h.Hash.Sum(nil)
}

func (h *Hmac) Calc(data []byte) []byte {
	h.Hash.Write(data)
	return h.Hash.Sum(nil)
}

func (h *Hmac) CalcStr(data []byte) string {
	return h.StringCoding.Encode(h.Calc(data))
}

func (h *Hmac) CalcStrStr(data string) string {
	return h.CalcStr([]byte(data))
}

func (h *Hmac) Check(sum []byte) bool {
	return hmac.Equal(h.Sum(), sum)
}

func (h *Hmac) CheckStr(sum string) error {

	sumB, err := h.StringCoding.Decode(sum)
	if err != nil {
		return fmt.Errorf("failed to decode sum: %s", err)
	}

	if !hmac.Equal(h.Sum(), sumB) {
		return errors.New("invalid HMAC")
	}

	return nil
}

type DigestBuilder = func() hash.Hash

func NewHmac(secret string, digestBuilder ...DigestBuilder) *Hmac {
	var builder = OptionalArg(sha256.New, digestBuilder...)
	hm := hmac.New(builder, []byte(secret))
	h := &Hmac{Hash: hm}
	h.StringCoding = &Base64StringCoding{}
	return h
}
