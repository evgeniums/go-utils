package utils

func CopyMap(m map[string]interface{}) map[string]interface{} {
	cp := make(map[string]interface{})
	for k, v := range m {
		vm, ok := v.(map[string]interface{})
		if ok {
			cp[k] = CopyMap(vm)
		} else {
			cp[k] = v
		}
	}
	return cp
}

func CopyMapOneLevel[T1 comparable, T2 any](m map[T1]T2) map[T1]T2 {
	cp := make(map[T1]T2)
	for k, v := range m {
		cp[k] = v
	}
	return cp
}

func AppendMap[T any](m1 map[string]T, m2 map[string]T) {
	for k, v := range m2 {
		m1[k] = v
	}
}

func AppendMapNew(m1 map[string]interface{}, m2 map[string]interface{}) map[string]interface{} {
	nm := CopyMap(m1)
	for k, v := range m2 {
		nm[k] = v
	}
	return nm
}

func AllMapKeys[T1 comparable, T2 any](m map[T1]T2) []T1 {
	keys := make([]T1, 0)
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func AllMapValues[T1 comparable, T2 any](m map[T1]T2) []T2 {
	vals := make([]T2, 0)
	for _, v := range m {
		vals = append(vals, v)
	}
	return vals
}
