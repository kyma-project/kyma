package git

import (
	"fmt"
	git2go "github.com/libgit2/git2go/v31"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
)

const (
	tagRef    = "refs/tags/%s"
	branchRef = "refs/remotes/%s"
)

type Go2GitClient struct {
}

func NewGit2Go() *Go2GitClient {
	return &Go2GitClient{}
}

func (g *Go2GitClient) LastCommit(options Options) (string, error) {
	//TODO: Looks like reference can be commit hash, tag or branch
	_, err := git2go.NewOid(options.Reference)
	if err == nil {
		return options.Reference, nil
	}

	authCallbacks, err := getAuth(options.Auth)
	if err != nil {
		return "", err
	}

	tmpPath, err := ioutil.TempDir("/tmp", "last-commit")
	if err != nil {
		return "", err
	}
	log.Println(tmpPath)
	repo, err := git2go.Clone(options.URL, tmpPath, &git2go.CloneOptions{
		FetchOptions: &git2go.FetchOptions{
			RemoteCallbacks: *authCallbacks,
			DownloadTags:    git2go.DownloadTagsAll,
		},
	})
	if err != nil {
		return "", err
	}

	iter, err := repo.NewReferenceIterator()
	if err != nil {
		return "", err
	}

	log.Println("all references")
	for ; ; {
		item, err := iter.Next()
		if err != nil {
			break
		}

		log.Println(item.Name())
	}

	//Tags
	log.Println("Tag")

	tags, err := repo.Tags.List()
	if err != nil {
		return "", err
	}
	for tag, _ := range tags {
		log.Println(tag)
	}

	ref, err := repo.References.Dwim(fmt.Sprintf("refs/tags/first"))
	if err != nil {
		return "", err
	}

	log.Println(ref.Target().String())

	// NewBranchIterator
	branchIter, err := repo.NewBranchIterator(git2go.BranchAll)
	if err != nil {
		log.Panic(err)
	}
	log.Println("Branch")

	err = branchIter.ForEach(func(b *git2go.Branch, bt git2go.BranchType) error {
		log.Println(b.Reference.Name())
		return nil
	})
	if err != nil {
		return "", err
	}

	branchName := fmt.Sprintf("origins/%s", options.Reference)
	branch, err := repo.LookupBranch(branchName, git2go.BranchAll)
	if err != nil {
		//TODO: check for not found
		git2go.IsErrorCode(err, git2go.ErrNotFound)
		return "", err
	}

	return branch.Target().String(), nil
}

func (g *Go2GitClient) Clone(path string, options Options) (string, error) {
	authCallbacks, err := getAuth(options.Auth)
	if err != nil {
		return "", err
	}

	repo, err := git2go.Clone(options.URL, path, &git2go.CloneOptions{
		FetchOptions: &git2go.FetchOptions{
			RemoteCallbacks: *authCallbacks,
		},
	})
	if err != nil {
		return "", err
	}

	oid, err := git2go.NewOid(options.Reference)
	if err != nil {
		return "", err
	}

	commit, err := repo.LookupCommit(oid)
	if err != nil {
		return "", err
	}

	err = repo.ResetToCommit(commit, git2go.ResetHard, &git2go.CheckoutOptions{})
	if err != nil {
		return "", err
	}

	ref, err := repo.Head()
	if err != nil {
		return "", err
	}

	return ref.Target().String(), nil
}

func getAuth(options *AuthOptions) (*git2go.RemoteCallbacks, error) {
	switch authType := options.Type; authType {
	case RepositoryAuthBasic:
		{
			username, ok := options.Credentials[UsernameKey]
			if !ok {
				return nil, fmt.Errorf("missing field %s", UsernameKey)

			}
			password, ok := options.Credentials[PasswordKey]
			if !ok {
				return nil, fmt.Errorf("missing field %s", PasswordKey)
			}

			cred, err := git2go.NewCredentialUserpassPlaintext(username, password)
			if err != nil {
				return nil, errors.Wrap(err, "while creating basic auth")
			}
			return &git2go.RemoteCallbacks{
				CredentialsCallback: authCallback(cred),
			}, nil
		}
	case RepositoryAuthSSHKey:
		{
			//TODO: ssh with passphrase
			key, ok := options.Credentials[KeyKey]
			if !ok {
				return nil, fmt.Errorf("missing field %s", KeyKey)
			}

			_, err := ssh.ParsePrivateKey([]byte(key))
			if err != nil {
				return nil, err
			}
			//TODO: check if username is needed
			cred, err := git2go.NewCredentialSSHKeyFromMemory("", "", key, "")
			if err != nil {
				return nil, err
			}
			return &git2go.RemoteCallbacks{
				CredentialsCallback:      authCallback(cred),
				CertificateCheckCallback: sshCheckCallback(),
			}, nil

		}
	default:
		return nil, errors.Errorf("Unknow auth type: %s", authType)
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
