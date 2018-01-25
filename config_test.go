package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadTestingConfig(t *testing.T) {
	configPath := "../../config/testing.json"
	cf := Read(configPath)
	assert.Equal(t, "localhost", cf.Redis.Host)
	assert.NotEqual(t, int32(0), cf.Redis.Port)
}

func TestReadK8SConfig(t *testing.T) {
	configPath := "../../config/k8s.json"
	cf := Read(configPath)
	assert.Equal(t, "redis.default", cf.Redis.Host)
	assert.NotEqual(t, int32(0), cf.Redis.Port)
}

func TestReadLocalConfig(t *testing.T) {
	configPath := "../../config/local.json"
	cf := Read(configPath)
	assert.Equal(t, "localhost", cf.Redis.Host)
	assert.NotEqual(t, int32(0), cf.Redis.Port)
}

func TestReadTestingHdfsConfig(t *testing.T) {
	configPath := "../../config/testing.json"
	cf := Read(configPath)
	if cf.Hdfs != nil {
		assert.Equal(t, "35.201.180.4", cf.Hdfs.Host)
		assert.NotEqual(t, int32(0), cf.Hdfs.Port)
		assert.Equal(t, "root", cf.Hdfs.Username)
	}
}
