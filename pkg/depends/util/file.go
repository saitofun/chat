package util

import "os"

func IsExist(path string) bool {
	if _, err := os.Stat(string(path)); err != nil {
		return os.IsExist(err)
	}
	return false
}

func IsDir(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	return stat.IsDir()
}

func IsFile(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !stat.IsDir()
}
