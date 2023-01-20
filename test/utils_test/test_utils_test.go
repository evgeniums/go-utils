package utils_test

import (
	"testing"

	"github.com/evgeniums/go-backend-helpers/pkg/test_utils"
)

func TestGlobalTestVars(t *testing.T) {

	t.Logf(`Global testing flags can be set with ldflags when running tests like: go test -ldflags="-X 'github.com/evgeniums/go-backend-helpers/pkg/test_utils.Testing=true'"`)

	t.Logf("Current values:")
	t.Logf("test_utils.Testing=%v", test_utils.Testing)
	t.Logf("test_utils.ExternalConfigPath=%s", test_utils.ExternalConfigPath)
	t.Logf("test_utils.Phone=%v", test_utils.Phone)
}
