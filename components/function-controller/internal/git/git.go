package git

import (
	"fmt"
	"strings"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/pkg/errors"
)

const (
	refsHeadsFormat = "refs/heads/%s"
	refsTagsFormat  = "refs/tags/%s"
)

//go:generate mockery -name=Client -output=automock -outpkg=automock -case=underscore
type Client interface {
	ListRefs(repoUrl string, auth transport.AuthMethod) ([]*plumbing.Reference, error)
	PlainClone(path string, isBare bool, options *gogit.CloneOptions) (*gogit.Repository, error)
}

type Git struct {
	client Client
}

type Options struct {
	URL       string
	Reference string
	Auth      *AuthOptions
}

type RepositoryAuthType string

const (
	RepositoryAuthBasic  RepositoryAuthType = "basic"
	RepositoryAuthSSHKey                    = "key"
)

func New() *Git {
	return &Git{client: &client{}}
}

func (g *Git) LastCommit(options Options) (string, error) {
	if plumbing.IsHash(options.Reference) {
		return options.Reference, nil
	}

	authMethod, err := options.Auth.ToAuthMethod()
	if err != nil {
		return "", errors.Wrap(err, "while parsing auth fields")
	}

	refs, err := g.client.ListRefs(options.URL, authMethod)
	if err != nil {
		return "", errors.Wrapf(err, "while listing remotes from repository: %s", options.URL)
	}

	headPattern := fmt.Sprintf(refsHeadsFormat, options.Reference)
	for _, elem := range refs {
		if strings.EqualFold(elem.Name().String(), headPattern) {
			return elem.Hash().String(), nil
		}
	}
	return "", fmt.Errorf("reference not found: %s", options.Reference)
}

func (g *Git) Clone(path string, options Options) (string, error) {
	authMethod, err := options.Auth.ToAuthMethod()
	if err != nil {
		return "", errors.Wrap(err, "while parsing auth fields")
	}

	repo, err := g.client.PlainClone(path, false, &gogit.CloneOptions{
		URL:  options.URL,
		Auth: authMethod,
	})
	if err != nil {
		return "", errors.Wrapf(err, "while cloning repository: %s", options.URL)
	}

	tree, err := repo.Worktree()
	if err != nil {
		return "", errors.Wrapf(err, "while getting WorkTree reference for repository: %s", options.URL)
	}

	commit, err := g.LastCommit(options)
	if err != nil {
		return "", err
	}

	err = tree.Checkout(&gogit.CheckoutOptions{
		Hash: plumbing.NewHash(commit),
	})
	if err != nil {
		return "", errors.Wrapf(err, "while checkout repository: %s, to commit: %s", options.URL, options.Reference)
	}

	head, err := repo.Head()
	if err != nil {
		return "", errors.Wrapf(err, "while getting HEAD reference for repository: %s", options.URL)
	}

	return head.Hash().String(), err
}
