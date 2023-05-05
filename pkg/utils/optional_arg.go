package utils

// Select either default or optional value if the optional value is given in argumments.
func OptionalArg[T any](defaultArg T, optional ...T) T {

	arg := defaultArg
	if len(optional) > 0 {
		arg = optional[0]
	}
	return arg
}

func OptionalString(defaultVal string, optional ...string) string {

	if len(optional) == 0 {
		return defaultVal
	}

	if optional[0] == "" {
		return defaultVal
	}
	return optional[0]
}
