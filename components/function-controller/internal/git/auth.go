package git

import (
	"fmt"

	git2go "github.com/libgit2/git2go/v31"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

const (
	UsernameKey = "username"
	PasswordKey = "password"
	KeyKey      = "key"
)

type RepositoryAuthType string

const (
	RepositoryAuthBasic  RepositoryAuthType = "basic"
	RepositoryAuthSSHKey                    = "key"
)

type AuthOptions struct {
	Type        RepositoryAuthType
	Credentials map[string]string
	SecretName  string
}

func GetAuth(options *AuthOptions) (git2go.RemoteCallbacks, error) {
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
			passphrase, _ := options.Credentials[PasswordKey]
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
	return git2go.RemoteCallbacks{}, errors.Errorf("unknown authentication type: %s", options.Type)
}
