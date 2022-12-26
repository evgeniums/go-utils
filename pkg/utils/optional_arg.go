package utils

// Select either default or optional value if the optional value is given in argumments.
func OptionalArg[T any](defaultArg T, optional ...T) T {

	arg := defaultArg
	if len(optional) == 1 {
		arg = optional[0]
	}
	return arg
}
