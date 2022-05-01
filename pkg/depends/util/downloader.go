package util

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

func DownloadFile(url, dst string) error {
	if dir := filepath.Dir(dst); dir != "" {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
	}
	tmp := path.Join("/tmp", "chat.profanity.words.tmp.downloading")
	fl, err := os.Create(tmp)
	if err != nil {
		return err
	}
	defer fl.Close()

	rsp, err := http.Get(url)
	if err != nil {
		return err
	}
	if rsp.StatusCode != 200 {
		return errors.New("download failed")
	}
	defer rsp.Body.Close()
	if _, err = io.Copy(fl, rsp.Body); err != nil {
		return err
	}
	return os.Rename(tmp, dst)
}
