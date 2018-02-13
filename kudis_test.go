package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKudisConfig(t *testing.T) {

	host := "imhost"
	addr := "imhost:52000"
	cf := KudisConfig{
		Host: host,
	}

	cf.LoadDefaults()
	assert.Equal(t, cf.Port, int32(52000))

	assert.Equal(t, cf.Addr(), addr)

}
