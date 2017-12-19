package fileutils

import (
	"bufio"
	"fmt"
	"os"
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
