package crypt_utils

import "encoding/base64"

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
