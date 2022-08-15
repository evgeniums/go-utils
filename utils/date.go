package utils

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Date time.Time

var DateNil Date

func (d *Date) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")

	if s == "" {
		*d = Date{}
		return nil
	}

	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	*d = Date(t)
	return nil
}

func (d *Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Date) String() string {
	t := time.Time(*d)
	if t.Equal(time.Time{}) {
		return ""
	}
	formatted := fmt.Sprintf("%04d-%02d-%02d", t.Year(), t.Month(), t.Day())
	return formatted
}

func (d *Date) RuDate() string {
	t := time.Time(*d)
	formatted := fmt.Sprintf("%02d.%02d.%04d", t.Day(), t.Month(), t.Year())
	return formatted
}

func (d *Date) RuDateShort() string {
	t := time.Time(*d)
	formatted := fmt.Sprintf("%02d.%02d.%02d", t.Day(), t.Month(), t.Year()-2000)
	return formatted
}

func (d *Date) AsNumber() string {
	t := time.Time(*d)
	formatted := fmt.Sprintf("%04d%02d%02d", t.Year(), t.Month(), t.Day())
	return formatted
}

func BeginningOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func EndOfDay(t time.Time) time.Time {
	b := BeginningOfDay(t)
	r := b.Add(time.Hour * 24).Add(-time.Second)
	return r
}

func StrToDate(s string) (Date, error) {
	t, err := StrToTime(s)
	return Date(t), err
}

func StrToTime(s string) (time.Time, error) {

	if s == "" {
		return time.Time{}, nil
	}

	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		t, err = time.Parse("02.01.2006", s)
		if err != nil {
			return time.Time{}, err
		}
	}
	return t, nil
}

func Today() Date {
	return Date(time.Now())
}

func DateOfTime(t time.Time) Date {
	return Date(BeginningOfDay(t))
}
