package utils

func MaskCard(pan string, maskPrefix ...bool) string {
	size := len(pan)
	res := ""
	noPrefix := OptionalArg(false, maskPrefix...)
	for i := 0; i < size; i++ {
		if (!noPrefix && i < 6) || i > (size-5) {
			res += string(pan[i])
		} else {
			res += "*"
		}
	}
	return res
}

func FormatCard(pan string) string {
	result := ""
	for i := 0; i < len(pan); i++ {
		if i != 0 && i%4 == 0 {
			result += " "
		}
		result += string(pan[i])
	}

	return result
}
