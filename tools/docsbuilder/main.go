package main

import (
	"fmt"
	"github.com/kyma-project/kyma/tools/docsbuilder/internal/content"
	"github.com/kyma-project/kyma/tools/docsbuilder/internal/docker"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"io/ioutil"
	"log"
)

type config struct {
	DocsDir       string `envconfig:"default=../../docs"`
	DocsBuildFile string `envconfig:"default=../../docs/docs-build.yaml"`
	Docker        dockerConfig
}

type dockerConfig struct {
	DockerfilePath string `envconfig:"default=../../docs/Dockerfile,DOCKERFILE_PATH"`
	ImageTag       string `envconfig:"default=latest"`
	ImageSuffix    string `envconfig:"default=-docs"`
	PushRoot       string `envconfig:"optional"`
}

func main() {
	cfg, err := loadConfig("APP")

	docs, err := content.Read(cfg.DocsBuildFile)
	if err != nil {
		log.Fatal(err)
	}

	additionalBuildArgs := fmt.Sprintf("--label version=%s --label component=docs", cfg.Docker.ImageTag)

	//dockerfile, err := loadTextFile(cfg.Docker.DockerfilePath)
	//if err != nil {
	//	log.Fatal(err)
	//}

	for _, doc := range docs {
		imageName := fmt.Sprintf("%s%s-%s:%s", cfg.Docker.PushRoot, doc.Name, cfg.Docker.ImageSuffix, cfg.Docker.ImageTag)

		imageCfg := &docker.ImageBuildConfig{
			Name:                imageName,
			BuildDirectory:      content.ConstructPath(doc, cfg.DocsDir),
			DockerfilePath:      cfg.Docker.DockerfilePath,
			AdditionalBuildArgs: additionalBuildArgs,
		}

		log.Printf("\n>>> Building image %s in %s...\n\n", imageCfg.Name, imageCfg.BuildDirectory)
		buildOut, err := docker.Build(imageCfg)
		fmt.Print(buildOut)
		if err != nil {
			log.Fatal(err)
		}

		if cfg.Docker.PushRoot == "" {
			log.Println("Empty 'PushRoot' config property. Skipping pushing image...")
			continue
		}

		log.Printf("\n>>> Pushing image %s...\n\n", imageCfg.Name)
		pushOut, err := docker.Push(imageCfg.Name)
		fmt.Print(pushOut)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func loadConfig(prefix string) (config, error) {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, prefix)
	return cfg, err
}

func loadTextFile(path string) (string, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return "", errors.Wrapf(err, "while reading file %s", path)
	}

	//regExp := regexp.MustCompile(`\r?\n`)
	//return regExp.ReplaceAllString(string(file), " "), nil

	return string(file), nil
}