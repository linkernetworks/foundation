package timeutils

import (
	"time"
)

func Now() *time.Time {
	t := time.Now()
	return &t
}
