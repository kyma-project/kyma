package git

import (
	git2go "github.com/libgit2/git2go/v31"
	"github.com/pkg/errors"
)

type git2goCloner struct {
}

func (g *git2goCloner) cloneRepo(options Options, outputPath string) (*git2go.Repository, error) {
	authCallbacks, err := GetAuth(options.Auth)
	if err != nil {
		return nil, errors.Wrap(err, "while getting authentication opts")
	}

	repo, err := git2go.Clone(options.URL, outputPath, &git2go.CloneOptions{
		FetchOptions: &git2go.FetchOptions{
			RemoteCallbacks: authCallbacks,
			DownloadTags:    git2go.DownloadTagsAll,
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "while cloning the repository")
	}
	return repo, nil
}
