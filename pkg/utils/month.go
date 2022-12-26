package utils

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Month int

func CurrentMonth() Month {
	t := time.Now()
	var m Month
	m.SetTime(t)
	return m
}

func MonthFromString(str string) (Month, error) {

	if str == "" {
		return CurrentMonth(), nil
	}

	t, err := time.Parse("2006-01", str)
	if err != nil {
		return 0, err
	}
	var m Month
	m.SetTime(t)
	return m, nil
}

func (m *Month) String() string {
	str := fmt.Sprintf("%04d-%02d", m.Year(), m.Month())
	return str
}

func (m *Month) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	var err error
	*m, err = MonthFromString(s)
	return err
}

func (m *Month) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.String())
}

func (m *Month) Year() int {
	year := int(*m) / 100
	return year
}

func (m *Month) Month() int {
	month := int(*m) - (m.Year() * 100)
	return month
}

func (m *Month) Set(year int, month int) {
	*m = Month(year*100 + month)
}

func (m *Month) SetTime(t time.Time) {
	m.Set(t.Year(), int(t.Month()))
}

func (m *Month) Time() time.Time {
	t, _ := time.Parse("2006-01", m.String())
	return t
}

func (m *Month) Prev() Month {
	var prev Month
	month := m.Month()
	year := m.Year()
	if month == 1 {
		prev.Set(year-1, 12)
	} else {
		prev.Set(year, month-1)
	}
	return prev
}

func (m *Month) Next() Month {
	var next Month
	month := m.Month()
	year := m.Year()
	if month == 12 {
		next.Set(year+1, 1)
	} else {
		next.Set(year, month+1)
	}
	return next
}
