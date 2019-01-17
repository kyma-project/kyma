package strategy

import (
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/applications"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/certificates"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/model"
)

type Strategy interface {
	ToCredentials(secretData map[string][]byte, appCredentials *applications.Credentials) model.Credentials
	CredentialsProvided(credentials *model.Credentials) bool
	CreateSecretData(credentials *model.Credentials) (map[string][]byte, apperrors.AppError) // TODO - allias secretData?
	ToAppCredentials(credentials *model.Credentials, secretName string) applications.Credentials
}

// TODO - split Strategy interface to two?
type Factory interface {
	NewSecretModificationStrategy(credentials *model.Credentials) (Strategy, apperrors.AppError)
	NewSecretAccessStrategy(credentials *applications.Credentials) (Strategy, apperrors.AppError)
}

type factory struct {
	certificateGenerator certificates.Generator
}

func NewSecretsStrategyFactory(certificateGenerator certificates.Generator) Factory {
	return &factory{
		certificateGenerator: certificateGenerator,
	}
}

func (s *factory) NewSecretModificationStrategy(credentials *model.Credentials) (Strategy, apperrors.AppError) {
	// TODO - decide what to do with all that validation?
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

	return credentialsCount == 1
}

func (s *factory) NewSecretAccessStrategy(credentials *applications.Credentials) (Strategy, apperrors.AppError) {
	switch credentials.Type {
	case applications.CredentialsBasicType:
		return &basicAuth{}, nil
	case applications.CredentialsOAuthType:
		return &oauth{}, nil
	case applications.CredentialsCertificateGenType:
		return &certificateGen{
			certificateGenerator: s.certificateGenerator,
		}, nil
	default:
		return nil, apperrors.Internal("Failed to initialize secret access strategy")
	}
}
