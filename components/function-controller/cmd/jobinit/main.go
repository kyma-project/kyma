package main

import (
	"fmt"

	"github.com/kyma-project/kyma/components/function-controller/internal/gitops"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

const envPrefix = "APP"

type config struct {
	RepositoryUrl      string
	RepositoryCommit   string
	MountPath          string `envconfig:"default=/workspace/src/"`
	RepositoryUsername string `envconfig:"optional"`
	RepositoryPassword string `envconfig:"optional"`
}

func main() {
	cfg := config{}
	if err := envconfig.InitWithPrefix(&cfg, envPrefix); err != nil {
		panic(errors.Wrap(err, "while reading env variables"))
	}

	operator := gitops.NewOperator()
	auth := operator.ConvertToMap(cfg.RepositoryUsername, cfg.RepositoryPassword)

	commit, err := operator.CloneRepoFromCommit(cfg.MountPath, cfg.RepositoryUrl, cfg.RepositoryCommit, auth)
	if err != nil {
		panic(errors.Wrapf(err, "while cloning repository: %s, from commit: %s", cfg.RepositoryUrl, cfg.RepositoryCommit))
	}

	fmt.Printf("Cloned repository: %s, from commit: %s", cfg.RepositoryUrl, commit)
}
