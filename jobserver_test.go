package config

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJobControllerConfig(t *testing.T) {
	cf := JobServerConfig{}
	jsontext := `{
		"host": "localhost",
		"port": 50051,
		"deploymentTargets": {
			"default": {
				"type": "kubernetes",
				"kubernetes": {
					"config": "",
					"context": "",
					"namespace": "default"
				}
			}
		}
	}`
	err := json.Unmarshal([]byte(jsontext), &cf)
	assert.NoError(t, err)

	assert.Equal(t, cf.Addr(), "localhost:50051")

	cf.Port = 0
	cf.LoadDefaults()
	assert.Equal(t, cf.Port, int32(50051))
}
