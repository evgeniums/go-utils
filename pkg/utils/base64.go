package utils

import "encoding/base64"

type StringCoding interface {
	Encode(data []byte) string
	Decode(data string) ([]byte, error)
}

type Base64StringCoding struct {
}

func (b *Base64StringCoding) Encode(data []byte) string {
	return Base64Encode(data)
}

func (b *Base64StringCoding) Decode(data string) ([]byte, error) {
	return Base64Decode(data)
}

type WithStringCoder interface {
	Coder() StringCoding
}

type WithStringCoderBase struct {
	StringCoding StringCoding
}

func (w *WithStringCoderBase) Construct(encoder ...StringCoding) {
	if len(encoder) == 0 {
		w.StringCoding = &Base64StringCoding{}
	} else {
		w.StringCoding = encoder[0]
	}
}

func (w *WithStringCoderBase) Coder() StringCoding {
	return w.StringCoding
}

func Base64Encode(data []byte) string {
	return base64.RawStdEncoding.WithPadding(base64.StdPadding).EncodeToString(data)
}

func Base64Decode(data string) ([]byte, error) {
	return base64.RawStdEncoding.WithPadding(base64.StdPadding).DecodeString(data)
}
