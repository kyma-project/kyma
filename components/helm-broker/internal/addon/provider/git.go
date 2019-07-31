package provider

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"strings"

	getter "github.com/hashicorp/go-getter"
	"github.com/kyma-project/kyma/components/helm-broker/internal/addon"
	"k8s.io/apimachinery/pkg/util/rand"
)

// GitGetter provides functionality for loading addon from any Git repository.
type GitGetter struct {
	underlying *getter.GitGetter

	dst          string
	idxPath      string
	addonDirPath string
}

// NewGit returns new instance of GitGetter
func NewGit(addr, src string) (RepositoryGetter, error) {
	finalDst := path.Join(src, rand.String(10))
	gitAddr, indexPath := getter.SourceDirSubdir(addr)
	if indexPath == "" {
		return nil, errors.New("index path needs to be provided. Check documentation about using git protocol")
	}

	ru, err := url.Parse(gitAddr)
	if err != nil {
		return nil, err
	}

	hashicorpGitGetter := &getter.GitGetter{}
	if err = hashicorpGitGetter.Get(finalDst, ru); err != nil {
		return nil, err
	}

	return &GitGetter{
		underlying:   hashicorpGitGetter,
		dst:          finalDst,
		idxPath:      indexPath,
		addonDirPath: strings.TrimRight(indexPath, path.Base(indexPath)),
	}, nil
}

// Cleanup  removes folder where git repository was cloned.
func (g *GitGetter) Cleanup() error {
	return os.RemoveAll(g.dst)
}

// IndexReader returns index reader
func (g *GitGetter) IndexReader() (io.ReadCloser, error) {
	return os.Open(path.Join(g.dst, g.idxPath))
}

// AddonLoadInfo returns information how to load addon
// TODO: add feature for uploading documentation: https://github.com/kyma-project/kyma/issues/5040
func (g *GitGetter) AddonLoadInfo(name addon.Name, version addon.Version) (LoadType, string, error) {
	var (
		addonDirName = fmt.Sprintf("%s-%s", name, version)
		pathToAddon  = path.Join(g.dst, g.addonDirPath, addonDirName)
	)

	return DirectoryLoadType, pathToAddon, nil
}

// AddonDocURL returns url for addon documentation
// TODO: will be implemented in: https://github.com/kyma-project/kyma/issues/5040
func (g *GitGetter) AddonDocURL(name addon.Name, version addon.Version) string {
	return ""
}
