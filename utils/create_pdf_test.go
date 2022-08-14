package utils_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/evgeniums/go-backend-helpers/utils"
)

var (
	mdFile         = filepath.Join(filepath.Dir(scriptPath), "test_assets/test_pdf.md")
	pdfFile        = filepath.Join(filepath.Dir(scriptPath), "test_assets/test_pdf.pdf")
	testAssetsPath = filepath.Join(filepath.Dir(scriptPath), "test_assets")
)

func TestGeneratePdf(t *testing.T) {

	os.Remove(pdfFile)

	md, err := ioutil.ReadFile(mdFile)
	if err != nil {
		t.Logf("Failed to read md file: %s", err)
	}

	err = utils.CreatePdf(md, pdfFile, testAssetsPath)
	if err != nil {
		t.Fatalf("Failed to convert md to pdf: %s", err)
	}
}
