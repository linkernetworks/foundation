package timeutils

import (
	"time"
)

func Now() *time.Time {
	t := time.Now()
	return &t
}

func TruncateRedisDateTime(t *time.Time, d time.Duration) *time.Time {
	if t != nil {
		result := t.Truncate(d)
		return &result
	}
	return nil
}
