package main

import (
	"flag"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/kyma-project/kyma/components/helm-broker/internal/addon"
)

func main() {
	srcDir := flag.String("s", "", "Source directory with repository.")
	dstDir := flag.String("d", "", "Destination directory where index.yaml is saved. If not set than file is printed to stdout.")
	flag.Parse()

	l := logrus.New()
	l.Formatter = &logrus.TextFormatter{
		FullTimestamp: true,
	}

	if srcDir == nil {
		l.Panicln("src dir must be defined")
	}

	srcObjs, err := ioutil.ReadDir(*srcDir)
	if err != nil {
		l.Panicln(errors.Wrap(err, "while listing source dir"))
	}

	var loader *addon.Loader
	yLog := l.WithField("service", "addon checker")

	removeTempDir := func(path string) {
		if err := os.RemoveAll(path); err != nil {
			l.Panicln(errors.Wrap(err, "while removing temp dir"))
		}
	}

	processSingleAddon := func(fullPath string) *internal.Addon {
		tmpDir, err := ioutil.TempDir("", "indexbuilder")
		if err != nil {
			l.Panicln(errors.Wrap(err, "while creating temp dir"))
		}
		defer removeTempDir(tmpDir)

		loader = addon.NewLoader(tmpDir, yLog)
		b, _, err := loader.LoadDir(fullPath)
		if err != nil {
			l.Panicln(errors.Wrap(err, "while loading addon"))
		}

		return b
	}

	var allAddons []*internal.Addon

	for _, oi := range srcObjs {
		if !oi.IsDir() {
			continue
		}

		fPath := path.Join(*srcDir, oi.Name())
		b := processSingleAddon(fPath)
		allAddons = append(allAddons, b)
	}

	var dst io.Writer
	if dstDir == nil || *dstDir == "" {
		dst = os.Stdout
	} else {
		fullFilename := path.Join(*dstDir, "index.yaml")
		f, err := os.OpenFile(fullFilename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			l.Panicln(errors.Wrapf(err, "while opening file to write (file: %s)", fullFilename))
		}
		dst = f
	}

	if err := render(allAddons, dst); err != nil {
		l.Panicln(errors.Wrap(err, "while rendering index"))
	}
}
