package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKudisConfig(t *testing.T) {

	host := "imhost"
	addr := "imhost:52087"
	cf := KudisConfig{
		Host: host,
	}

	cf.LoadDefaults()
	assert.Equal(t, cf.Port, int32(52087))

	assert.Equal(t, cf.Addr(), addr)

}
