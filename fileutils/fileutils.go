package fileutils

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"mime"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"

	"bitbucket.org/linkernetworks/aurora/src/logger"
	"bitbucket.org/linkernetworks/aurora/src/utils/sysutils"
)

type FileInfo struct {
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	Type    string    `json:"type"`
	ModTime time.Time `json:"mtime"`
	IsDir   bool      `json:"isDir"`
}

//Exists - check path if exist or not legal.
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

//Write the lines into the path
func WriteLines(filepath string, lines []string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}

//FindFilesByExtension find files by specific extension name
func FindFilesByExtension(pathS, ext string) []string {

	var files []string
	filepath.Walk(pathS, func(path string, f os.FileInfo, _ error) error {
		if !f.IsDir() {
			if filepath.Ext(path) == ext {
				files = append(files, f.Name())
			}
		}
		return nil
	})
	return files
}

func CopyFile(srcDir, destDir, file string) error {
	cmd := exec.Command("cp", srcDir+"/"+file, destDir+"/"+file)

	err, _, _ := sysutils.ExecuteCommand(cmd)
	return err
}

//Rsync copies the whole folder using os console command to avoid edge effect of golang file copy.
func Rsync(src, dst string) error {
	cmd := exec.Command("rsync", "-av", src, dst)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logger.Infoln("[rsync]", err)
		return err
	}
	defer stdout.Close()
	if err := cmd.Start(); err != nil {
		logger.Infoln("[rsync]", err)
		return err
	}
	opBytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		logger.Infoln("[rsync]", err)
		return err
	}
	logger.Infoln("[rsync]", string(opBytes))
	return nil
}

func ScanDir(p string) ([]FileInfo, error) {
	fileInfos := []FileInfo{}
	files, err := ioutil.ReadDir(p)
	if err != nil {
		return fileInfos, err
	}

	for _, file := range files {
		fileInfos = append(fileInfos, FileInfo{
			Name:    file.Name(),
			Size:    file.Size(),
			ModTime: file.ModTime(),
			IsDir:   file.IsDir(),
			Type:    mime.TypeByExtension(path.Ext(file.Name())),
		})
	}

	return fileInfos, nil
}

//This function remove all files under the directory but the directory itself
func RemoveDirContents(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if dir != path {
			os.RemoveAll(path)
		}
		return nil
	})
}
