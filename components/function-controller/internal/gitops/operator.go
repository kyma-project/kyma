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
	PlainClone(path string, isBare bool, options *git.CloneOptions) (*git.Repository, error)
}

type Config struct {
	RepoUrl string
	Branch  string
	BaseDir string
	Secret  map[string]string
}

type Operator struct {
	gitInterface GitInterface
}

func NewOperator() *Operator {
	return &Operator{gitInterface: NewGit()}
}

func (g *Operator) GetLastCommit(config Config) (commitHash string, err error) {
	auth, err := convertToBasicAuth(config.Secret)
	if err != nil {
		return commitHash, errors.Wrap(err, "while parsing auth fields")
	}

	repo, err := g.gitInterface.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL:           config.RepoUrl,
		ReferenceName: plumbing.NewBranchReferenceName(config.Branch),
		Auth:          auth,
		SingleBranch:  true,
		NoCheckout:    true,
		Tags:          git.NoTags,
	})
	if err != nil {
		return commitHash, errors.Wrapf(err, "while cloning repository: %s, branch: %s", config.RepoUrl, config.Branch)
	}

	head, err := repo.Head()
	if err != nil {
		return commitHash, errors.Wrapf(err, "while getting HEAD reference for repository: %s, branch: %s", config.RepoUrl, config.Branch)
	}

	commitHash = head.Hash().String()

	return commitHash, nil
}

func (o *Operator) CloneRepoFromCommit(path, repoUrl, commit string, auth map[string]string) (commitHash string, err error) {
	basicAuth, err := convertToBasicAuth(auth)
	if err != nil {
		return "", errors.Wrap(err, "while parsing auth fields")
	}

	repo, err := o.gitInterface.PlainClone(path, false, &git.CloneOptions{
		URL:  repoUrl,
		Auth: basicAuth,
	})
	if err != nil {
		return commitHash, errors.Wrapf(err, "while cloning repository: %s", repoUrl)
	}

	tree, err := repo.Worktree()
	if err != nil {
		return "", errors.Wrapf(err, "while getting WorkTree reference for repository: %s", repoUrl)
	}

	err = tree.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(commit),
	})
	if err != nil {
		return "", errors.Wrapf(err, "while checkout repository: %s, to commit: %s", repoUrl, commit)
	}

	head, err := repo.Head()
	if err != nil {
		return commitHash, errors.Wrapf(err, "while getting HEAD reference for repository: %s", repoUrl)
	}

	commitHash = head.Hash().String()

	return head.Hash().String(), err
}

func (o *Operator) ConvertToMap(username, password string) (auth map[string]string) {
	if username != "" && password != "" {
		auth = map[string]string{}
		auth[usernameKey] = username
		auth[passwordKey] = password
	}
	return auth
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
