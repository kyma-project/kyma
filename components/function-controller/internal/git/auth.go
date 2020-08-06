package git

import (
	"fmt"
	"net"

	"golang.org/x/crypto/ssh"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/pkg/errors"
)

const (
	UsernameKey = "username"
	PasswordKey = "password"
	KeyKey      = "key"
)

type AuthOptions struct {
	Type        RepositoryAuthType
	Credentials map[string]string
	SecretName  string
}

func (a *AuthOptions) ToAuthMethod() (transport.AuthMethod, error) {
	if a == nil {
		return nil, nil
	}

	switch a.Type {
	case RepositoryAuthBasic:
		basic, err := a.toBasicAuth()
		if err != nil {
			return nil, errors.Wrapf(err, "while converting authentication config to %s auth method", a.Type)
		}
		return transport.AuthMethod(basic), nil
	case RepositoryAuthSSHKey:
		key, err := a.toKeyAuth()
		if err != nil {
			return nil, errors.Wrapf(err, "while converting authentication config to %s auth method", a.Type)
		}
		return transport.AuthMethod(key), nil
	default:
		return nil, fmt.Errorf("unknown authentication type: %s", a.Type)
	}
}

func (a *AuthOptions) toBasicAuth() (*http.BasicAuth, error) {
	if a.Credentials == nil {
		return &http.BasicAuth{}, nil
	}

	username, ok := a.Credentials[UsernameKey]
	if !ok {
		return nil, fmt.Errorf("missing field %s", UsernameKey)
	}

	password, ok := a.Credentials[PasswordKey]
	if !ok {
		return nil, fmt.Errorf("missing field %s", PasswordKey)
	}

	return &http.BasicAuth{
		Username: username,
		Password: password,
	}, nil
}

func (a *AuthOptions) toKeyAuth() (*gitssh.PublicKeys, error) {
	if a.Credentials == nil {
		return &gitssh.PublicKeys{}, nil
	}

	key, ok := a.Credentials[KeyKey]
	if !ok {
		return nil, fmt.Errorf("missing field %s", KeyKey)
	}

	password, _ := a.Credentials[PasswordKey]

	var signer ssh.Signer
	var err error
	if password == "" {
		signer, err = ssh.ParsePrivateKey([]byte(key))
	} else {
		signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(key), []byte(password))
	}
	if err != nil {
		return nil, errors.Wrapf(err, "while creating public keys authentication method")
	}

	auth := gitssh.PublicKeys{
		User:   "git",
		Signer: signer,
		HostKeyCallbackHelper: gitssh.HostKeyCallbackHelper{HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		}},
	}

	return &auth, nil
}
