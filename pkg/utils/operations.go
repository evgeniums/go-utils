package utils

func List(vals ...interface{}) []interface{} {
	l := make([]interface{}, 0, len(vals))
	return append(l, vals...)
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
