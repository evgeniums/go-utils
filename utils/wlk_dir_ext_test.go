package utils_test

import (
	"path/filepath"
	"testing"

	"github.com/evgeniums/go-backend-helpers/utils"
)

func TestWalkDirExt(t *testing.T) {

	handler := func(filePath string) error {
		t.Logf("File path: %s", filePath)
		return nil
	}

	root := filepath.Dir(scriptPath)
	utils.WalkDirExt(handler, root, "go")
	utils.WalkDirExt(handler, root, "pem")
}
