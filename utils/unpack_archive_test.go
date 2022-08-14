package utils_test

import (
	"io/ioutil"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/evgeniums/go-backend-helpers/utils"
)

var (
	_, scriptPath, _, _ = runtime.Caller(0)

	singleArchive   = filepath.Join(filepath.Dir(scriptPath), "test_assets/archive-single.rar")
	multipleArchive = filepath.Join(filepath.Dir(scriptPath), "test_assets/archive-multiple.rar")
)

func testUnpackArchive(t *testing.T, fileName string, fileCount int) {
	dat, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Fatalf("Failed to read archive: %s", err)
	}

	arch, err := utils.UnpackTempArchive(dat, "pk_", "rar")
	if err != nil {
		t.Fatalf("Failed to unpack archive: %s", err)
	}
	defer arch.Clean()

	t.Logf("Archive directory: %v", arch.Dir)
	for _, file := range arch.Files {
		t.Logf("* file: %v", file.Name())
	}
	if len(arch.Files) != fileCount {
		t.Fatalf("Invalid number of files in archive: expected %d, got %d", fileCount, len(arch.Files))
	}
}

func TestUnpackSingleArchive(t *testing.T) {
	testUnpackArchive(t, singleArchive, 1)
}

func TestUnpackMultipleArchive(t *testing.T) {
	testUnpackArchive(t, multipleArchive, 2)
}
