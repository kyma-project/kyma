package docker

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
)

const (
	keyUsername        = "username"
	keyPassword        = "password"
	keyRegistryAddress = "registryAddress"
)

var (
	errUsernameNotFound = errors.New("username field not found")
	errPasswordNotFound = errors.New("password field not found")
	errRegistryNotFound = errors.New("registry field not found")
)

type registryCfgCredentials struct {
	username, password, serverAddress []byte
	provideEncoder
}

type provideEncoder func(enc *base64.Encoding, w io.Writer) io.WriteCloser

func (e provideEncoder) encodeUserAndPassword(username, password []byte) (string, error) {
	var buff bytes.Buffer
	base64encoder := e(base64.StdEncoding, &buff)
	// encode username:password to buffer
	for _, bytes := range [3][]byte{
		username,
		[]byte(":"),
		password,
	} {
		if _, err := base64encoder.Write(bytes); err != nil {
			return "", err
		}
	}
	// close to flush
	if err := base64encoder.Close(); err != nil {
		return "", err
	}
	return buff.String(), nil
}

func createAuthMap(usernameAndPassword string) map[string]interface{} {
	return map[string]interface{}{
		"auth": usernameAndPassword,
	}
}

func (r *registryCfgCredentials) MarshalJSON() ([]byte, error) {
	userAndPassword, err := r.provideEncoder.encodeUserAndPassword(r.username, r.password)
	if err != nil {
		return nil, err
	}
	config := map[string]interface{}{
		"auths": map[string]interface{}{
			string(r.serverAddress): createAuthMap(userAndPassword),
		},
	}

	return json.Marshal(&config)
}

// NewRegistryCfgMarshaler creates registry configuration marshaler
func NewRegistryCfgMarshaler(data map[string][]byte) (json.Marshaler, error) {
	result := registryCfgCredentials{
		provideEncoder: base64.NewEncoder,
	}
	var found bool

	result.username, found = data[keyUsername]
	if !found {
		return nil, errUsernameNotFound
	}

	result.password, found = data[keyPassword]
	if !found {
		return nil, errPasswordNotFound
	}

	result.serverAddress, found = data[keyRegistryAddress]
	if !found {
		return nil, errRegistryNotFound
	}

	return &result, nil
}
