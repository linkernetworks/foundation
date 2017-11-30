package hdfs

import (
	"errors"
	"io/ioutil"
	"os"

	"bitbucket.org/linkernetworks/cv-tracker/src/logger"
	"github.com/colinmarc/hdfs"
)

const (
	HadoopUserName        = "HADOOP_USER_NAME"
	HadoopNamenodeAddress = "HADOOP_NAMENODE_ADDRESS"
	HadoopNamenodePort    = "HADOOP_NAMENODE_PORT"
)

func NewClient() (*hdfs.Client, error) {
	if len(os.Getenv(HadoopUserName)) < 0 {
		logger.Error("Env: HADOOP_USER_NAME isn't set!")
		return nil, errors.New("Env: HADOOP_USER_NAME isn't set!")
	}
	if len(os.Getenv(HadoopNamenodeAddress)) < 0 {
		logger.Errorf("Env: HADOOP_NAMENODE_ADDRESS isn't set!")
		return nil, errors.New("Env: HADOOP_NAMENODE_ADDRESS isn't set!")
	}
	if len(os.Getenv(HadoopNamenodePort)) < 0 {
		logger.Errorf("Env: HADOOP_NAMENODE_PORT isn't set!")
		return nil, errors.New("Env: HADOOP_NAMENODE_PORT isn't set!")
	}

	address := os.Getenv(HadoopNamenodeAddress) + ":" + os.Getenv(HadoopNamenodePort)
	user := os.Getenv(HadoopUserName)

	logger.Infof("Connect to hdfs address: %s, with user: %s", address, user)

	client, error := hdfs.NewForUser(address, user)
	return client, error
}

func NewClientForUser(address string, user string) (*hdfs.Client, error) {
	if address == "" {
		logger.Errorf("address: can't be empty!")
		return nil, errors.New("address: can't be empty!")
	}
	if user == "" {
		logger.Errorf("user: can't be empty!")
		return nil, errors.New("user: can't be empty!")
	}

	logger.Infof("Connect to hdfs address: %s, with user: %s", address, user)
	client, error := hdfs.NewForUser(address, user)
	return client, error
}

func CreateFile(client *hdfs.Client, filePath string, replication int, blockSize int64, perm os.FileMode) error {
	writer, err := client.CreateFile(filePath, replication, blockSize, perm)
	if err != nil {
		logger.Errorf("Create file %s failed on hdfs", filePath)
		return err
	} else {
		// create file with no error, try to upload the file contentType
		databytes, dataErr := ioutil.ReadFile(filePath)
		if dataErr != nil {
			logger.Errorf("Read file content error for file: %s, Error: %v", filePath, dataErr)
			return dataErr
		} else {
			_, writeErr := writer.Write(databytes)
			defer writer.Close()
			if writeErr != nil {
				logger.Errorf("Write file: %s to hdfs failed! Error: %v", filePath, writeErr)
				return writeErr
			} else {
				logger.Infof("Write file: %s to hdfs successfully!", filePath)
			}
		}
		return nil
	}
}

func CreateFileWithBytes(client *hdfs.Client, filePath string, data []byte, replication int, blockSize int64, perm os.FileMode) error {
	writer, err := client.CreateFile(filePath, replication, blockSize, perm)
	if err != nil {
		logger.Errorf("Create file %s failed on hdfs", filePath)
		return err
	} else {
		// create file with no error, try to upload the file contentType

		_, writeErr := writer.Write(data)
		defer writer.Close()
		if writeErr != nil {
			logger.Errorf("Write file: %s to hdfs failed! Error: %v", filePath, writeErr)
			return writeErr
		} else {
			logger.Infof("Write file: %s to hdfs successfully!", filePath)
		}

		return nil
	}
}

func RemoveFile(client *hdfs.Client, filePath string) error {
	err := client.Remove(filePath)
	if err != nil {
		logger.Errorf("Remove file: %s failed!", filePath)
		return err
	}
	return nil
}
