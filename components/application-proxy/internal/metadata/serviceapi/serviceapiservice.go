package serviceapi

import (
	"github.com/kyma-project/kyma/components/application-proxy/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-proxy/internal/metadata/applications"
	"github.com/kyma-project/kyma/components/application-proxy/internal/metadata/model"
	"github.com/kyma-project/kyma/components/application-proxy/internal/metadata/secrets"
)

const (
	ClientIDKey        = "clientId"
	ClientSecretKey    = "clientSecret"
	UsernameKey        = "username"
	PasswordKey        = "password"
	TypeOAuth          = "OAuth"
	TypeBasic          = "Basic"
	TypeCertificateGen = "CertificateGen"
	PrivateKeyKey      = "key"
	CertificateKey     = "crt"
	CommonNameKey      = "commonName"
)

// Service manages API definition of a service
type Service interface {
	// Read reads API from Application API definition. It also reads all additional information.
	Read(*applications.ServiceAPI) (*model.API, apperrors.AppError)
}

type defaultService struct {
	secretsRepository secrets.Repository
}

func NewService(secretsRepository secrets.Repository) Service {

	return defaultService{
		secretsRepository: secretsRepository,
	}
}

func (sas defaultService) Read(applicationAPI *applications.ServiceAPI) (*model.API, apperrors.AppError) {
	api := &model.API{
		TargetUrl: applicationAPI.TargetUrl,
	}

	if applicationAPI.Credentials != nil {
		credentialsSecretName := applicationAPI.Credentials.SecretName
		credentialsType := applicationAPI.Credentials.Type

		secret, err := sas.secretsRepository.Get(credentialsSecretName)

		if err != nil {
			return nil, err
		}

		if credentialsType == TypeOAuth {
			api.Credentials = &model.Credentials{
				OAuth: getOAuthCredentials(secret, applicationAPI.Credentials.Url),
			}
		} else if credentialsType == TypeBasic {
			api.Credentials = &model.Credentials{
				BasicAuth: getBasicAuthCredentials(secret),
			}
		} else if credentialsType == TypeCertificateGen {
			api.Credentials = &model.Credentials{
				CertificateGen: getCertificateGenCredentials(secret),
			}
		} else {
			api.Credentials = nil
		}
	}

	return api, nil
}

func getOAuthCredentials(secret map[string][]byte, url string) *model.OAuth {
	return &model.OAuth{
		ClientID:     string(secret[ClientIDKey]),
		ClientSecret: string(secret[ClientSecretKey]),
		URL:          url,
	}
}

func getBasicAuthCredentials(secret map[string][]byte) *model.BasicAuth {
	return &model.BasicAuth{
		Username: string(secret[UsernameKey]),
		Password: string(secret[PasswordKey]),
	}
}

func getCertificateGenCredentials(secret map[string][]byte) *model.CertificateGen {
	return &model.CertificateGen{
		CommonName:  string(secret[CommonNameKey]),
		Certificate: secret[CertificateKey],
		PrivateKey:  secret[PrivateKeyKey],
	}
}
