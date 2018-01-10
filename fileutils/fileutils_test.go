package fileutils

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestCopyFile(t *testing.T) {
	srcDir, err := ioutil.TempDir(".", "source")
	assert.NoError(t, err)

	destDir, err := ioutil.TempDir(".", "destination")
	assert.NoError(t, err)

	testFile := "test"
	f, err := os.Create(srcDir + "/" + testFile)
	assert.NoError(t, err)

	f.Write([]byte{12, 12, 12})
	f.Close()

	err = CopyFile(srcDir, destDir, testFile)
	assert.NoError(t, err)

	_, err = os.Stat(destDir + "/" + testFile)
	assert.NoError(t, err)
	err = os.RemoveAll(srcDir)
	assert.NoError(t, err)
	err = os.RemoveAll(destDir)
	assert.NoError(t, err)
}
