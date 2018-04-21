package timeutils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTruncateRedisDateTime(t *testing.T) {
	t1, err := time.Parse(time.RFC3339Nano, "2017-12-18T11:23:10.244286455+08:00")
	assert.NoError(t, err)
	expect1, err := time.Parse(time.RFC3339Nano, "2017-12-18T11:23:10.244000000+08:00")
	assert.NoError(t, err)
	assert.Equal(t, TruncateRedisDateTime(&t1, time.Millisecond), &expect1)
}

func TestGetNowTime(t *testing.T) {
	now := Now()
	assert.NotNil(t, now)
}
func TestMax(t *testing.T) {
	t1 := time.Duration(1)
	t2 := time.Duration(2)

	max := Max(t1, t2)
	assert.Equal(t, t2, max)
	max = Max(t2, t1)
	assert.Equal(t, t2, max)
}
