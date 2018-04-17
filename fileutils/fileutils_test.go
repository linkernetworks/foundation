package fileutils

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteLines(t *testing.T) {
	testLines := []string{"testString\ntestString", "testString", "", " ", "中文", "!@#$"}
	err := WriteLines("test/path/dont/exit", testLines)
	assert.Error(t, err)

	targetDir, err := ioutil.TempDir(".", "target")
	assert.NoError(t, err)
	defer os.RemoveAll(targetDir)

	err = WriteLines(targetDir+"/test", testLines)
	assert.NoError(t, err)

	content, err := ioutil.ReadFile(targetDir + "/test")
	assert.NoError(t, err)

	writeLines := strings.Split(string(content), "\n")
	var buffer bytes.Buffer
	for _, testLine := range testLines {
		buffer.WriteString(testLine + "\n")
	}
	testLines = strings.Split(buffer.String(), "\n")
	assert.True(t, reflect.DeepEqual(testLines, writeLines))
}

func TestCopyFile(t *testing.T) {
	srcDir, err := ioutil.TempDir(".", "source")
	assert.NoError(t, err)
	defer os.RemoveAll(srcDir)

	destDir, err := ioutil.TempDir(".", "destination")
	assert.NoError(t, err)
	defer os.RemoveAll(destDir)

	testFile := "test"
	f, err := os.Create(srcDir + "/" + testFile)
	assert.NoError(t, err)

	f.Write([]byte{12, 12, 12})
	f.Close()

	err = CopyFile(srcDir, destDir, testFile)
	assert.NoError(t, err)

	_, err = os.Stat(destDir + "/" + testFile)
	assert.NoError(t, err)
}

func TestScanDir(t *testing.T) {
	srcDir, err := ioutil.TempDir(".", "test-")
	defer os.RemoveAll(srcDir)
	assert.NoError(t, err)

	testFile := "test"
	f, err := os.Create(srcDir + "/" + testFile)
	f.Close()
	assert.NoError(t, err)

	fileInfos, err := ScanDir(srcDir)
	assert.NoError(t, err)

	assert.Equal(t, fileInfos[0].Name, testFile)
	assert.Equal(t, fileInfos[0].Size, int64(0))
	assert.Equal(t, fileInfos[0].IsDir, false)
	assert.Equal(t, fileInfos[0].Type, "")
}

func TestRemoveDirContents(t *testing.T) {
	dir, err := ioutil.TempDir(".", "test-")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	ioutil.TempFile(dir, "file-")
	ioutil.TempFile(dir, "file-")
	ioutil.TempFile(dir, "file-")
	count := 0
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		count++
		return nil
	})
	assert.Equal(t, count, 4)

	RemoveDirContents(dir)
	count = 0
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		count++
		return nil
	})
	assert.Equal(t, count, 1)
}
