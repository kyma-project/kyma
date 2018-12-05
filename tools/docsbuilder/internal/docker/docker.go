package docker

import (
	"fmt"

	"github.com/kyma-project/kyma/tools/docsbuilder/internal/sh"
	"github.com/pkg/errors"
)

// Yes, I'm aware that we should use here Docker SDK.
// But it is just a temporary solution, so...
// It is still much better what we used to have.

type ImageBuildConfig struct {
	Name                string
	BuildDirectory      string
	DockerfilePath      string
	AdditionalBuildArgs string
}

func Build(cfg *ImageBuildConfig) (string, error) {
	buildCommand := fmt.Sprintf(`cat %s | docker build -f - . -t %s %s`, cfg.DockerfilePath, cfg.Name, cfg.AdditionalBuildArgs)
	out, err := sh.RunInDir(buildCommand, cfg.BuildDirectory)
	if err != nil {
		return out, errors.Wrapf(err, "while executing %s", buildCommand)
	}

	return out, nil
}

func Push(imageName string) (string, error) {
	pushCommand := fmt.Sprintf("docker push %s", imageName)
	out, err := sh.Run(pushCommand)
	if err != nil {
		return out, errors.Wrapf(err, "while executing %s", pushCommand)
	}

	return out, nil
}
