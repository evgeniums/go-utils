package utils

import "encoding/json"

func DumpJson(obj interface{}) string {
	b, _ := json.Marshal(obj)
	return string(b)
}

func DumpPrettyJson(obj interface{}) string {
	b, _ := json.MarshalIndent(obj, "", "   ")
	return string(b)
}
