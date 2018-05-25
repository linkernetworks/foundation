package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadTestingConfig(t *testing.T) {
	configPath := "example.json"
	cf := MustRead(configPath)
	assert.Equal(t, "localhost", cf.Redis.Host)
	assert.NotEqual(t, int32(0), cf.Redis.Port)
}

func TestReadDataConifg(t *testing.T) {
	configPath := "example.json"
	cf := MustRead(configPath)
	if cf.Data != nil {
		assert.Equal(t, "./data/batches", cf.GetWorkspaceRootDir())
		assert.Equal(t, "./data/batches/archives", cf.GetArchiveDir())
		assert.Equal(t, "./data/images", cf.GetImageDir())
		assert.Equal(t, "./data/thumbnails", cf.GetThumbnailDir())
		assert.Equal(t, "./data/models", cf.GetModelDir())
		assert.Equal(t, "./data/models/archives", cf.GetModelArchiveDir())
	}
}
