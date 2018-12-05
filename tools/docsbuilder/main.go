package main

import (
	"fmt"
	"github.com/kyma-project/kyma/tools/docsbuilder/internal/sh"
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
	const dockerPushRoot = "test-"
	const dockerImageTag = "latest"
	const dockerImageSuffix = "docs"
	const docsDir = "../../docs"

	docsBuildPath := fmt.Sprintf("%s/docs-build.yaml", docsDir)
	dockerfilePath := fmt.Sprintf("%s/Dockerfile", docsDir)

	docs, err := readDocs(docsBuildPath)
	if err != nil {
		log.Fatal(err)
	}

	dockerImageComponentLabel := "component=docs"
	dockerImageVersionLabel := fmt.Sprintf("version=%s", dockerImageTag)
	imageLabels := fmt.Sprintf("--label %s --label %s", dockerImageVersionLabel, dockerImageComponentLabel)

	for _, doc := range docs {
		imageName := fmt.Sprintf("%s%s-%s:%s", dockerPushRoot, doc.Name, dockerImageSuffix, dockerImageTag)
		buildCommand := fmt.Sprintf("cat %s | docker build -f - . -t %s %s", dockerfilePath, imageName, imageLabels)

		// build image
		log.Printf("> Building image %s...\n\n", imageName)

		docDir := doc.Name
		if doc.Directory != "" {
			docDir = doc.Directory
		}

		dir := fmt.Sprintf("%s/%s", docsDir, docDir)

		out, err := sh.RunInDir(buildCommand, dir)
		fmt.Print(out)
		if err != nil {
			log.Fatal(errors.Wrapf(err, "while executing %s", buildCommand))
		}


		// push image

		log.Printf("> Pushing image %s...\n\n", imageName)
		pushCommand := fmt.Sprintf("docker push %s", imageName)
		pushOut, err := sh.Run(pushCommand)
		fmt.Print(pushOut)

		if err != nil {
			log.Fatal(errors.Wrapf(err, "while executing %s", buildCommand))
		}
	}
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
