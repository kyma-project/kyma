package file

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func init() {
	http.DefaultClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
}

func CompareLocalAndRemote(localFilePath string, remoteFilePath string) (bool, error) {
	remoteBytes, err := download(remoteFilePath)
	if err != nil {
		return false, err
	}

	localBytes, err := load(localFilePath)
	if err != nil {
		return false, err
	}

	return bytes.Equal(localBytes, remoteBytes), nil
}

func Exists(url string) (bool, error) {
	res, err := http.Get(url)
	if err != nil {
		return false, errors.Wrapf(err, "while requesting file from URL %s", url)
	}
	defer func() {
		err := res.Body.Close()
		if err != nil {
			log.Print(err)
		}
	}()

	return res.StatusCode == http.StatusOK, nil
}

func Open(relativePath string) (*os.File, error) {
	absPath, err := filepath.Abs(relativePath)
	if err != nil {
		return nil, errors.Wrapf(err, "while constructing absolute path from %s", relativePath)
	}

	file, err := os.Open(absPath)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading file from path %s", absPath)
	}

	return file, nil
}

func download(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrapf(err, "while requesting file from URL %s", url)
	}
	defer func() {
		err := res.Body.Close()
		if err != nil {
			log.Print(err)
		}
	}()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Invalid status code while downloading file from URL %s: %d. Expected: %d", url, res.StatusCode, http.StatusOK)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "while reading response body while downloading file from URL %s", url)
	}

	return bytes, nil
}

func load(path string) ([]byte, error) {
	file, err := Open(path)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(file)
}

