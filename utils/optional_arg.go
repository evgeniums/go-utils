package utils

func OptionalArg[T any](defaultArg T, optional ...T) T {

	arg := defaultArg
	if len(optional) == 1 {
		arg = optional[0]
	}
	return arg
}
