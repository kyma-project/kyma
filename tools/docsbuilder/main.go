package main

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type Doc struct {
	Name      string `yaml:"name"`
	Directory string `yaml:"directory"`
}

func main() {

	//TODO: ENVs
	const docsDir = "docs"
	const dockerPushRoot = "test"
	const dockerImageTag = "latest"
	const dockerImageLabel = "component=docs"
	const dockerImageSuffix = docsDir
	const path = "../../docs/docs-build.yaml"



	docs, err := readDocs(path)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v", docs)

}

func readDocs(path string) ([]Doc, error) {
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "while reading file %s", path)
	}

	var docs []Doc
	err = yaml.Unmarshal(yamlFile, &docs)
	if err != nil {
		return nil, errors.Wrapf(err, "while unmarshalling content of the file %s", path)
	}

	return docs, nil
}
