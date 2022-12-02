package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

type Config map[string]string

func main() {
	filename := "/Users/I567085/src/2022jul/kyma/components/function-controller/config/crd/bases/serverless.kyma-project.io_functions.yaml"
	stream, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	var obj v1.CustomResourceDefinition
	//var obj interface{}
	err = yaml.Unmarshal(stream, &obj)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Value: %#v\n", obj)
}
