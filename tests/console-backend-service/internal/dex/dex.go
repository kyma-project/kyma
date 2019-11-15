package dex

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/pkg/errors"
)

const sciEnabledMessage = "SCI is enabled"

func SkipTestIfSCIEnabled(t *testing.T) {
	if !isSCIEnabled() {
		return
	}

	t.Skip(sciEnabledMessage)
}

func ExitIfSCIEnabled() {
	if !isSCIEnabled() {
		return
	}

	log.Println(sciEnabledMessage)
	os.Exit(0)
}

func isSCIEnabled() bool {
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
