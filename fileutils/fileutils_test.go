package fileutils

import (
	"bytes"
	"crypto/sha256"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/mholt/archiver"
	"github.com/stretchr/testify/assert"
)

func TestExists(t *testing.T) {
	testDir, err := ioutil.TempDir(".", "target")
	assert.NoError(t, err)
	defer os.RemoveAll(testDir)

	_, err = os.Create(testDir + "/test")
	assert.NoError(t, err)

	existBool, err := Exists(testDir + "/test")
	assert.True(t, existBool)
}

func TestNotExists(t *testing.T) {
	existBool, _ := Exists("test/path/dont/exist")
	assert.False(t, existBool)
}

func TestWriteLines(t *testing.T) {
	testLines := []string{"testString\ntestString", "testString", "", " ", "中文", "!@#$"}
	err := WriteLines("test/path/dont/exist", testLines)
	assert.Error(t, err)

	targetDir, err := ioutil.TempDir(".", "target")
	assert.NoError(t, err)
	defer os.RemoveAll(targetDir)

	err = WriteLines(targetDir+"/test", testLines)
	assert.NoError(t, err, "Fail to write the lines.")

	_, err = os.Stat(targetDir + "/test")
	assert.NoError(t, err)

	// Check if the content of the file match the input string.
	content, err := ioutil.ReadFile(targetDir + "/test")
	assert.NoError(t, err)

	writeLines := strings.Split(string(content), "\n")
	var buffer bytes.Buffer
	for _, testLine := range testLines {
		buffer.WriteString(testLine + "\n")
	}
	testLines = strings.Split(buffer.String(), "\n")
	assert.True(t, reflect.DeepEqual(testLines, writeLines), "Content of the file didn't match the input.")
}

func TestFindFilesDontExistByExtension(t *testing.T) {
	var findResult []string
	assert.Equal(t, FindFilesByExtension(".", "apple"), findResult)
}

func TestFindFilesByExtension(t *testing.T) {
	testDir, err := ioutil.TempDir(".", "target")
	assert.NoError(t, err)
	defer os.RemoveAll(testDir)

	var findResult []string
	_, err = os.Create(testDir + "/apple")
	_, err = os.Create(testDir + "/pineapple")
	_, err = os.Create(testDir + "/pen")
	testSubDir, err := ioutil.TempDir("./"+testDir, "appleDir")
	_, err = os.Create(testSubDir + "/appleInDir")
	_, err = os.Create(testSubDir + "/bananaInDir")

	findResult = []string{"apple", "appleInDir", "pineapple"}
	assert.Equal(t, FindFilesByExtension(".", "apple"), findResult)
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

func TestRsync(t *testing.T) {
	srcDir, err := ioutil.TempDir(".", "source")
	assert.NoError(t, err, "asdf")
	defer os.RemoveAll(srcDir)

	destDir, err := ioutil.TempDir(".", "destination")
	assert.NoError(t, err, "asdf")
	defer os.RemoveAll(destDir)

	_, err = os.Create(srcDir + "/testFile1")
	_, err = os.Create(srcDir + "/testFile2")
	_, err = os.Create(srcDir + "/testFile3")
	testSubDir, err := ioutil.TempDir("./"+srcDir, "testSubDir")
	_, err = os.Create(testSubDir + "/testFile4")
	_, err = os.Create(testSubDir + "/testFile5")

	// Check if the sync successful.
	err = Rsync(srcDir, destDir)
	assert.NoError(t, err, "asdf")

	_, err = os.Stat(destDir + "/" + srcDir)
	assert.NoError(t, err, "asdf")

	// Check if the contents of the srcDir and destDir the same
	// Compare the SHA265 checksum of the zip files from two folders
	err = archiver.Zip.Make("src.zip", []string{srcDir})
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll("src.zip")

	err = archiver.Zip.Make("dest.zip", []string{destDir + "/" + srcDir})
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll("dest.zip")

	fSrc, err := os.Open("src.zip")
	if err != nil {
		log.Fatal(err)
	}
	defer fSrc.Close()

	fDest, err := os.Open("dest.zip")
	if err != nil {
		log.Fatal(err)
	}
	defer fDest.Close()

	var hashSrc, hashDest int64
	hasher := sha256.New()

	if hashSrc, err = io.Copy(hasher, fSrc); err != nil {
		log.Fatal(err)
	}

	if hashDest, err = io.Copy(hasher, fDest); err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, hashSrc, hashDest, "The contents of the srcDir and destDir not the same.")
}

func TestScanDir(t *testing.T) {
	srcDir, err := ioutil.TempDir(".", "test")
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
