package provider

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"strings"

	"io/ioutil"

	getter "github.com/hashicorp/go-getter"
	"github.com/kyma-project/kyma/components/helm-broker/internal/addon"
	"github.com/kyma-project/kyma/components/helm-broker/internal/assetstore"
	"github.com/mholt/archiver"
	exerr "github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/rand"
)

// GitGetter provides functionality for loading addon from any Git repository.
// TODO: Write unit tests for GitGetter
type GitGetter struct {
	underlying *getter.GitGetter

	dst          string
	idxPath      string
	addonDirPath string
	docsURL      string

	cli    assetstore.Client
	tmpDir string
}

// GitGetterConfiguration holds additional data for GitGetter
type GitGetterConfiguration struct {
	Cli    assetstore.Client
	TmpDir string
}

// NewGit returns new instance of GitGetter
func (g GitGetterConfiguration) NewGit(addr, src string) (RepositoryGetter, error) {
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
		tmpDir:       g.TmpDir,
		cli:          g.Cli,
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
func (g *GitGetter) AddonLoadInfo(name addon.Name, version addon.Version) (LoadType, string, error) {
	var (
		addonDirName = fmt.Sprintf("%s-%s", name, version)
		pathToAddon  = path.Join(g.dst, g.addonDirPath, addonDirName)
	)

	return DirectoryLoadType, pathToAddon, nil
}

// AddonDocURL returns url for addon documentation
func (g *GitGetter) AddonDocURL(name addon.Name, version addon.Version) (string, error) {
	var (
		addonDirName = fmt.Sprintf("%s-%s", name, version)
		pathToAddon  = path.Join(g.dst, g.addonDirPath, addonDirName)
		pathToDocs   = path.Join(pathToAddon, "/docs")
		pathToTgz    = fmt.Sprintf("%s/docs-%s.tgz", g.tmpDir, addonDirName)
	)

	_, err := os.Stat(pathToDocs)
	switch {
	case err == nil:
	case os.IsNotExist(err):
		return "", nil
	default:
		return "", exerr.Wrap(err, "while checking if doc exists")
	}

	err = archiver.Archive([]string{pathToDocs}, pathToTgz)
	if err != nil {
		return "", exerr.Wrapf(err, "while creating archive '%s'", pathToTgz)
	}

	file, err := os.Open(pathToTgz)
	if err != nil {
		return "", exerr.Wrapf(err, "while opening file '%s'", pathToTgz)
	}
	defer func() {
		os.Remove(pathToTgz)
		file.Close()
	}()

	docs, err := ioutil.ReadAll(file)
	if err != nil {
		return "", exerr.Wrapf(err, "while reading file '%s'", file.Name())
	}

	uploaded, err := g.cli.Upload(pathToTgz, docs)
	if err != nil {
		return "", exerr.Wrapf(err, "while uploading Tgz '%s' to uploadService", pathToTgz)
	}

	return uploaded.RemotePath, nil
}
