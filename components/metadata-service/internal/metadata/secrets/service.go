package secrets

import (
	"github.com/kyma-project/kyma/components/metadata-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/metadata-service/internal/k8sconsts"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/model"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/remoteenv"
)

const (
	OauthClientIDKey     = "clientId"
	OauthClientSecretKey = "clientSecret"

	BasicAuthUsernameKey = "username"
	BasicAuthPasswordKey = "password"
)

type Service interface {
	Get(remoteEnvironment string, credentials remoteenv.Credentials) (model.Credentials, apperrors.AppError)
	Create(remoteEnvironment, serviceID string, credentials *model.Credentials) (remoteenv.Credentials, apperrors.AppError)
	Update(remoteEnvironment, serviceID string, credentials *model.Credentials) (remoteenv.Credentials, apperrors.AppError)
	Delete(name string) apperrors.AppError
}

type service struct {
	nameResolver k8sconsts.NameResolver
	repository   Repository
}

func NewService(repository Repository, nameResolver k8sconsts.NameResolver) Service {
	return &service{
		nameResolver: nameResolver,
		repository:   repository,
	}
}

func (s *service) Create(remoteEnvironment, serviceID string, credentials *model.Credentials) (remoteenv.Credentials, apperrors.AppError) {
	if credentials == nil {
		return remoteenv.Credentials{}, nil
	}

	if basicCredentialsProvided(credentials) && oauthCredentialsProvided(credentials) {
		return remoteenv.Credentials{}, apperrors.WrongInput("Creating access service failed: Multiple authentication methods provided.")
	}

	name := s.nameResolver.GetResourceName(remoteEnvironment, serviceID)

	err := s.createCredentialsSecret(remoteEnvironment, serviceID, name, credentials)
	if err != nil {
		return remoteenv.Credentials{}, err
	}

	remoteEnvCredentials := s.modelToRemoteEnvCredentials(credentials, remoteEnvironment, serviceID)

	return remoteEnvCredentials, nil
}

func (s *service) Get(remoteEnvironment string, credentials remoteenv.Credentials) (model.Credentials, apperrors.AppError) {
	if credentials.Type == remoteenv.CredentialsOAuthType {
		data, err := s.repository.Get(remoteEnvironment, credentials.SecretName)
		if err != nil {
			return model.Credentials{}, err
		}

		clientId, clientSecret := readOauthMap(data)

		return model.Credentials{
			Oauth: &model.Oauth{
				ClientID:     clientId,
				ClientSecret: clientSecret,
				URL:          credentials.AuthenticationUrl,
			},
		}, nil
	}

	if credentials.Type == remoteenv.CredentialsBasicType {
		data, err := s.repository.Get(remoteEnvironment, credentials.SecretName)
		if err != nil {
			return model.Credentials{}, err
		}

		username, password := readBasicAuthMap(data)

		return model.Credentials{
			Basic: &model.Basic{
				Username: username,
				Password: password,
			},
		}, nil
	}

	return model.Credentials{}, nil
}

func (s *service) Update(remoteEnvironment, serviceID string, credentials *model.Credentials) (remoteenv.Credentials, apperrors.AppError) {
	if credentials == nil {
		return remoteenv.Credentials{}, nil
	}

	if basicCredentialsProvided(credentials) && oauthCredentialsProvided(credentials) {
		return remoteenv.Credentials{}, apperrors.WrongInput("Creating access service failed: Multiple authentication methods provided.")
	}

	name := s.nameResolver.GetResourceName(remoteEnvironment, serviceID)

	err := s.updateCredentialsSecret(remoteEnvironment, serviceID, name, credentials)
	if err != nil {
		return remoteenv.Credentials{}, err
	}

	remoteEnvCredentials := s.modelToRemoteEnvCredentials(credentials, remoteEnvironment, serviceID)

	return remoteEnvCredentials, nil
}

func (s *service) Delete(name string) apperrors.AppError {
	return s.repository.Delete(name)
}

