package main

import (
	"fmt"
	"io/ioutil"
	"os"

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
	loader := ybundle.NewLoader(dir, logger)

	if len(os.Args) < 2 {
		fmt.Println("Provide path to a bundle")
		os.Exit(1)
	}
	bundleName := os.Args[1]

	fmt.Printf("==> Checking %s\n", bundleName)
	_, _, err = loader.LoadDir(bundleName)
	if err != nil {
		fmt.Printf("[ERROR] %v\n", err)
		os.Exit(1)
	}

	fmt.Print("Check OK")
}
