package utils

import "os"

func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return os.IsExist(err)
}

func IsFile(path string) bool {
	fileinfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !fileinfo.IsDir()
}

func IsDir(path string) bool {
	fileinfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileinfo.IsDir()
}
