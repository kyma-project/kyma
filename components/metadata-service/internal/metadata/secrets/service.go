package secrets

import (
	"github.com/kyma-project/kyma/components/metadata-service/internal/apperrors"
)

const (
	OauthClientIDKey     = "clientId"
	OauthClientSecretKey = "clientSecret"

	BasicAuthUsernameKey = "username"
	BasicAuthPasswordKey = "password"
)

type Service interface {
	CreateOauthSecret(remoteEnvironment, name, clientID, clientSecret, serviceID string) apperrors.AppError
	GetOauthSecret(remoteEnvironment, name string) (clientId, clientSecret string, err apperrors.AppError)
	UpdateOauthSecret(remoteEnvironment, name, clientID, clientSecret, serviceID string) apperrors.AppError
	CreateBasicAuthSecret(remoteEnvironment, name, username, password, serviceID string) apperrors.AppError
	GetBasicAuthSecret(remoteEnvironment, name string) (username, password string, err apperrors.AppError)
	UpdateBasicAuthSecret(remoteEnvironment, name, username, password, serviceID string) apperrors.AppError
	DeleteSecret(name string) apperrors.AppError
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{
		repository: repository,
	}
}

func (s *service) CreateOauthSecret(remoteEnvironment, name, clientID, clientSecret, serviceID string) apperrors.AppError {
	err := verifySecretData(remoteEnvironment, name, serviceID)
	if err != nil {
		return err
	}

	data := makeOauthMap(clientID, clientSecret)
	return s.repository.Create(remoteEnvironment, name, serviceID, data)
}

func (s *service) GetOauthSecret(remoteEnvironment, name string) (string, string, apperrors.AppError) {
	data, err := s.repository.Get(remoteEnvironment, name)
	if err != nil {
		return "", "", err
	}

	clientId, clientSecret := readOauthMap(data)
	return clientId, clientSecret, nil
}

func (s *service) UpdateOauthSecret(remoteEnvironment, name, clientID, clientSecret, serviceID string) apperrors.AppError {
	err := verifySecretData(remoteEnvironment, name, serviceID)
	if err != nil {
		return err
	}

	data := makeOauthMap(clientID, clientSecret)
	return s.repository.Upsert(remoteEnvironment, name, serviceID, data)
}

func (s *service) CreateBasicAuthSecret(remoteEnvironment, name, username, password, serviceID string) apperrors.AppError {
	err := verifySecretData(remoteEnvironment, name, serviceID)
	if err != nil {
		return err
	}

	data := makeBasicAuthMap(username, password)
	return s.repository.Create(remoteEnvironment, name, serviceID, data)
}

func (s *service) GetBasicAuthSecret(remoteEnvironment, name string) (string, string, apperrors.AppError) {
	data, err := s.repository.Get(remoteEnvironment, name)
	if err != nil {
		return "", "", err
	}

	username, password := readBasicAuthMap(data)
	return username, password, nil
}

func (s *service) UpdateBasicAuthSecret(remoteEnvironment, name, username, password, serviceID string) apperrors.AppError {
	err := verifySecretData(remoteEnvironment, name, serviceID)
	if err != nil {
		return err
	}

	data := makeBasicAuthMap(username, password)
	return s.repository.Upsert(remoteEnvironment, name, serviceID, data)
}

func (s *service) DeleteSecret(name string) apperrors.AppError {
	return s.repository.Delete(name)
}

func verifySecretData(remoteEnvironment, name, serviceID string) apperrors.AppError {
	if remoteEnvironment == "" || name == "" || serviceID == "" {
		return apperrors.Internal("Incomplete secret data.")
	}

	return nil
}

func makeOauthMap(clientID, clientSecret string) map[string][]byte {
	return map[string][]byte{
		OauthClientIDKey:     []byte(clientID),
		OauthClientSecretKey: []byte(clientSecret),
	}
}

func readOauthMap(data map[string][]byte) (clientID, clientSecret string) {
	return string(data[OauthClientIDKey]), string(data[OauthClientSecretKey])
}

func makeBasicAuthMap(username, password string) map[string][]byte {
	return map[string][]byte{
		BasicAuthUsernameKey: []byte(username),
		BasicAuthPasswordKey: []byte(password),
	}
}

func readBasicAuthMap(data map[string][]byte) (username, password string) {
	return string(data[BasicAuthUsernameKey]), string(data[BasicAuthPasswordKey])
}
