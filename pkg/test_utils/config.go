package test_utils

var ExternalConfigPath = ""
var Testing = ""

func IsTesting() bool {
	return Testing == "true"
}
