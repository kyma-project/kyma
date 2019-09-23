package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var (
	validConfigPath         string
	invalidConfigPath       string
	invalidConfigFormatPath string
	emptyFilePath           string
)

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

func runTests(m *testing.M) int {
	testDataDir, err := ioutil.TempDir(".", "testdata")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	defer func() {
		err := os.RemoveAll(testDataDir)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}()

	err = setupTestData(testDataDir)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	return m.Run()
}

const (
	validConfig = `{
  "connectionConfig": {
    "connectorUrl": "https://connector.com",
    "token": "token"
  },
  "runtimeConfig": {
    "runtimeId": "runtimeId",
    "tenant": "tenant"
  }
}`
	invalidConfigData = `{
  "connectionConfig": {
    "invalidData": "invalid data"
  },
  "runtimeConfig": {
    "invalidData": "abcd-efgh-ijkl"
  }
}`
	invalidConfigFormat = `invalid format`
)

func setupTestData(testDataDir string) error {

	validConfigPath = filepath.Join(testDataDir, "valid-config.json")
	err := ioutil.WriteFile(validConfigPath, []byte(validConfig), os.ModePerm)
	if err != nil {
		return err
	}

	invalidConfigPath = filepath.Join(testDataDir, "invalid-config.json")
	err = ioutil.WriteFile(invalidConfigPath, []byte(invalidConfigData), os.ModePerm)
	if err != nil {
		return err
	}

	invalidConfigFormatPath = filepath.Join(testDataDir, "invalid-config-format.json")
	err = ioutil.WriteFile(invalidConfigFormatPath, []byte(invalidConfigFormat), os.ModePerm)
	if err != nil {
		return err
	}

	emptyFilePath = filepath.Join(testDataDir, "empty-file.json")
	err = ioutil.WriteFile(emptyFilePath, []byte(""), os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}
