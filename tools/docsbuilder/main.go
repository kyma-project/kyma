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
	Docker        docker.Config
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
		imageCfg := &docker.ImageBuildConfig{
			Name:                docker.ImageName(doc.Name, cfg.Docker),
			BuildDirectory:      content.ConstructPath(doc, cfg.DocsDir),
			DockerfilePath:      dockerfilePath,
			AdditionalBuildArgs: additionalBuildArgs,
		}

		log.Printf("\n>>> Building image %s in %s...\n\n", imageCfg.Name, imageCfg.BuildDirectory)
		buildOut, err := docker.Build(imageCfg)
		log.Print(buildOut)
		if err != nil {
			log.Fatal(err)
		}

		if !cfg.Docker.PushImages {
			log.Println("Skipping pushing image...")
			continue
		}

		log.Printf("\n>>> Pushing image %s...\n\n", imageCfg.Name)
		pushOut, err := docker.Push(imageCfg.Name)
		log.Print(pushOut)
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
