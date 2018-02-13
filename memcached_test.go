package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemcachedConfig(t *testing.T) {
	inf := "testinterface"
	cf := MemcachedConfig{
		Interface: inf,
	}

	err := cf.LoadDefaults()
	assert.Equal(t, cf.Port, int32(11211))
	assert.True(t, cf.Unresolved())
	assert.NoError(t, err)

	host := "imhost"
	port := 6380
	addr := "imhost:6380"
	cf.SetHost(host)
	cf.SetPort(int32(port))
	assert.Equal(t, cf.Addr(), addr)
	assert.False(t, cf.Unresolved())
	assert.Nil(t, cf.GetPublic())
	assert.Equal(t, cf.GetInterface(), inf)
}
