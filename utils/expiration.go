package utils

import "time"

func IsExpired(t time.Time) bool {
	now := time.Now()
	return now.After(t)
}

func IsExpiredUTC(t time.Time) bool {
	now := time.Now().UTC()
	return now.After(t)
}

func ExpirationSecond(ttl int) time.Time {
	return time.Now().Add(time.Second * time.Duration(ttl))
}

func ExpirationSecondUTC(ttl int) time.Time {
	return time.Now().UTC().Add(time.Second * time.Duration(ttl))
}

func ExpirationHour(ttl int) time.Time {
	return time.Now().Add(time.Hour * time.Duration(ttl))
}

func IsExpiredDelay(t time.Time, seconds int) bool {
	t1 := t.Add(time.Second * time.Duration(seconds))
	now := time.Now().UTC()
	return now.After(t1)
}

func IsExpiredDelayLocal(t time.Time, seconds int) bool {
	t1 := t.Add(time.Second * time.Duration(seconds))
	now := time.Now()
	return now.After(t1)
}
