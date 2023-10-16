package utils

import (
	"os"
	"path/filepath"
)

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

func MakePath(filename string) error {
	dir := filepath.Dir(filename)
	err := os.MkdirAll(dir, os.ModePerm)
	return err
}
