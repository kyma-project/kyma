package secrets

import (
	"github.com/kyma-project/kyma/components/metadata-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/metadata-service/internal/k8sconsts"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/applications"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/model"
)

const (
	OauthClientIDKey     = "clientId"
	OauthClientSecretKey = "clientSecret"

	BasicAuthUsernameKey = "username"
	BasicAuthPasswordKey = "password"
)

type Service interface {
	Get(application string, credentials applications.Credentials) (model.Credentials, apperrors.AppError)
	Create(application, serviceID string, credentials *model.Credentials) (applications.Credentials, apperrors.AppError)
	Update(application, serviceID string, credentials *model.Credentials) (applications.Credentials, apperrors.AppError)
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

func (s *service) Create(application, serviceID string, credentials *model.Credentials) (applications.Credentials, apperrors.AppError) {
	if credentials == nil {
		return applications.Credentials{}, nil
	}

	if basicCredentialsProvided(credentials) && oauthCredentialsProvided(credentials) {
		return applications.Credentials{}, apperrors.WrongInput("Creating access service failed: Multiple authentication methods provided.")
	}

	name := s.nameResolver.GetResourceName(application, serviceID)

	err := s.createCredentialsSecret(application, serviceID, name, credentials)
	if err != nil {
		return applications.Credentials{}, err
	}

	applicationCredentials := s.modelToApplicationCredentials(credentials, application, serviceID)

	return applicationCredentials, nil
}

func (s *service) Get(application string, credentials applications.Credentials) (model.Credentials, apperrors.AppError) {
	if credentials.Type == applications.CredentialsOAuthType {
		data, err := s.repository.Get(application, credentials.SecretName)
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

	if credentials.Type == applications.CredentialsBasicType {
		data, err := s.repository.Get(application, credentials.SecretName)
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

func (s *service) Update(application, serviceID string, credentials *model.Credentials) (applications.Credentials, apperrors.AppError) {
	if credentials == nil {
		return applications.Credentials{}, nil
	}

	if basicCredentialsProvided(credentials) && oauthCredentialsProvided(credentials) {
		return applications.Credentials{}, apperrors.WrongInput("Creating access service failed: Multiple authentication methods provided.")
	}

	name := s.nameResolver.GetResourceName(application, serviceID)

	err := s.updateCredentialsSecret(application, serviceID, name, credentials)
	if err != nil {
		return applications.Credentials{}, err
	}

	applicationCredentials := s.modelToApplicationCredentials(credentials, application, serviceID)

	return applicationCredentials, nil
}

func (s *service) Delete(name string) apperrors.AppError {
	return s.repository.Delete(name)
}

func (s *service) createCredentialsSecret(application, id, name string, credentials *model.Credentials) apperrors.AppError {
	if oauthCredentialsProvided(credentials) {
		return s.createOauthSecret(
			application,
			name,
			credentials.Oauth.ClientID,
			credentials.Oauth.ClientSecret,
			id,
		)
	}

	if basicCredentialsProvided(credentials) {
		return s.createBasicAuthSecret(application,
			name,
			credentials.Basic.Username,
			credentials.Basic.Password,
			id,
		)
	}

	return nil
}

func (s *service) updateCredentialsSecret(application, id, name string, credentials *model.Credentials) apperrors.AppError {
	if oauthCredentialsProvided(credentials) {
		return s.updateOauthSecret(
			application,
			name,
			credentials.Oauth.ClientID,
			credentials.Oauth.ClientSecret,
			id,
		)
	}

	if basicCredentialsProvided(credentials) {
		return s.updateBasicAuthSecret(application,
			name,
			credentials.Basic.Username,
			credentials.Basic.Password,
			id,
		)
	}

	return nil
}

func (s *service) createOauthSecret(application, name, clientID, clientSecret, serviceID string) apperrors.AppError {
	err := verifySecretData(application, name, serviceID)
	if err != nil {
		return err
	}

	data := makeOauthMap(clientID, clientSecret)
	return s.repository.Create(application, name, serviceID, data)
}

func (s *service) getOauthSecret(application, name string) (string, string, apperrors.AppError) {
	data, err := s.repository.Get(application, name)
	if err != nil {
		return "", "", err
	}

	clientId, clientSecret := readOauthMap(data)
	return clientId, clientSecret, nil
}

func (s *service) updateOauthSecret(application, name, clientID, clientSecret, serviceID string) apperrors.AppError {
	err := verifySecretData(application, name, serviceID)
	if err != nil {
		return err
	}

	data := makeOauthMap(clientID, clientSecret)
	return s.repository.Upsert(application, name, serviceID, data)
}

func (s *service) createBasicAuthSecret(application, name, username, password, serviceID string) apperrors.AppError {
	err := verifySecretData(application, name, serviceID)
	if err != nil {
		return err
	}

	data := makeBasicAuthMap(username, password)
	return s.repository.Create(application, name, serviceID, data)
}

func (s *service) getBasicAuthSecret(application, name string) (string, string, apperrors.AppError) {
	data, err := s.repository.Get(application, name)
	if err != nil {
		return "", "", err
	}

	username, password := readBasicAuthMap(data)
	return username, password, nil
}

func (s *service) updateBasicAuthSecret(application, name, username, password, serviceID string) apperrors.AppError {
	err := verifySecretData(application, name, serviceID)
	if err != nil {
		return err
	}

	data := makeBasicAuthMap(username, password)
	return s.repository.Upsert(application, name, serviceID, data)
}

func (s *service) modelToApplicationCredentials(credentials *model.Credentials, application, serviceID string) applications.Credentials {
	applicationCredentials := applications.Credentials{}

	if oauthCredentialsProvided(credentials) {
		applicationCredentials.AuthenticationUrl = credentials.Oauth.URL
		applicationCredentials.Type = applications.CredentialsOAuthType
	}

	if basicCredentialsProvided(credentials) {
		applicationCredentials.Type = applications.CredentialsBasicType
	}

	applicationCredentials.SecretName = s.nameResolver.GetResourceName(application, serviceID)

	return applicationCredentials
}

func oauthCredentialsProvided(credentials *model.Credentials) bool {
	return credentials != nil && credentials.Oauth != nil && credentials.Oauth.ClientID != "" && credentials.Oauth.ClientSecret != ""
}

func basicCredentialsProvided(credentials *model.Credentials) bool {
	return credentials != nil && credentials.Basic != nil && credentials.Basic.Username != "" && credentials.Basic.Password != ""
}

func verifySecretData(application, name, serviceID string) apperrors.AppError {
	if application == "" || name == "" || serviceID == "" {
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
