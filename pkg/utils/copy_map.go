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

func AppendMap[T any](m1 map[string]T, m2 map[string]T) map[string]T {
	for k, v := range m2 {
		m1[k] = v
	}

	return m1
}
