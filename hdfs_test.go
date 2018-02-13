package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHdfsConfig(t *testing.T) {
	inf := "testinterface"
	cf := HdfsConfig{
		Interface: inf,
	}

	err := cf.LoadDefaults()
	assert.Equal(t, cf.Port, int32(8020))
	assert.Equal(t, cf.Username, "root")
	assert.True(t, cf.Unresolved())
	assert.NoError(t, err)

	user := "testusername"
	host := "imhost"
	port := 6380
	addr := "imhost:6380"
	cf.SetUsername(user)
	cf.SetHost(host)
	cf.SetPort(int32(port))
	assert.Equal(t, cf.Username, user)
	assert.Equal(t, cf.Addr(), addr)
	assert.False(t, cf.Unresolved())
	assert.Nil(t, cf.GetPublic())
	assert.Equal(t, cf.GetInterface(), inf)
}
