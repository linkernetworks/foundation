package fileutils

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"bitbucket.org/linkernetworks/aurora/src/utils/sysutils"
)

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
			r, err := regexp.MatchString(ext, f.Name())
			if err == nil && r {
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

//FolderCopy copy whole folder using os console command to avoid edge effect of golang file copy.
func FolderCopy(src, dst string) error {
	cmd := exec.Command("cp", "-R", src, dst)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println(err)
		return err
	}
	defer stdout.Close()
	if err := cmd.Start(); err != nil {
		log.Println(err)
		return err
	}
	opBytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println(string(opBytes))
	return nil
}
