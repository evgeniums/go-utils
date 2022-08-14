package utils

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"

	"github.com/mholt/archiver"
)

type TempArchive struct {
	Name  string
	Dir   string
	Files []fs.FileInfo
}

func (t *TempArchive) Clean() {
	os.RemoveAll(t.Dir)
}

func UnpackTempArchive(archData []byte, prefix string, extention string) (*TempArchive, error) {
	archFile, err := ioutil.TempFile("", prefix+".*."+extention)
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %s", err)
	}
	defer os.Remove(archFile.Name())

	_, err = archFile.Write(archData)
	if err != nil {
		return nil, fmt.Errorf("failed to write archive data to temporarty file %v: %s", archFile.Name(), err)
	}

	err = archFile.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close archive %v: %s", archFile.Name(), err)
	}

	arch := &TempArchive{}

	arch.Dir, err = ioutil.TempDir("", prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to create temporarty directory: %s", err)
	}

	err = archiver.Unarchive(archFile.Name(), arch.Dir)
	if err != nil {
		os.RemoveAll(arch.Dir)
		return nil, fmt.Errorf("failed to unarchive file %v to %v: %s", archFile.Name(), arch.Dir, err)
	}

	arch.Files, err = ioutil.ReadDir(arch.Dir)
	if err != nil {
		os.RemoveAll(arch.Dir)
		return nil, fmt.Errorf("failed to list files in temporary directory: %s", err)
	}

	return arch, nil
}
