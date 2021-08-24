package git

import "C"
import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	git2go "github.com/libgit2/git2go/v31"
	"github.com/pkg/errors"
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
	cloneRepo(options Options, outputPath string) (*git2go.Repository, error)
}

type Git2GoClient struct {
	cloner
}

func NewGit2Go() *Git2GoClient {
	return &Git2GoClient{
		cloner: &git2goCloner{},
	}
}

func (g *Git2GoClient) LastCommit(options Options) (string, error) {
	//commit
	_, err := git2go.NewOid(options.Reference)
	if err == nil {
		return options.Reference, nil
	}

	tmpPath, err := ioutil.TempDir(tempDir, "fn-git")
	if err != nil {
		return "", err
	}
	defer removeDir(tmpPath)

	repo, err := g.cloner.cloneRepo(options, tmpPath)
	if err != nil {
		return "", errors.Wrap(err, "while cloning the repository")
	}

	//branch
	ref, err := g.lookupBranch(repo, options.Reference)
	if err == nil {
		return ref.Target().String(), nil
	}
	if !git2go.IsErrorCode(err, git2go.ErrNotFound) {
		return "", err
	}

	//tag
	ref, err = repo.References.Dwim(options.Reference)
	if err == nil {
		return ref.Target().String(), nil
	}
	if !git2go.IsErrorCode(err, git2go.ErrNotFound) {
		return "", errors.Wrap(err, "while lookup branch")
	}
	return "", errors.Errorf("Could find commit,branch or tag with given ref: %s", options.Reference)
}

func (g *Git2GoClient) Clone(path string, options Options) (string, error) {
	repo, err := g.cloneRepo(options, path)
	if err != nil {
		return "", errors.Wrap(err, "while cloning the repository")
	}

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

func (g *Git2GoClient) lookupBranch(repo *git2go.Repository, branchName string) (*git2go.Reference, error) {
	iter, err := repo.NewReferenceIterator()
	if err != nil {
		return nil, err
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

func authCallback(cred *git2go.Credential) func(url, username string, allowed_types git2go.CredentialType) (*git2go.Credential, error) {
	return func(url, username string, allowed_types git2go.CredentialType) (*git2go.Credential, error) {
		return cred, nil
	}
}

func sshCheckCallback() func(cert *git2go.Certificate, valid bool, hostname string) git2go.ErrorCode {
	return func(cert *git2go.Certificate, valid bool, hostname string) git2go.ErrorCode {
		return git2go.ErrOk
	}
}

func removeDir(path string) {
	if os.RemoveAll(path) != nil {
		log.Printf("Error while deleting directory: %s", path)
	}
}
