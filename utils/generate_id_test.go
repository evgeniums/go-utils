package utils_test

import (
	"testing"

	"github.com/evgeniums/go-backend-helpers/utils"
)

func TestGenerateId(t *testing.T) {
	for i := 0; i < 300; i++ {
		t.Logf("%d: %v", i, utils.GenerateID())
	}
}
