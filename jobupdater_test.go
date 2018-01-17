package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJobUpdaterConfig(t *testing.T) {
	cf := JobUpdaterConfig{}
	jsontext := `{
		"logger": {
            "dir": "./logs",
            "filePattern": "jobupdater.log.%Y%m%d",
            "linkName": "jobupdater",
            "level": "debug"
        }
	}`
	err := json.Unmarshal([]byte(jsontext), &cf)
	assert.NoError(t, err)
}
