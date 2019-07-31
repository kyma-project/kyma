package main

import (
	"io/ioutil"
	"os"

	"github.com/pkg/errors"

	"fmt"
	"path/filepath"

	"github.com/kyma-project/kyma/components/helm-broker/cmd/targz/archiver"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: targz <input_directory> <output_directory>")
		return
	}
	inDir := os.Args[1]
	outDir := os.Args[2]

	if err := archiveAddons(inDir, outDir); err != nil {
		panic(err)
	}
}

func archiveAddons(inDir, outDir string) error {
	addons, err := ioutil.ReadDir(inDir)
	if err != nil {
		return errors.Wrap(err, "while reading input directory")
	}

	for _, addonInfo := range addons {
		if addonInfo.IsDir() {
			addonPath := filepath.Join(inDir, addonInfo.Name())
			addonContent, err := ioutil.ReadDir(addonPath)
			if err != nil {
				return errors.Wrapf(err, "while reading addon '%s'", addonPath)
			}
			addonFileNames := make([]string, len(addonContent))

			for i, elem := range addonContent {
				addonFileNames[i] = filepath.Join(addonPath, elem.Name())
			}

			tarGzFile := filepath.Join(outDir, addonInfo.Name()+".tgz")
			err = archiver.TarGz.Make(tarGzFile, addonFileNames)
			if err != nil {
				return errors.Wrapf(err, "while creating archive '%s'", tarGzFile)
			}
		}
	}
	return nil
}
