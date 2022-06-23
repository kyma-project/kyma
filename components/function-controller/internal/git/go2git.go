package git

import (
	"crypto/md5"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	git2go "github.com/libgit2/git2go/v31"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	tempDir          = "/tmp"
	branchRefPattern = "refs/remotes/origin"
)

type Options struct {
	URL       string
	Reference string
	Auth      *AuthOptions
}

type cloner interface {
	git2goClone(url, outputPath string, remoteCallbacks git2go.RemoteCallbacks) (*git2go.Repository, error)
}

type fetcher interface {
	git2goFetch(url, outputPath string, remoteCallbacks git2go.RemoteCallbacks) (*git2go.Repository, error)
}

type Git2GoClient struct {
	cloner
	fetcher
}

func NewGit2Go(logger *zap.SugaredLogger) *Git2GoClient {
	return &Git2GoClient{
		cloner:  &git2goCloner{},
		fetcher: &git2goFetcher{logger: logger},
	}
}

func mkRepoDir(options Options) (string, error) {
	nameHash := md5.Sum([]byte(options.URL))
	path := path.Join(tempDir, fmt.Sprintf("%x", nameHash))

	err := os.MkdirAll(path, 0700)
	return path, err
}

func (g *Git2GoClient) LastCommit(options Options) (string, error) {
	//commit
	_, err := git2go.NewOid(options.Reference)
	if err == nil {
		return options.Reference, nil
	}

	// FIXME: This is NOT thread safe. If we ever decide to go with more than one worker, we need to refactor this. But for now it's fine.
	repoDir, err := mkRepoDir(options)
	if err != nil {
		return "", errors.Wrap(err, "while creating temporary directory")
	}
	repo, err := g.fetchRepo(options, repoDir)
	if err != nil {
		return "", errors.Wrap(err, "while fetching the repository")
	}
	defer repo.Free()
	//branch
	ref, err := g.lookupBranch(repo, options.Reference)
	if err == nil {
		return ref.Target().String(), nil
	}
	if !git2go.IsErrorCode(err, git2go.ErrNotFound) {
		return "", errors.Wrap(err, "while lookup branch")
	}
	//tag
	commit, err := g.lookupTag(repo, options.Reference)
	if err != nil {
		return "", errors.Wrap(err, "while lookup tag")
	}
	return commit.Id().String(), nil
}

func (g *Git2GoClient) Clone(path string, options Options) (string, error) {
	repo, err := g.cloneRepo(options, path)
	if err != nil {
		return "", errors.Wrap(err, "while cloning the repository")
	}
	defer repo.Free()

	oid, err := git2go.NewOid(options.Reference)
	if err != nil {
		return "", errors.Wrap(err, "while creating oid from reference")
	}

	commit, err := repo.LookupCommit(oid)
	if err != nil {
		return "", errors.Wrap(err, "while lookup for commit")
	}

	err = repo.ResetToCommit(commit, git2go.ResetHard, &git2go.CheckoutOptions{})
	if err != nil {
		return "", errors.Wrap(err, "while resetting to commit")
	}

	ref, err := repo.Head()
	if err != nil {
		return "", errors.Wrap(err, "while getting head")
	}

	return ref.Target().String(), nil
}

func (g *Git2GoClient) cloneRepo(opts Options, path string) (*git2go.Repository, error) {
	authCallbacks, err := GetAuth(opts.Auth)
	if err != nil {
		return nil, errors.Wrap(err, "while getting authentication opts")
	}
	return g.cloner.git2goClone(opts.URL, path, authCallbacks)
}
func (g *Git2GoClient) fetchRepo(opts Options, path string) (*git2go.Repository, error) {
	authCallbacks, err := GetAuth(opts.Auth)
	if err != nil {
		return nil, errors.Wrap(err, "while getting authentication opts")
	}
	return g.fetcher.git2goFetch(opts.URL, path, authCallbacks)
}

func (g *Git2GoClient) lookupBranch(repo *git2go.Repository, branchName string) (*git2go.Reference, error) {
	iter, err := repo.NewReferenceIterator()
	if err != nil {
		return nil, errors.Wrap(err, "while creating reference iterator")
	}
	for {
		item, err := iter.Next()
		if err != nil {
			if git2go.IsErrorCode(err, git2go.ErrorCodeIterOver) {
				return nil, git2go.MakeGitError2(int(git2go.ErrorCodeNotFound))
			}
			return nil, errors.Wrap(err, "while listing reference")
		}
		if g.isBranch(item, branchName) {
			return item, nil
		}
	}
}

func (g *Git2GoClient) isBranch(ref *git2go.Reference, branchName string) bool {
	if strings.Contains(ref.Name(), branchRefPattern) {
		splittedName := strings.Split(ref.Name(), "/")
		if len(splittedName) < 4 {
			return false
		}
		return splittedName[3] == branchName
	}
	return false
}

/*
Some repositories like bitbucket set tags in different way.
The tag has reference to object, not to commit.
Using this reference we can checkout head to it. From head we can extract commit id.
This method will also works with repositories like GitLab in which the tag is reference to the commit.
The reference has the same id as commit and won't produce errors
*/
func (g *Git2GoClient) lookupTag(repo *git2go.Repository, tagName string) (*git2go.Commit, error) {
	ref, err := repo.References.Dwim(tagName)
	if err != nil {
		if git2go.IsErrorCode(err, git2go.ErrorCodeNotFound) {
			return nil, err
		}
		return nil, errors.Wrap(err, "while creating dwim from tag name")
	}

	if err = repo.SetHeadDetached(ref.Target()); err != nil {
		return nil, errors.Wrapf(err, "while checkout to ref: %s", ref.Target().String())
	}
	head, err := repo.Head()
	if err != nil {
		return nil, errors.Wrap(err, "while getting head")
	}

	commit, err := repo.LookupCommit(head.Target())
	if err != nil {
		return nil, errors.Wrap(err, "while getting commit from head")
	}
	return commit, nil
}

func removeDir(path string) {
	if os.RemoveAll(path) != nil {
		log.Printf("Error while deleting directory: %s", path)
	}
}
