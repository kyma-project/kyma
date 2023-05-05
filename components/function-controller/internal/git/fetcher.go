package git

import (
	git2go "github.com/libgit2/git2go/v34"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type git2goFetcher struct {
	logger *zap.SugaredLogger
}

func (g *git2goFetcher) git2goFetch(url, outputPath string, remoteCallbacks git2go.RemoteCallbacks) (*git2go.Repository, error) {
	repo, err := g.openInitRepo(outputPath)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing/opening repository")
	}

	remote, err := g.lookupCreateRemote(repo, url, outputPath)
	if err != nil {
		return nil, errors.Wrap(err, "while creating/using remote")
	}
	defer remote.Free()

	err = remote.Fetch(nil,
		&git2go.FetchOptions{
			RemoteCallbacks: remoteCallbacks,
			DownloadTags:    git2go.DownloadTagsAll,
		}, "")
	if err != nil {
		return nil, errors.Wrap(err, "while fetching remote")
	}
	return repo, nil
}

func (g *git2goFetcher) openInitRepo(outputPath string) (*git2go.Repository, error) {
	var repo *git2go.Repository
	var err error
	repo, err = git2go.OpenRepository(outputPath)
	if err == nil {
		return repo, nil
	}
	g.logger.Errorf("failed to open existing repo at [%s]: %v", outputPath, err)
	return git2go.InitRepository(outputPath, true)
}

func (g *git2goFetcher) lookupCreateRemote(repo *git2go.Repository, url, outputPath string) (*git2go.Remote, error) {
	remote, err := repo.Remotes.Lookup("origin")
	if err == nil {
		return remote, nil
	}
	g.logger.Errorf("failed to use existing origin remote at [%s]: %v", outputPath, err)
	return repo.Remotes.Create("origin", url)
}
