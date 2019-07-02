package overrides

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path"
)

const fileDir = "/kyma/installation"
const fileName = "versions.yaml"

// loadComponentsVersions reads the versions.yaml file if it exists in the /kyma/installation dir
func loadComponentsVersions() (*bytes.Buffer, error) {

	filePath := path.Join(fileDir, fileName)
	_, err := os.Stat(filePath)
	if err != nil {

		if os.IsNotExist(err) {
			return nil, nil
		}

		log.Println("Error reading file")
		return nil, err
	}

	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(contents), nil
}
