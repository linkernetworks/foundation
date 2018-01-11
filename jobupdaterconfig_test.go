package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJobUpdaterConfig(t *testing.T) {
	cf := JobUpdaterConfig{}
	jsontext := `{
		"host": "localhost",
		"port": 27018
	}`
	err := json.Unmarshal([]byte(jsontext), &cf)
	assert.NoError(t, err)
}
