package strategy

import (
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/synchronization/apiresources/secrets/model"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/synchronization/applications"
)

type SecretData map[string][]byte

type ModificationStrategy interface {
	CredentialsProvided(credentials *model.CredentialsWithCSRF) bool
	CreateSecretData(credentials *model.CredentialsWithCSRF) (SecretData, apperrors.AppError)
	ToCredentialsInfo(credentials *model.CredentialsWithCSRF, secretName string) applications.Credentials
	ShouldUpdate(currentData SecretData, newData SecretData) bool
}

type AccessStrategy interface {
	ToCredentials(secretData SecretData, appCredentials *applications.Credentials) model.CredentialsWithCSRF
}

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
	case applications.CredentialsBasicType:
		return &basicAuth{}, nil
	case applications.CredentialsOAuthType:
		return &oauth{}, nil
	default:
		return nil, apperrors.Internal("Failed to initialize secret access strategy")
	}
}
