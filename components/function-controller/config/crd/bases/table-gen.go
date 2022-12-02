package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
  v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
  "path/filepath"
)

type Config map[string]string

func main() {
    x :=
	filename, _ := filepath.Abs("/Users/I567085/src/2022jul/kyma/components/function-controller/config/crd/bases/serverless.kyma-project.io_functions.yaml")
	yamlFile, err := ioutil.ReadFile(filename)

	if err != nil {
		panic(err)
	}

	var config Config

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Value: %#v\n", config)
}
