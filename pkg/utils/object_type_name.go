package utils

import "reflect"

func ObjectTypeName(obj interface{}) string {
	t := reflect.TypeOf(obj)
	if reflect.ValueOf(obj).Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}
