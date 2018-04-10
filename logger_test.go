package config

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoggerConfig(t *testing.T) {
	cf := LoggerConfig{}
	jsontext := `{
            "dir": "./logs",
            "filePattern": "migration.log.%Y%m%d",
            "linkName": "migration",
            "level": "debug",
            "maxAge": "720h"
        }`
	err := json.Unmarshal([]byte(jsontext), &cf)
	assert.NoError(t, err)

	assert.Equal(t, cf.Dir, "./logs")
	assert.Equal(t, cf.FilePattern, "migration.log.%Y%m%d")
	assert.Equal(t, cf.LinkName, "migration")
	assert.Equal(t, cf.Level, "debug")
	assert.Equal(t, cf.MaxAge, "720h")

	d, err := time.ParseDuration(cf.MaxAge)
	assert.NoError(t, err)
	assert.NotZero(t, d)
}
