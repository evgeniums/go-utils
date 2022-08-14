package utils

import (
	"time"
)

type DaysInterval struct {
	From int
	To   int
}

func DaysToDuration(days int) time.Duration {
	return 24 * time.Hour * time.Duration(days)
}

func TimeAfterOrEqual(a time.Time, b time.Time) bool {
	return a.After(b) || a.Equal(b)
}

func TimeBeforeOrEqual(a time.Time, b time.Time) bool {
	return a.Before(b) || a.Equal(b)
}

func (s *DaysInterval) ElapsedInInterval(timePoint time.Time) bool {

	now := time.Now()

	from := timePoint.Add(DaysToDuration(s.From))
	to := timePoint.Add(DaysToDuration(s.To))

	if s.From == 0 {
		return TimeBeforeOrEqual(now, to)
	}

	if s.To == 0 {
		return TimeBeforeOrEqual(from, now)
	}

	return TimeBeforeOrEqual(from, now) && TimeBeforeOrEqual(now, to)
}

type DaysIntervalMap map[DaysInterval]interface{}

func (m *DaysIntervalMap) FindMatchingElapsed(timePoint time.Time) interface{} {

	for interval, value := range *m {
		if interval.ElapsedInInterval(timePoint) {
			return value
		}
	}

	return nil
}
