package utils_test

import (
	"testing"

	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

func CheckEqual(t *testing.T, value string, sample string) {
	if value != sample {
		t.Errorf("Invalid value: expected %v, got %v", sample, value)
	}
}

func TestFloatToStr(t *testing.T) {

	f1 := float64(1000.0)
	f2 := float64(1000.10)
	f3 := float64(1000.11)
	f4 := float64(1000.111)
	f5 := float64(1000.1111)
	f6 := float64(1000.1175)

	t.Logf(utils.FloatToStr(f1))
	t.Logf(utils.FloatToStr(f2))
	t.Logf(utils.FloatToStr(f3))
	t.Logf(utils.FloatToStr(f4))
	t.Logf(utils.FloatToStr(f5))
	t.Logf(utils.FloatToStr(f6))

	CheckEqual(t, utils.FloatToStr(f1), "1000")
	CheckEqual(t, utils.FloatToStr(f2), "1000.1")
	CheckEqual(t, utils.FloatToStr(f3), "1000.11")
	CheckEqual(t, utils.FloatToStr(f4), "1000.11")
	CheckEqual(t, utils.FloatToStr(f5), "1000.11")
	CheckEqual(t, utils.FloatToStr(f6), "1000.12")
}
