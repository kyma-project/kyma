package secrets

import (
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/k8sconsts"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/applications"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/model"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/secrets/strategy"
)

type modificationFunction func(application, name, serviceID string, data map[string][]byte) apperrors.AppError

type Service interface {
	Get(application string, credentials applications.Credentials) (model.Credentials, apperrors.AppError)
	Create(application, serviceID string, credentials *model.Credentials) (applications.Credentials, apperrors.AppError)
	Update(application, serviceID string, credentials *model.Credentials) (applications.Credentials, apperrors.AppError)
	Delete(name string) apperrors.AppError
}

type service struct {
	nameResolver    k8sconsts.NameResolver
	repository      Repository
	strategyFactory strategy.Factory
}

func NewService(repository Repository, nameResolver k8sconsts.NameResolver, strategyFactory strategy.Factory) Service {
	return &service{
		nameResolver:    nameResolver,
		repository:      repository,
		strategyFactory: strategyFactory,
	}
}

func (s *service) Create(application, serviceID string, credentials *model.Credentials) (applications.Credentials, apperrors.AppError) {
	return s.modifySecret(application, serviceID, credentials, s.repository.Create)
}

func (s *service) Get(application string, credentials applications.Credentials) (model.Credentials, apperrors.AppError) {
	accessStrategy, err := s.strategyFactory.NewSecretAccessStrategy(&credentials)
	if err != nil {
		return model.Credentials{}, err.Append("Failed to get secret") // TODO - improve msg
	}

	data, err := s.repository.Get(application, credentials.SecretName)
	if err != nil {
		return model.Credentials{}, err
	}

	return accessStrategy.ToCredentials(data, &credentials), nil
}

func (s *service) Update(application, serviceID string, credentials *model.Credentials) (applications.Credentials, apperrors.AppError) {
	return s.modifySecret(application, serviceID, credentials, s.repository.Upsert)
}

func (s *service) Delete(name string) apperrors.AppError {
	return s.repository.Delete(name)
}

func (s *service) modifySecret(application, serviceID string, credentials *model.Credentials, modFunction modificationFunction) (applications.Credentials, apperrors.AppError) {
	if credentials == nil {
		return applications.Credentials{}, nil
	}

	modStrategy, err := s.strategyFactory.NewSecretModificationStrategy(credentials)
	if err != nil {
		return applications.Credentials{}, err.Append("Failed to initialize strategy")
	}

	if !modStrategy.CredentialsProvided(credentials) {
		return applications.Credentials{}, nil
	}

	name := s.nameResolver.GetResourceName(application, serviceID)

	secretData, err := modStrategy.CreateSecretData(credentials)
	if err != nil {
		return applications.Credentials{}, err.Append("Failed to create secret data")
	}

	err = modFunction(application, name, serviceID, secretData)
	if err != nil {
		return applications.Credentials{}, err
	}

	return modStrategy.ToAppCredentials(credentials, name), nil
}
