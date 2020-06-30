package gitops

import (
	"fmt"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/pkg/errors"
)

const (
	usernameKey = "username"
	passwordKey = "password"
)

//go:generate mockery -name=GitInterface -output=automock -outpkg=automock -case=underscore
type GitInterface interface {
	Clone(s storage.Storer, worktree billy.Filesystem, o *git.CloneOptions) (*git.Repository, error)
}

type Config struct {
	RepoUrl      string
	Branch       string
	ActualCommit string
	BaseDir      string
	Secret       map[string]string
}

type Operator struct {
	gitOperator GitInterface
}

func NewOperator() *Operator {
	return &Operator{gitOperator: NewGit()}
}

func (g *Operator) CheckBranchChanges(config Config) (commitHash string, changesOccurred bool, err error) {
	auth, err := convertToBasicAuth(config.Secret)
	if err != nil {
		return commitHash, changesOccurred, errors.Wrap(err, "while parsing auth fields")
	}

	repo, err := g.gitOperator.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL:           config.RepoUrl,
		ReferenceName: plumbing.NewBranchReferenceName(config.Branch),
		Auth:          auth,
		SingleBranch:  true,
		NoCheckout:    true,
		Tags:          git.NoTags,
	})
	if err != nil {
		return commitHash, changesOccurred, errors.Wrapf(err, "while cloning repository: %s, branch: %s", config.RepoUrl, config.Branch)
	}

	head, err := repo.Head()
	if err != nil {
		return commitHash, changesOccurred, errors.Wrapf(err, "while getting HEAD reference for repository: %s, branch: %s", config.RepoUrl, config.Branch)
	}

	commitHash = head.Hash().String()

	if commitHash != config.ActualCommit {
		changesOccurred = true
	}

	return commitHash, changesOccurred, nil
}

func convertToBasicAuth(secret map[string]string) (*http.BasicAuth, error) {
	if secret == nil {
		return &http.BasicAuth{}, nil
	}

	username, ok := secret[usernameKey]
	if !ok {
		return nil, fmt.Errorf("missing field %s", usernameKey)
	}

	password, ok := secret[passwordKey]
	if !ok {
		return nil, fmt.Errorf("missing field %s", passwordKey)
	}

	return &http.BasicAuth{
		Username: username,
		Password: password,
	}, nil
}
