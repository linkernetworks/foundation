package config

import (
	"testing"

	"bitbucket.org/linkernetworks/aurora/src/entity"
	"github.com/stretchr/testify/assert"

	"gopkg.in/mgo.v2/bson"
)

func TestWorkspacePVSubpath(t *testing.T) {
	configPath := "../../config/testing.json"
	w := entity.Workspace{
		ID: bson.ObjectIdHex("5a4c86314ce27e00019452dd"),
	}
	cf := MustRead(configPath)
	subpath := cf.GetWorkspacePVSubpath(&w)
	assert.Equal(t, "batches/batch-5a4c86314ce27e00019452dd", subpath)
}

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
