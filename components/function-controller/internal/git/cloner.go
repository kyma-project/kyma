package git

import (
	git2go "github.com/libgit2/git2go/v34"
)

type git2goCloner struct {
}

func (g *git2goCloner) git2goClone(url, outputPath string, remoteCallbacks git2go.RemoteCallbacks) (*git2go.Repository, error) {
	return git2go.Clone(url, outputPath, &git2go.CloneOptions{
		FetchOptions: git2go.FetchOptions{
			RemoteCallbacks: remoteCallbacks,
			DownloadTags:    git2go.DownloadTagsAll,
		},
	})
}
