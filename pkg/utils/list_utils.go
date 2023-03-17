package utils

func ListInterfaces[T any](in ...T) []interface{} {
	result := make([]interface{}, len(in))
	for i := 0; i < len(in); i++ {
		result[i] = in[i]
	}
	return result
}