func (s *service) createCredentialsSecret(remoteEnvironment, id, name string, credentials *model.Credentials) apperrors.AppError {
	if oauthCredentialsProvided(credentials) {
		return s.createOauthSecret(
			remoteEnvironment,
			name,
			credentials.Oauth.ClientID,
			credentials.Oauth.ClientSecret,
			id,
		)
	}

	if basicCredentialsProvided(credentials) {
		return s.createBasicAuthSecret(remoteEnvironment,
			name,
			credentials.Basic.Username,
			credentials.Basic.Password,
			id,
		)
	}

	return nil
}

func (s *service) updateCredentialsSecret(remoteEnvironment, id, name string, credentials *model.Credentials) apperrors.AppError {
	if oauthCredentialsProvided(credentials) {
		return s.updateOauthSecret(
			remoteEnvironment,
			name,
			credentials.Oauth.ClientID,
			credentials.Oauth.ClientSecret,
			id,
		)
	}

	if basicCredentialsProvided(credentials) {
		return s.updateBasicAuthSecret(remoteEnvironment,
			name,
			credentials.Basic.Username,
			credentials.Basic.Password,
			id,
		)
	}

	return nil
}

func (s *service) createOauthSecret(remoteEnvironment, name, clientID, clientSecret, serviceID string) apperrors.AppError {
	err := verifySecretData(remoteEnvironment, name, serviceID)
	if err != nil {
		return err
	}

	data := makeOauthMap(clientID, clientSecret)
	return s.repository.Create(remoteEnvironment, name, serviceID, data)
}

func (s *service) getOauthSecret(remoteEnvironment, name string) (string, string, apperrors.AppError) {
	data, err := s.repository.Get(remoteEnvironment, name)
	if err != nil {
		return "", "", err
	}

	clientId, clientSecret := readOauthMap(data)
	return clientId, clientSecret, nil
}

func (s *service) updateOauthSecret(remoteEnvironment, name, clientID, clientSecret, serviceID string) apperrors.AppError {
	err := verifySecretData(remoteEnvironment, name, serviceID)
	if err != nil {
		return err
	}

	data := makeOauthMap(clientID, clientSecret)
	return s.repository.Upsert(remoteEnvironment, name, serviceID, data)
}

func (s *service) createBasicAuthSecret(remoteEnvironment, name, username, password, serviceID string) apperrors.AppError {
	err := verifySecretData(remoteEnvironment, name, serviceID)
	if err != nil {
		return err
	}

	data := makeBasicAuthMap(username, password)
	return s.repository.Create(remoteEnvironment, name, serviceID, data)
}

func (s *service) getBasicAuthSecret(remoteEnvironment, name string) (string, string, apperrors.AppError) {
	data, err := s.repository.Get(remoteEnvironment, name)
	if err != nil {
		return "", "", err
	}

	username, password := readBasicAuthMap(data)
	return username, password, nil
}

func (s *service) updateBasicAuthSecret(remoteEnvironment, name, username, password, serviceID string) apperrors.AppError {
	err := verifySecretData(remoteEnvironment, name, serviceID)
	if err != nil {
		return err
	}

	data := makeBasicAuthMap(username, password)
	return s.repository.Upsert(remoteEnvironment, name, serviceID, data)
}

func (s *service) modelToRemoteEnvCredentials(credentials *model.Credentials, remoteEnvironment, serviceID string) remoteenv.Credentials {
	remoteEnvCredentials := remoteenv.Credentials{}

	if oauthCredentialsProvided(credentials) {
		remoteEnvCredentials.AuthenticationUrl = credentials.Oauth.URL
		remoteEnvCredentials.Type = remoteenv.CredentialsOAuthType
	}

	if basicCredentialsProvided(credentials) {
		remoteEnvCredentials.Type = remoteenv.CredentialsBasicType
	}

	remoteEnvCredentials.SecretName = s.nameResolver.GetResourceName(remoteEnvironment, serviceID)

	return remoteEnvCredentials
}

func oauthCredentialsProvided(credentials *model.Credentials) bool {
	return credentials != nil && credentials.Oauth != nil && credentials.Oauth.ClientID != "" && credentials.Oauth.ClientSecret != ""
}

func basicCredentialsProvided(credentials *model.Credentials) bool {
	return credentials != nil && credentials.Basic != nil && credentials.Basic.Username != "" && credentials.Basic.Password != ""
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
