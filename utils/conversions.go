package utils

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

func StrToFloat(s string) (float64, error) {
	if s == "" {
		return 0, nil
	}
	str := strings.ReplaceAll(s, ",", ".")
	str = strings.ReplaceAll(str, " ", "")
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0, err
	}
	return float64(f), nil
}

func FloatToStr(val float64) string {
	v := math.Round(float64(val)*100) / 100
	str := strconv.FormatFloat(float64(v), 'f', -1, 32)
	return str
}

func FloatToStr2(val float64) string {
	v := math.Round(float64(val)*100) / 100
	str := strconv.FormatFloat(float64(v), 'f', 2, 32)
	return str
}

func FloatToStr2Comma(val float64) string {
	str := FloatToStr2(val)
	str = strings.ReplaceAll(str, ".", ",")
	return str
}

func FloatToStr2Hyphen(val float64) string {
	str := FloatToStr2(val)
	str = strings.ReplaceAll(str, ".", "-")
	return str
}

func RoublyToKopeyki(roubles float64) int {
	return int(math.Round(float64(roubles) * 100.00))
}

func KopeykiToRoubly(kopeyki int) float64 {
	v := float64(kopeyki) / 100.00
	return float64(v)
}

func RoundMoneyUp(value float64) float64 {
	r := math.Ceil(float64(value)*100) / 100
	return float64(r)
}

func RoundMoneyDown(value float64) float64 {
	r := math.Floor(float64(value)*100) / 100
	return float64(r)
}

func RoundMoney(value float64) float64 {
	r := math.Round(float64(value)*100) / 100
	return r
}
func FormatInn(value int64) string {
	return fmt.Sprintf("%012d", value)
}

func TimeToStr(t time.Time) string {
	str := fmt.Sprintf("%02d.%02d.%04d %02d:%02d", t.Day(), t.Month(), t.Year(), t.Hour(), t.Minute())
	return str
}
