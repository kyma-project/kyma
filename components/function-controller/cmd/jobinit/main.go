package main

import (
	"log"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/kyma-project/kyma/components/function-controller/internal/git"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

const envPrefix = "APP"

type config struct {
	RepositoryUrl      string
	RepositoryCommit   string
	MountPath          string `envconfig:"default=/workspace"`
	RepositoryUsername string `envconfig:"optional"`
	RepositoryPassword string `envconfig:"optional"`
}

func main() {
	log.Println("Start repo fetcher...")
	cfg := config{}
	if err := envconfig.InitWithPrefix(&cfg, envPrefix); err != nil {
		log.Fatalln("while reading env variables")
	}
	operator := git.New()

	log.Println("Check for auth config...")
	auth := &http.BasicAuth{
		Username: cfg.RepositoryUsername,
		Password: cfg.RepositoryPassword,
	}

	log.Printf("Clone repo from url: %s and commit: %s...\n", cfg.RepositoryUrl, cfg.RepositoryCommit)
	commit, err := operator.Clone(cfg.MountPath, cfg.RepositoryUrl, cfg.RepositoryCommit, auth)
	if err != nil {
		log.Fatalln(errors.Wrapf(err, "while cloning repository: %s, from commit: %s", cfg.RepositoryUrl, cfg.RepositoryCommit))
	}

	log.Printf("Cloned repository: %s, from commit: %s, to path: %s", cfg.RepositoryUrl, commit, cfg.MountPath)
}
