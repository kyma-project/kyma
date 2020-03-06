package strategy

import (
	"kyma-project.io/compass-runtime-agent/internal/apperrors"
	"kyma-project.io/compass-runtime-agent/internal/kyma/apiresources/gateway-for-app/secrets/model"
	"kyma-project.io/compass-runtime-agent/internal/kyma/applications"
	"kyma-project.io/compass-runtime-agent/internal/kyma/applications/converters"
)

type SecretData map[string][]byte

//go:generate mockery -name=ModificationStrategy
type ModificationStrategy interface {
	CredentialsProvided(credentials *model.CredentialsWithCSRF) bool
	CreateSecretData(credentials *model.CredentialsWithCSRF) (SecretData, apperrors.AppError)
	ToCredentialsInfo(credentials *model.CredentialsWithCSRF, secretName string) applications.Credentials
	ShouldUpdate(currentData SecretData, newData SecretData) bool
}

//go:generate mockery -name=AccessStrategy
type AccessStrategy interface {
	ToCredentials(secretData SecretData, appCredentials *applications.Credentials) model.CredentialsWithCSRF
}

//go:generate mockery -name=Factory
type Factory interface {
	NewSecretModificationStrategy(credentials *model.CredentialsWithCSRF) (ModificationStrategy, apperrors.AppError)
	NewSecretAccessStrategy(credentials *applications.Credentials) (AccessStrategy, apperrors.AppError)
}

type factory struct {
}

func NewSecretsStrategyFactory() Factory {
	return &factory{}
}

func (s *factory) NewSecretModificationStrategy(credentials *model.CredentialsWithCSRF) (ModificationStrategy, apperrors.AppError) {
	if !credentialsValid(credentials) {
		return nil, apperrors.WrongInput("Error: only one credential type have to be provided.")
	}

	if credentials.Basic != nil {
		return &basicAuth{}, nil
	}

	if credentials.Oauth != nil {
		return &oauth{}, nil
	}

	return nil, apperrors.WrongInput("Invalid credential type provided")
}

func credentialsValid(credentials *model.CredentialsWithCSRF) bool {
	credentialsCount := 0

	if credentials.Basic != nil {
		credentialsCount++
	}

	if credentials.Oauth != nil {
		credentialsCount++
	}

	return credentialsCount == 1
}

func (s *factory) NewSecretAccessStrategy(credentials *applications.Credentials) (AccessStrategy, apperrors.AppError) {
	switch credentials.Type {
	case converters.CredentialsBasicType:
		return &basicAuth{}, nil
	case converters.CredentialsOAuthType:
		return &oauth{}, nil
	default:
		return nil, apperrors.Internal("Failed to initialize secret access strategy")
	}
}
