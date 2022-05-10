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

	remote, err := useRemote(repo, url)
	if err != nil {
		return nil, errors.Wrap(err, "while using remote")
	}
	defer remote.Free()

	err = remote.Fetch(nil, &git2go.FetchOptions{RemoteCallbacks: remoteCallbacks}, "")
	if err != nil {
		return nil, errors.Wrap(err, "while fetching remote")
	}

	return repo, nil
}

func useRemote(repo *git2go.Repository, url string) (*git2go.Remote, error) {
	remote, err := repo.Remotes.Lookup("origin")
	if err == nil {
		return remote, nil
	}
	if git2go.IsErrorCode(err, git2go.ErrNotFound) {
		return repo.Remotes.Create("origin", url)
	}
	return nil, errors.Wrap(err, "while looking up remote")
}
