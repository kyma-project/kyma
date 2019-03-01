package strategy

import (
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/applications"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/certificates"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/model"
)

type SecretData map[string][]byte

type ModificationStrategy interface {
	CredentialsProvided(credentials *model.Credentials) bool
	CreateSecretData(credentials *model.Credentials) (SecretData, apperrors.AppError)
	ToCredentialsInfo(credentials *model.Credentials, secretName string) applications.Credentials
	ShouldUpdate(currentData SecretData, newData SecretData) bool
}

type AccessStrategy interface {
	ToCredentials(secretData SecretData, appCredentials *applications.Credentials) model.Credentials
}

type Factory interface {
	NewSecretModificationStrategy(credentials *model.Credentials) (ModificationStrategy, apperrors.AppError)
	NewSecretAccessStrategy(credentials *applications.Credentials) (AccessStrategy, apperrors.AppError)
}

type factory struct {
	certificateGenerator certificates.Generator
}

func NewSecretsStrategyFactory(certificateGenerator certificates.Generator) Factory {
	return &factory{
		certificateGenerator: certificateGenerator,
	}
}

func (s *factory) NewSecretModificationStrategy(credentials *model.Credentials) (ModificationStrategy, apperrors.AppError) {
	if !credentialsValid(credentials) {
		return nil, apperrors.WrongInput("Error: only one credential type have to be provided.")
	}

	if credentials.Basic != nil {
		return &basicAuth{}, nil
	}

	if credentials.Oauth != nil {
		return &oauth{}, nil
	}

	if credentials.CertificateGen != nil {
		return &certificateGen{
			certificateGenerator: s.certificateGenerator,
		}, nil
	}

	if credentials.CSRFToken !=nil {
		return &csrfToken{}, nil
	}

	return nil, apperrors.WrongInput("Invalid credential type provided")
}

func credentialsValid(credentials *model.Credentials) bool {
	credentialsCount := 0

	if credentials.Basic != nil {
		credentialsCount++
	}

	if credentials.Oauth != nil {
		credentialsCount++
	}

	if credentials.CertificateGen != nil {
		credentialsCount++
	}

	if credentials.CSRFToken != nil {
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
	case applications.CredentialsCertificateGenType:
		return &certificateGen{
			certificateGenerator: s.certificateGenerator,
		}, nil
	case applications.CredentialsCSRFTokenType:
		return &csrfToken{}, nil
	default:
		return nil, apperrors.Internal("Failed to initialize secret access strategy")
	}
}
