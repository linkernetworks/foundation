package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadTestingConfig(t *testing.T) {
	configPath := "../../config/testing.json"
	cf := MustRead(configPath)
	assert.Equal(t, "localhost", cf.Redis.Host)
	assert.NotEqual(t, int32(0), cf.Redis.Port)
}

func TestReadK8SConfig(t *testing.T) {
	configPath := "../../config/k8s.json"
	cf := MustRead(configPath)
	assert.Equal(t, "redis.default", cf.Redis.Host)
	assert.NotEqual(t, int32(0), cf.Redis.Port)
}

func TestReadLocalConfig(t *testing.T) {
	configPath := "../../config/local.json"
	cf := MustRead(configPath)
	assert.Equal(t, "localhost", cf.Redis.Host)
	assert.NotEqual(t, int32(0), cf.Redis.Port)
}

func TestReadTestingHdfsConfig(t *testing.T) {
	configPath := "../../config/testing.json"
	cf := MustRead(configPath)
	if cf.Hdfs != nil {
		assert.Equal(t, "35.201.180.4", cf.Hdfs.Host)
		assert.NotEqual(t, int32(0), cf.Hdfs.Port)
		assert.Equal(t, "root", cf.Hdfs.Username)
	}
}

func TestReadDataConifg(t *testing.T) {
	configPath := "../../config/testing.json"
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
