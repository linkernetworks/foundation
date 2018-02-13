package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJupyterConfig(t *testing.T) {
	cf := JupyterConfig{
		Cache: &JupyterCacheConfig{},
	}

	cf.LoadDefaults()
	assert.Equal(t, cf.Cache.Expire, 600)

}
