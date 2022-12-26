package utils

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
)

type WalkDirExtHandler = func(filePath string) error

func WalkDirExt(handler WalkDirExtHandler, root string, ext string) error {

	regEx, err := regexp.Compile(fmt.Sprintf("^.+\\.(%s)$", ext))
	if err != nil {
		return err
	}
	each := func(path string, d fs.DirEntry, err error) error {

		if err != nil {
			return err
		}

		if !d.IsDir() && regEx.MatchString(d.Name()) {
			return handler(path)
		}

		return nil
	}

	return filepath.WalkDir(root, each)
}
