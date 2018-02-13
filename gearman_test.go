package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGearmanConfig(t *testing.T) {

	inf := "testinterface"
	cf := GearmanConfig{
		Interface: inf,
	}
	assert.True(t, cf.Unresolved())

	err := cf.LoadDefaults()
	assert.Equal(t, cf.Port, int32(4730))
	assert.Equal(t, cf.Host, "localhost")
	assert.NoError(t, err)

	host := "imhost"
	port := 6380
	addr := "imhost:6380"
	cf.SetHost(host)
	cf.SetPort(int32(port))
	assert.Equal(t, cf.Addr(), addr)
	assert.Nil(t, cf.GetPublic())
	assert.Equal(t, cf.GetInterface(), inf)
}
