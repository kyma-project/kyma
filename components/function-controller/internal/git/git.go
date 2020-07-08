package git

import (
	"fmt"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/pkg/errors"
)

const refsHeadsPrefix = "refs/heads/%s"

//go:generate mockery -name=Client -output=automock -outpkg=automock -case=underscore
type Client interface {
	ListRefs(repoUrl string, auth transport.AuthMethod) ([]*plumbing.Reference, error)
	PlainClone(path string, isBare bool, options *gogit.CloneOptions) (*gogit.Repository, error)
}

type Git struct {
	client Client
}

func New() *Git {
	return &Git{client: newClient()}
}

func (g *Git) LastCommit(repoUrl, branch string, auth transport.AuthMethod) (commitHash string, err error) {
	refs, err := g.client.ListRefs(repoUrl, auth)
	if err != nil {
		return commitHash, errors.Wrapf(err, "while listing remotes from repository: %s", repoUrl)
	}

	pattern := fmt.Sprintf(refsHeadsPrefix, branch)
	for _, elem := range refs {
		if elem.Name().String() == pattern {
			commitHash = elem.Hash().String()
		}
	}
	if commitHash == "" {
		err = fmt.Errorf("branch %s don't exist with pattern %s", branch, pattern)
	}

	return commitHash, err
}

func (o *Git) Clone(path, repoUrl, commit string, auth transport.AuthMethod) (string, error) {
	repo, err := o.client.PlainClone(path, false, &gogit.CloneOptions{
		URL:  repoUrl,
		Auth: auth,
	})
	if err != nil {
		return "", errors.Wrapf(err, "while cloning repository: %s", repoUrl)
	}

	tree, err := repo.Worktree()
	if err != nil {
		return "", errors.Wrapf(err, "while getting WorkTree reference for repository: %s", repoUrl)
	}

	err = tree.Checkout(&gogit.CheckoutOptions{
		Hash: plumbing.NewHash(commit),
	})
	if err != nil {
		return "", errors.Wrapf(err, "while checkout repository: %s, to commit: %s", repoUrl, commit)
	}

	head, err := repo.Head()
	if err != nil {
		return "", errors.Wrapf(err, "while getting HEAD reference for repository: %s", repoUrl)
	}

	return head.Hash().String(), err
}
