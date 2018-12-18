package docker

import (
	"fmt"

	"github.com/kyma-project/kyma/tools/docsbuilder/internal/sh"
	"github.com/pkg/errors"
)

// Yes, I'm aware that we should use here Docker SDK.
// But it is just a temporary solution, so...
// It is still much better what we used to have.

// Config contains configuration for Docker image build and push
type Config struct {
	DockerfilePath string `envconfig:"default=../../docs/Dockerfile,DOCKERFILE_PATH"`
	ImageTag       string `envconfig:"default=latest"`
	ImageSuffix    string `envconfig:"default=-docs"`
	ImagePrefix    string `envconfig:"optional"`
	PushImages     bool   `envconfig:"default=true"`
}

// ImageBuildConfig contains configuration for building Docker images
type ImageBuildConfig struct {
	Name                string
	BuildDirectory      string
	DockerfilePath      string
	AdditionalBuildArgs string
}

// ImageName constructs final name of Docker image
func ImageName(docName string, cfg Config) string {
	return fmt.Sprintf("%s%s%s:%s", cfg.ImagePrefix, docName, cfg.ImageSuffix, cfg.ImageTag)
}

// Build builds Docker image
func Build(cfg *ImageBuildConfig) (string, error) {
	command := buildCommand(cfg)
	out, err := sh.RunInDir(command, cfg.BuildDirectory)
	if err != nil {
		return out, errors.Wrapf(err, "while executing %s", command)
	}

	return out, nil
}

// Push pushes Docker image
func Push(imageName string) (string, error) {
	command := pushCommand(imageName)
	out, err := sh.Run(command)
	if err != nil {
		return out, errors.Wrapf(err, "while executing %s", command)
	}

	return out, nil
}

func buildCommand(cfg *ImageBuildConfig) string {
	return fmt.Sprintf(`cat %s | docker build -f - . -t %s %s`, cfg.DockerfilePath, cfg.Name, cfg.AdditionalBuildArgs)
}

func pushCommand(imageName string) string {
	return fmt.Sprintf("docker push %s", imageName)
}
