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

// GitGetter provides functionality for loading bundle from any Git repository.
type GitGetter struct {
	underlying *getter.GitGetter

	dst           string
	idxPath       string
	bundleDirPath string
}

// NewGit returns new instance of GitGetter
func NewGit(addr, dst string) (RepositoryGetter, error) {
	gitAddr, indexPath := getter.SourceDirSubdir(addr)
	if indexPath == "" {
		return nil, errors.New("index path needs to be provided! Check documentation")
	}

	ru, err := url.Parse(gitAddr)
	if err != nil {
		return nil, err
	}

	hashicorpGitGetter := &getter.GitGetter{}
	if err = hashicorpGitGetter.Get(dst, ru); err != nil {
		return nil, err
	}

	return &GitGetter{
		underlying:    hashicorpGitGetter,
		dst:           path.Join(dst, rand.String(10)),
		idxPath:       indexPath,
		bundleDirPath: strings.TrimRight(indexPath, path.Base(indexPath)),
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

// BundleLoadInfo returns information how to load bundle
// TODO: add feature for uploading documentation: https://github.com/kyma-project/kyma/issues/5040
func (g *GitGetter) BundleLoadInfo(name addon.Name, version addon.Version) (LoadType, string, error) {
	var (
		bundleDirName = fmt.Sprintf("%s-%s", name, version)
		pathToBundle  = path.Join(g.dst, g.bundleDirPath, bundleDirName)
	)

	return DirectoryLoadType, pathToBundle, nil
}

// BundleDocURL returns url for bundle documentation
// TODO: will be implemented in: https://github.com/kyma-project/kyma/issues/5040
func (g *GitGetter) BundleDocURL(name addon.Name, version addon.Version) string {
	return ""
}
