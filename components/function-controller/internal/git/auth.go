package git

import (
	"fmt"
	"strings"

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
	RepositoryAuthSSHKey RepositoryAuthType = "key"
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

			return git2go.RemoteCallbacks{
				CredentialsCallback: authBasicCallback(username, password),
			}, nil
		}
	case RepositoryAuthSSHKey:
		{
			key, ok := options.Credentials[KeyKey]
			if !ok {
				return git2go.RemoteCallbacks{}, fmt.Errorf("missing field %s", KeyKey)
			}
			passphrase := options.Credentials[PasswordKey]
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
				CredentialsCallback:      authSSHCallback(cred),
				CertificateCheckCallback: sshCheckCallback(),
			}, nil

		}
	}
	return git2go.RemoteCallbacks{}, errors.Errorf("unknown authentication type: %s", options.Type)
}

func authSSHCallback(cred *git2go.Credential) func(url, username string, allowed_types git2go.CredentialType) (*git2go.Credential, error) {
	return func(url string, username_from_url string, allowed_types git2go.CredentialType) (*git2go.Credential, error) {
		return cred, nil
	}
}

func authBasicCallback(username, password string) func(url, username string, allowed_types git2go.CredentialType) (*git2go.Credential, error) {
	return func(url string, username_from_url string, allowed_types git2go.CredentialType) (*git2go.Credential, error) {
		cred, err := git2go.NewCredentialUserpassPlaintext(username, password)
		if err != nil {
			return nil, errors.Wrap(err, "while creating credentials with user and password")
		}
		return cred, nil
	}
}

func sshCheckCallback() func(cert *git2go.Certificate, valid bool, hostname string) git2go.ErrorCode {
	return func(cert *git2go.Certificate, valid bool, hostname string) git2go.ErrorCode {
		return git2go.ErrOk
	}
}

func IsAuthErr(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()
	if strings.Contains(errMsg, "unexpected http status code: 403") {
		return true
	}

	/*
		When using invalid personal access token with basic auth, libgit2 return such error.
	*/

	if strings.Contains(errMsg, "too many redirects or authentication replays") {
		return true
	}
	return false
}
