package utils

import "os"

type FileWriteReopen struct {
	Path string
	File *os.File
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func (f *FileWriteReopen) Write(p []byte) (n int, err error) {

	if !fileExists(f.Path) {
		if f.File != nil {
			f.File.Close()
		}
		f.File, err = os.OpenFile(f.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return 0, err
		}
	}

	return f.File.Write(p)
}
