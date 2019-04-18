package loader

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	pkgPath "github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/path"
	"github.com/pkg/errors"
)

func (l *loader) loadPackage(src, name, filter string) (string, []string, error) {
	basePath, err := ioutil.TempDir(l.temporaryDir, name)
	if err != nil {
		return "", nil, err
	}

	archiveDir, err := ioutil.TempDir(l.temporaryDir, name)
	if err != nil {
		return "", nil, err
	}
	defer l.Clean(archiveDir)

	fileName := l.fileName(src)
	archivePath := filepath.Join(archiveDir, fileName)

	unpack, err := l.selectEngine(fileName)
	if err != nil {
		return "", nil, err
	}

	if err := l.download(archivePath, src); err != nil {
		return "", nil, err
	}

	files, err := unpack(archivePath, basePath)
	if err != nil {
		return "", nil, err
	}

	files, err = pkgPath.Filter(files, filter)
	if err != nil {
		return "", nil, err
	}

	return basePath, files, nil
}

func (l *loader) selectEngine(filename string) (func(src, dst string) ([]string, error), error) {
	extension := strings.ToLower(filepath.Ext(filename))

	switch {
	case extension == ".zip":
		return l.unpackZIP, nil
	case extension == ".tar" || extension == ".tgz" || strings.HasSuffix(strings.ToLower(filename), ".tar.gz"):
		return l.unpackTAR, nil
	}

	return nil, fmt.Errorf("not supported file type %s", extension)
}

func (l *loader) unpackTAR(src, dst string) ([]string, error) {
	var filenames []string
	file, err := os.Open(src)
	if err != nil {
		return nil, errors.Wrap(err, "while opening archive")
	}
	defer file.Close()

	var reader io.Reader
	extension := strings.ToLower(filepath.Ext(src))
	if extension == ".gz" || extension == ".gzip" || extension == ".tgz" {
		reader, err = gzip.NewReader(file)
		if err != nil {
			return nil, errors.Wrap(err, "while creating GZIP reader")
		}
	} else {
		reader = file
	}

	tarReader := tar.NewReader(reader)

unpack:
	for {
		header, err := tarReader.Next()
		switch {
		case err == io.EOF:
			break unpack
		case err != nil:
			return nil, errors.Wrap(err, "while unpacking archive")
		}

		target := filepath.Join(dst, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := l.createDir(target); err != nil {
				return nil, errors.Wrap(err, "while creating directory")
			}
		case tar.TypeReg:
			filenames = append(filenames, header.Name)

			if err := l.createFile(tarReader, target, header.Mode); err != nil {
				return nil, err
			}
		}
	}

	return filenames, nil
}

func (l *loader) unpackZIP(src, dest string) ([]string, error) {
	var filenames []string

	zipReader, err := zip.OpenReader(src)
	if err != nil {
		return nil, errors.Wrap(err, "while opening archive")
	}
	defer zipReader.Close()

	for _, file := range zipReader.File {
		path, isDir, err := l.handleZIPEntry(file, dest)
		if err != nil {
			return nil, errors.Wrap(err, "while handling ZIP entry")
		}

		if !isDir {
			filenames = append(filenames, path)
		}
	}

	return filenames, nil
}

func (l *loader) handleZIPEntry(file *zip.File, dst string) (string, bool, error) {
	fileReader, err := file.Open()
	if err != nil {
		return "", false, err
	}
	defer fileReader.Close()

	path := filepath.Join(dst, file.Name)

	if !strings.HasPrefix(path, filepath.Clean(dst)+string(os.PathSeparator)) {
		return "", false, fmt.Errorf("%s: illegal file path", path)
	}

	if file.FileInfo().IsDir() {
		if err := l.createDir(path); err != nil {
			return "", false, errors.Wrap(err, "while creating directory")
		}
	} else {
		if err := l.createFile(fileReader, path, int64(file.Mode())); err != nil {
			return "", false, err
		}
	}

	return file.Name, file.FileInfo().IsDir(), nil
}

func (l *loader) createFile(src io.Reader, dst string, mode int64) error {
	if err := l.createDir(filepath.Dir(dst)); err != nil {
		return errors.Wrap(err, "while creating directory")
	}

	outFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(mode))
	if err != nil {
		return errors.Wrap(err, "while opening file")
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, src)
	if err != nil {
		return errors.Wrap(err, "while copying data to file")
	}

	return nil
}

func (l *loader) createDir(dst string) error {
	return os.MkdirAll(dst, os.ModePerm)
}
