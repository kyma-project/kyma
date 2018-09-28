package dex

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/pkg/errors"
)

func IsSCIEnabled() bool {
	config := readConfigurationFile()
	return strings.Contains(config, "SAP CI")
}

func readConfigurationFile() string {
	path := os.Getenv("DEX_CONFIGURATION_FILE")
	data, err := ioutil.ReadFile(path)
	switch {
	case os.IsNotExist(err):
		return ""
	case err != nil:
		log.Fatal(errors.Wrap(err, "while reading DEX configuration file"))
	}

	return string(data)
}
