package utils

func IsNil[T interface{}](a T) bool {
	var nilT T
	return interface{}(a) == interface{}(nilT)
}
