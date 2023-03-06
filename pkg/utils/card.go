package utils

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
