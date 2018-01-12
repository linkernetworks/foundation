package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJobUpdaterConfig(t *testing.T) {
	cf := JobUpdaterConfig{}
	jsontext := `{
	}`
	err := json.Unmarshal([]byte(jsontext), &cf)
	assert.NoError(t, err)
}
