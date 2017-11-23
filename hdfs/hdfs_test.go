package hdfs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClientForUser(t *testing.T) {
	assert := assert.New(t)
	client, err := NewClientForUser("35.201.180.4:8020", "root")
	defer client.Close()

	assert.NoError(err)
	assert.NotEmpty(client)
}

func TestNewClientForUser_1_error(t *testing.T) {
	assert := assert.New(t)
	client, err := NewClientForUser("", "root")

	assert.Error(err)
	assert.Empty(client)
}

func TestNewClient_1(t *testing.T) {
	assert := assert.New(t)
	os.Setenv(HadoopUserName, "root")
	os.Setenv(HadoopNamenodeAddress, "35.201.180.4")
	os.Setenv(HadoopNamenodePort, "8020")
	client, err := NewClient()
	defer client.Close()

	assert.NoError(err)
	assert.NotEmpty(client)
}

func TestNewClient_2(t *testing.T) {
	assert := assert.New(t)
	os.Setenv(HadoopUserName, "root")
	os.Setenv(HadoopNamenodeAddress, "35.201.180.4")
	os.Setenv(HadoopNamenodePort, "")
	client, err := NewClient()

	assert.Error(err)
	assert.Empty(client)

}

func TestNewClient_3(t *testing.T) {
	assert := assert.New(t)
	os.Setenv(HadoopUserName, "root")
	os.Setenv(HadoopNamenodeAddress, "")
	os.Setenv(HadoopNamenodePort, "8020")
	client, err := NewClient()

	assert.Error(err)
	assert.Empty(client)
}

func TestNewClient_4(t *testing.T) {
	assert := assert.New(t)
	// client will use the current user as default user
	os.Setenv(HadoopUserName, "")
	os.Setenv(HadoopNamenodeAddress, "35.201.180.4")
	os.Setenv(HadoopNamenodePort, "8020")
	client, err := NewClient()

	assert.NoError(err)
	assert.NotEmpty(client)
}
