package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/kyma-project/kyma/components/helm-broker/internal/ybundle"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	logger.Formatter = &logrus.TextFormatter{
		FullTimestamp: true,
	}

	dir, err := ioutil.TempDir("", "bundlechecker")
	if err != nil {
		panic(err)
	}
	loader := ybundle.NewLoader(dir, logger.WithField("service", "bundle checker"))

	if len(os.Args) < 2 {
		fmt.Println("Provide path to a bundle")
		return
	}

	bundle, _, err := loader.LoadDir(os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	printBundleInfo(bundle)
}

func printBundleInfo(b *internal.Bundle) {
	fmt.Printf("Bundle name: %s\n", b.Name)
	fmt.Printf("Version: %s\n", b.Version.String())
	fmt.Printf("Description: %s\n", b.Description)
}
