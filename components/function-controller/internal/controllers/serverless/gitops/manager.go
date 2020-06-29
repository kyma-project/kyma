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
	usernameKey = "login"
	passwordKey = "token"
)

//go:generate mockery -name=GitOperator -output=automock -outpkg=automock -case=underscore
type GitOperator interface {
	Clone(s storage.Storer, worktree billy.Filesystem, o *git.CloneOptions) (*git.Repository, error)
}

type Manager struct {
	gitOperator GitOperator
}

func NewManager(operator GitOperator) (*Manager, error) {
	return &Manager{gitOperator: operator}, nil
}

func (g *Manager) GetLastCommit(repoUrl, branch string, secret map[string]interface{}) (string, error) {

	auth, err := convertToBasicAuth(secret)
	if err != nil {
		return "", errors.Wrap(err, "while parsing auth fields")
	}

	repo, err := g.gitOperator.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL:               repoUrl,
		ReferenceName:     plumbing.NewBranchReferenceName(branch),
		Auth:              auth,
		SingleBranch:      true,
		NoCheckout:        true,
		Tags:              git.NoTags,
	})
	if err != nil {
		return "", errors.Wrapf(err, "while cloning repository: %s, branch: %s", repoUrl, branch)
	}

	head, err := repo.Head()
	if err != nil {
		return "", errors.Wrapf(err, "while getting HEAD reference for repository: %s, branch: %s", repoUrl, branch)
	}

	cIter, err := repo.Log(&git.LogOptions{From: head.Hash()})
	if err != nil {
		return "", errors.Wrapf(err, "while listing commit history for repository: %s, branch: %s", repoUrl, branch)
	}

	commit, err := cIter.Next()
	if err != nil {
		return "", errors.Wrapf(err, "while getting last commit for repository: %s, branch: %s", repoUrl, branch)
	}

	return commit.Hash.String(), nil
}

func convertToBasicAuth(secret map[string]interface{}) (*http.BasicAuth, error) {
	if secret == nil {
		return nil, nil
	}

	username, ok := secret[usernameKey].(string)
	if !ok {
		return nil, fmt.Errorf("missing field %s", usernameKey)
	}

	password, ok := secret[passwordKey].(string)
	if !ok {
		return nil, fmt.Errorf("missing field %s", passwordKey)
	}

	return &http.BasicAuth{
		Username: username,
		Password: password,
	}, nil
}
