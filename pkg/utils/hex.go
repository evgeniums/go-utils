package utils

import (
	"encoding/hex"
	"strings"
)

type HexStringCoding struct {
	UpperCase bool
}

func (h *HexStringCoding) Encode(data []byte) string {
	res := hex.EncodeToString(data)
	if h.UpperCase {
		return strings.ToUpper(res)
	}
	return res
}

func (h *HexStringCoding) Decode(data string) ([]byte, error) {
	l := strings.ToLower(data)
	return hex.DecodeString(l)
}
