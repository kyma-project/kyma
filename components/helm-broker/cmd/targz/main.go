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

	if err := archiveBundles(inDir, outDir); err != nil {
		panic(err)
	}
}

func archiveBundles(inDir, outDir string) error {
	bundles, err := ioutil.ReadDir(inDir)
	if err != nil {
		return errors.Wrap(err, "while reading input directory")
	}

	for _, bundleInfo := range bundles {
		if bundleInfo.IsDir() {
			bundlePath := filepath.Join(inDir, bundleInfo.Name())
			bundleContent, err := ioutil.ReadDir(bundlePath)
			if err != nil {
				return errors.Wrapf(err, "while reading bundle '%s'", bundlePath)
			}
			bundleFileNames := make([]string, len(bundleContent))

			for i, elem := range bundleContent {
				bundleFileNames[i] = filepath.Join(bundlePath, elem.Name())
			}

			tarGzFile := filepath.Join(outDir, bundleInfo.Name()+".tgz")
			err = archiver.TarGz.Make(tarGzFile, bundleFileNames)
			if err != nil {
				return errors.Wrapf(err, "while creating archive '%s'", tarGzFile)
			}
		}
	}
	return nil
}
