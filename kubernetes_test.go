package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKubernetesConfigLoadDefaults(t *testing.T) {
	c := &KubernetesConfig{}
	assert.True(t, CanLoadDefaults(c))
	LoadDefaults(c)
	assert.Equal(t, "default", c.Namespace)
	assert.NotNil(t, c.OutCluster)
	assert.Equal(t, "ExternalIP", c.OutCluster.AddressType)
}
