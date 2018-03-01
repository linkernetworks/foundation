package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNotificationConfig(t *testing.T) {
	cf := NotificationConfig{}
	jsontext := `{
        "on_job_start": true,
        "on_job_success": true,
		"on_job_fail": true,
		"on_job_stop": true,
		"on_job_delete": true
	}`
	err := json.Unmarshal([]byte(jsontext), &cf)
	assert.NoError(t, err)

	assert.True(t, cf.OnJobStart)
	assert.True(t, cf.OnJobSuccess)
	assert.True(t, cf.OnJobFail)
	assert.True(t, cf.OnJobStop)
	assert.True(t, cf.OnJobDelete)
}
