package config

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path"
)

const fileDir = "/kyma/installation"
const fileName = "versions.yaml"

// LoadComponentsVersionsYaml reads the versions.yaml file if it exists in the /kyma/installer dir
func LoadComponentsVersionsYaml() (*bytes.Buffer, error) {

	filePath := path.Join(fileDir, fileName)
	_, err := os.Stat(filePath)
	if err != nil {

		if os.IsNotExist(err) {
			log.Printf("File %v does not exist. Error: %v", fileName, err)
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
