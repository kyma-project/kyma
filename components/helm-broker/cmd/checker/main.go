package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/kyma-project/kyma/components/helm-broker/internal/addon"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	logger.Formatter = &logrus.TextFormatter{
		FullTimestamp: true,
	}

	dir, err := ioutil.TempDir("", "addonchecker")
	if err != nil {
		panic(err)
	}
	loader := addon.NewLoader(dir, logger)

	if len(os.Args) < 2 {
		fmt.Println("Provide path to a addon")
		os.Exit(1)
	}
	addonName := os.Args[1]

	fmt.Printf("==> Checking %s\n", addonName)
	_, _, err = loader.LoadDir(addonName)
	if err != nil {
		fmt.Printf("[ERROR] %v\n", err)
		os.Exit(1)
	}

	fmt.Print("Check OK")
}
