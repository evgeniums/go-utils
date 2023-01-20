package test_utils

import (
	"path/filepath"
)

var Testing = ""
var ExternalConfigPath = ""
var Phone = ""

func IsTesting() bool {
	return Testing == "true"
}

func ExternalConfigFile(path string) string {
	if ExternalConfigPath == "" {
		return path
	}
	return filepath.Join(ExternalConfigPath, path)
}
