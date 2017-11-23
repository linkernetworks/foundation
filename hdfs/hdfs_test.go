package hdfs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	HadoopTestUserName        = "TEST_HADOOP_USER_NAME"
	HadoopTestNamenodeAddress = "TEST_HADOOP_NAMENODE_ADDRESS"
	HadoopTestNamenodePort    = "TEST_HADOOP_NAMENODE_PORT"
)

func TestNewClientForUser(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_HDFS"); !defined {
		t.SkipNow()
		return
	}

	assert := assert.New(t)
	url := os.Getenv(HadoopTestNamenodeAddress) + ":" + os.Getenv(HadoopTestNamenodePort)
	client, err := NewClientForUser(url, os.Getenv(HadoopTestUserName))
	defer client.Close()

	assert.NoError(err)
	assert.NotEmpty(client)
}

func TestNewClientForUser_1_error(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_HDFS"); !defined {
		t.SkipNow()
		return
	}

	assert := assert.New(t)
	client, err := NewClientForUser("", "root")

	assert.Error(err)
	assert.Empty(client)
}

func TestNewClient_1(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_HDFS"); !defined {
		t.SkipNow()
		return
	}

	assert := assert.New(t)
	os.Setenv(HadoopUserName, os.Getenv(HadoopTestUserName))
	os.Setenv(HadoopNamenodeAddress, os.Getenv(HadoopTestNamenodeAddress))
	os.Setenv(HadoopNamenodePort, os.Getenv(HadoopTestNamenodePort))
	client, err := NewClient()
	defer client.Close()

	assert.NoError(err)
	assert.NotEmpty(client)
}

func TestNewClient_2(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_HDFS"); !defined {
		t.SkipNow()
		return
	}

	assert := assert.New(t)
	os.Setenv(HadoopUserName, os.Getenv(HadoopTestUserName))
	os.Setenv(HadoopNamenodeAddress, os.Getenv(HadoopTestNamenodeAddress))
	os.Setenv(HadoopNamenodePort, "")
	client, err := NewClient()

	assert.Error(err)
	assert.Empty(client)

}

func TestNewClient_3(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_HDFS"); !defined {
		t.SkipNow()
		return
	}

	assert := assert.New(t)
	os.Setenv(HadoopUserName, os.Getenv(HadoopTestUserName))
	os.Setenv(HadoopNamenodeAddress, "")
	os.Setenv(HadoopNamenodePort, os.Getenv(HadoopTestNamenodePort))
	client, err := NewClient()

	assert.Error(err)
	assert.Empty(client)
}

func TestNewClient_4(t *testing.T) {
	if _, defined := os.LookupEnv("TEST_HDFS"); !defined {
		t.SkipNow()
		return
	}

	assert := assert.New(t)
	// client will use the current user as default user
	os.Setenv(HadoopUserName, "")
	os.Setenv(HadoopNamenodeAddress, os.Getenv(HadoopTestNamenodeAddress))
	os.Setenv(HadoopNamenodePort, os.Getenv(HadoopTestNamenodePort))
	client, err := NewClient()

	assert.NoError(err)
	assert.NotEmpty(client)
}
