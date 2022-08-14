package utils

import "encoding/json"

func MaskCard(pan string) string {
	size := len(pan)
	res := ""
	for i := 0; i < size; i++ {
		if i < 6 || i > (size-5) {
			res += string(pan[i])
		} else {
			res += "*"
		}
	}
	return res
}

type Pan string

func (p *Pan) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.Mask())
}

func (p *Pan) Mask() string {
	return MaskCard(string(*p))
}
