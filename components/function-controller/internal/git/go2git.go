package git

import "C"
import (
	"fmt"
	git2go "github.com/libgit2/git2go/v31"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

const (
	tempDir          = "/tmp"
	branchRefPattern = "refs/remotes/origin"
)

type git2goCloner struct {
}

func (g *git2goCloner) cloneRepo(options Options, outputPath string) (*git2go.Repository, error) {
	authCallbacks, err := getAuth(options.Auth)
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

func getAuth(options *AuthOptions) (git2go.RemoteCallbacks, error) {
	if options == nil {
		return git2go.RemoteCallbacks{}, nil
	}

	switch authType := options.Type; authType {
	case RepositoryAuthBasic:
		{
			username, ok := options.Credentials[UsernameKey]
			if !ok {
				return git2go.RemoteCallbacks{}, fmt.Errorf("missing field %s", UsernameKey)

			}
			password, ok := options.Credentials[PasswordKey]
			if !ok {
				return git2go.RemoteCallbacks{}, fmt.Errorf("missing field %s", PasswordKey)
			}

			cred, err := git2go.NewCredentialUserpassPlaintext(username, password)
			if err != nil {
				return git2go.RemoteCallbacks{}, errors.Wrap(err, "while creating basic auth")
			}
			return git2go.RemoteCallbacks{
				CredentialsCallback: authCallback(cred),
			}, nil
		}
	case RepositoryAuthSSHKey:
		{
			key, ok := options.Credentials[KeyKey]
			if !ok {
				return git2go.RemoteCallbacks{}, fmt.Errorf("missing field %s", KeyKey)
			}
			passphrase, ok := options.Credentials[PasswordKey]
			if !ok {
				return git2go.RemoteCallbacks{}, fmt.Errorf("missing field %s", PasswordKey)
			}

			var err error
			if passphrase == "" {
				_, err = ssh.ParsePrivateKey([]byte(key))
			} else {
				_, err = ssh.ParseRawPrivateKeyWithPassphrase([]byte(key), []byte(passphrase))
			}

			if err != nil {
				return git2go.RemoteCallbacks{}, errors.Wrapf(err, "while validation of key with passphrase set to: %t", passphrase != "")
			}
			cred, err := git2go.NewCredentialSSHKeyFromMemory("git", "", key, passphrase)
			if err != nil {
				return git2go.RemoteCallbacks{}, errors.Wrap(err, "while creating ssh credential in git2go")
			}
			return git2go.RemoteCallbacks{
				CredentialsCallback:      authCallback(cred),
				CertificateCheckCallback: sshCheckCallback(),
			}, nil

		}
	}
	return git2go.RemoteCallbacks{}, nil
}

func (g *Git2GoClient) lookupBranch(repo *git2go.Repository, branchName string) (*git2go.Reference, error) {
	iter, err := repo.NewReferenceIterator()
	if err != nil {
		return nil, err
	}
	for ; ; {
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
