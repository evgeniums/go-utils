package utils

func MaskPhone(phone string) string {
	if phone == "" {
		return ""
	}
	if len(phone) <= 4 {
		return "****"
	}
	masked := ""
	for i := 0; i < len(phone)-4; i++ {
		masked += "*"
	}
	masked += phone[len(phone)-4:]
	return masked
}
