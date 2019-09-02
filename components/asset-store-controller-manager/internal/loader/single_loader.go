package loader

import (
	"path/filepath"
)

func (l *loader) loadSingle(src, name string) (string, []string, error) {
	basePath, err := l.ioutilTempDir(l.temporaryDir, name)
	if err != nil {
		return "", nil, err
	}

	fileName := l.fileName(src)
	destination := filepath.Join(basePath, fileName)
	err = l.download(destination, src)
	if err != nil {
		return "", nil, err
	}

	return basePath, []string{fileName}, nil
}
