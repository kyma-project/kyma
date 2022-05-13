package git

import (
	git2go "github.com/libgit2/git2go/v31"
	"github.com/pkg/errors"
)

type git2goFetcher struct {
}

func (g *git2goFetcher) git2goFetch(url, outputPath string, remoteCallbacks git2go.RemoteCallbacks) (*git2go.Repository, error) {
	repo, err := git2go.InitRepository(outputPath, true)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing repository")
	}

	remote, err := repo.Remotes.Create("origin", url)
	if err != nil {
		return nil, errors.Wrap(err, "while creating remote")
	}
	defer remote.Free()
	err = remote.Fetch(nil, &git2go.FetchOptions{RemoteCallbacks: remoteCallbacks}, "")
	if err != nil {
		return nil, errors.Wrap(err, "while fetching remote")
	}
	return repo, nil
}
