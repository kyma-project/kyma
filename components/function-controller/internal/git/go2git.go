package git

import (
	"fmt"
	git2go "github.com/libgit2/git2go/v31"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"os"
)

const (
	tempDir          = "/tmp"
	tagRefPattern    = "refs/tags/%s"
	branchRefPattern = "refs/remotes/%s"
)

type Go2GitClient struct {
}

func NewGit2Go() *Go2GitClient {
	return &Go2GitClient{}
}

func (g *Go2GitClient) LastCommit(options Options) (string, error) {
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

	repo, err := g.cloneRepo(options, tmpPath)
	if err != nil {
		return "", errors.Wrap(err, "while cloning the repository")
	}

	//branch
	branch, err := repo.LookupBranch(fmt.Sprintf(branchRefPattern, options.Reference), git2go.BranchAll)
	if err == nil {
		return branch.Target().String(), nil
	}
	if !git2go.IsErrorCode(err, git2go.ErrNotFound) {
		return "", errors.Wrap(err, "while lookup branch")
	}

	//tag
	ref, err := repo.References.Dwim(fmt.Sprintf(tagRefPattern, options.Reference))
	if err == nil {
		return ref.Target().String(), nil
	}
	if !git2go.IsErrorCode(err, git2go.ErrNotFound) {
		return "", errors.Wrap(err, "while lookup branch")
	}
	return "", errors.Errorf("Could find commit/branch/tag with given ref: %s", options.Reference)

	//
	//log.Println(ref.Target().String())
	//
	//iter, err := repo.NewReferenceIterator()
	//if err != nil {
	//	return "", err
	//}
	//
	//log.Println("all references")
	//for ; ; {
	//	item, err := iter.Next()
	//	if err != nil {
	//		break
	//	}
	//
	//	log.Println(item.Name())
	//}
	//
	////Tags
	//log.Println("Tag")
	//
	//tags, err := repo.Tags.List()
	//if err != nil {
	//	return "", err
	//}
	//for tag, _ := range tags {
	//	log.Println(tag)
	//}
	//
	//ref, err := repo.References.Dwim(fmt.Sprintf("refs/tags/first"))
	//if err != nil {
	//	return "", err
	//}
	//
	//log.Println(ref.Target().String())
	//return "", nil
}

func (g *Go2GitClient) Clone(path string, options Options) (string, error) {
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

func (g *Go2GitClient) cloneRepo(options Options, outputPath string) (*git2go.Repository, error) {
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

func getAuth(options *AuthOptions) (git2go.RemoteCallbacks, error) {
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
			//TODO: ssh with passphrase
			key, ok := options.Credentials[KeyKey]
			if !ok {
				return git2go.RemoteCallbacks{}, fmt.Errorf("missing field %s", KeyKey)
			}

			_, err := ssh.ParsePrivateKey([]byte(key))
			if err != nil {
				return git2go.RemoteCallbacks{}, err
			}
			//TODO: check if username is needed
			cred, err := git2go.NewCredentialSSHKeyFromMemory("", "", key, "")
			if err != nil {
				return git2go.RemoteCallbacks{}, err
			}
			return git2go.RemoteCallbacks{
				CredentialsCallback:      authCallback(cred),
				CertificateCheckCallback: sshCheckCallback(),
			}, nil

		}
	default:
		return git2go.RemoteCallbacks{}, nil
	}

}

func authCallback(cred *git2go.Credential) func(url, username string, allowed_types git2go.CredentialType) (*git2go.Credential, error) {
	return func(url, username string, allowed_types git2go.CredentialType) (*git2go.Credential, error) {
		return cred, nil
	}
}

func sshCheckCallback() func(cert *git2go.Certificate, valid bool, hostname string) git2go.ErrorCode {
	return func(cert *git2go.Certificate, valid bool, hostname string) git2go.ErrorCode {
		return 0
	}
}

func removeDir(path string) {
	os.RemoveAll(path)
}
