package timeutils

import (
	"time"
)

func Now() *time.Time {
	t := time.Now()
	return &t
}

func Max(d1, d2 time.Duration) time.Duration {
	if d1 < d2 {
		return d2
	}
	return d1
}

func TruncateRedisDateTime(t *time.Time, d time.Duration) *time.Time {
	if t != nil {
		result := t.Truncate(d)
		return &result
	}
	return nil
}
