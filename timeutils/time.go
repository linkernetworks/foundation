package timeutils

import (
	"time"
)

func Now() *time.Time {
	t := time.Now()
	return &t
}

func TruncateTime(t *time.Time, d time.Duration) *time.Time {
	result := t.Truncate(d)
	return &result
}
