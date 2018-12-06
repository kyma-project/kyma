package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/kyma-project/kyma/tools/docsbuilder/internal/content"
	"github.com/kyma-project/kyma/tools/docsbuilder/internal/docker"
	"github.com/vrischmann/envconfig"
)

type config struct {
	DocsDir       string `envconfig:"default=../../docs"`
	DocsBuildFile string `envconfig:"default=../../docs/docs-build.yaml"`
	Docker        dockerConfig
}

type dockerConfig struct {
	DockerfilePath string `envconfig:"default=../../docs/Dockerfile,DOCKERFILE_PATH"`
	ImageTag       string `envconfig:"default=latest"`
	ImageSuffix    string `envconfig:"default=docs"`
	ImagePrefix    string `envconfig:"optional"`
	PushImages     bool   `envconfig:"default=true"`
}

func main() {
	cfg, err := loadConfig("APP")

	docs, err := content.Read(cfg.DocsBuildFile)
	if err != nil {
		log.Fatal(err)
	}

	dockerfilePath, err := filepath.Abs(cfg.Docker.DockerfilePath)
	if err != nil {
		log.Fatal(err)
	}

	cfg.Docker.DockerfilePath = dockerfilePath

	additionalBuildArgs := fmt.Sprintf("--label version=%s --label component=docs", cfg.Docker.ImageTag)

	for _, doc := range docs {
		imageName := fmt.Sprintf("%s%s-%s:%s", cfg.Docker.ImagePrefix, doc.Name, cfg.Docker.ImageSuffix, cfg.Docker.ImageTag)

		imageCfg := &docker.ImageBuildConfig{
			Name:                imageName,
			BuildDirectory:      content.ConstructPath(doc, cfg.DocsDir),
			DockerfilePath:      dockerfilePath,
			AdditionalBuildArgs: additionalBuildArgs,
		}

		log.Printf("\n>>> Building image %s in %s...\n\n", imageCfg.Name, imageCfg.BuildDirectory)
		buildOut, err := docker.Build(imageCfg)
		fmt.Print(buildOut)
		if err != nil {
			log.Fatal(err)
		}

		if !cfg.Docker.PushImages {
			log.Println("Skipping pushing image...")
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
