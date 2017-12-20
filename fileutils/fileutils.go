package fileutils

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

//MakeDirs create dir and make it as 777 (for jupyter notebook consideration)
func MakeDirs(paths []string) {
	for _, v := range paths {
		os.MkdirAll(v, 777)
	}
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
			r, err := regexp.MatchString(ext, f.Name())
			if err == nil && r {
				files = append(files, f.Name())
			}
		}
		return nil
	})
	return files
}
