package pathutil

import "os"

func Resolve(path string) (string, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return path, err
	}

	if (info.Mode() & os.ModeSymlink) == 0 {
		return path, nil
	}

	return os.Readlink(path)
}
