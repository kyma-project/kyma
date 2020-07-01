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
	fmt.Println("Start repo fetcher...")
	cfg := config{}
	if err := envconfig.InitWithPrefix(&cfg, envPrefix); err != nil {
		panic(errors.Wrap(err, "while reading env variables"))
	}

	fmt.Printf("Config: %s", cfg)

	fmt.Println("Create new gitops operator...")
	operator := gitops.NewOperator()

	//fmt.Println("Check for auth config...")
	//auth := operator.ConvertToMap(cfg.RepositoryUsername, cfg.RepositoryPassword)

	fmt.Printf("Clone repo from url: %s and commit: %s...\n", cfg.RepositoryUrl, cfg.RepositoryCommit)
	commit, err := operator.CloneRepoFromCommit(cfg.MountPath, cfg.RepositoryUrl, cfg.RepositoryCommit, nil/*auth*/)
	if err != nil {
		panic(errors.Wrapf(err, "while cloning repository: %s, from commit: %s", cfg.RepositoryUrl, cfg.RepositoryCommit))
	}

	fmt.Printf("Cloned repository: %s, from commit: %s", cfg.RepositoryUrl, commit)
}
