package secrets

import (
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/k8sconsts"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/applications"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/model"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/secrets/strategy"
)

type modificationFunction func(modStrategy strategy.ModificationStrategy, application, name, serviceID string, newData strategy.SecretData) apperrors.AppError

//TODO: Rename to CredentialsService add new service for the new secret
type Service interface {
	Get(application string, credentials applications.Credentials) (model.Credentials, apperrors.AppError)
	Create(application, serviceID string, credentials *model.Credentials) (applications.Credentials, apperrors.AppError)
	Upsert(application, serviceID string, credentials *model.Credentials) (applications.Credentials, apperrors.AppError)
	Delete(name string) apperrors.AppError
}

//TODO: Rename to credentialsService
type service struct {
	nameResolver    k8sconsts.NameResolver
	repository      Repository
	strategyFactory strategy.Factory
}

//TODO: Rename to NewCredentialsService
func NewService(repository Repository, nameResolver k8sconsts.NameResolver, strategyFactory strategy.Factory) Service {
	return &service{
		nameResolver:    nameResolver,
		repository:      repository,
		strategyFactory: strategyFactory,
	}
}

func (s *service) Create(application, serviceID string, credentials *model.Credentials) (applications.Credentials, apperrors.AppError) {
	return s.modifySecret(application, serviceID, credentials, s.createSecret)
}

func (s *service) Get(application string, credentials applications.Credentials) (model.Credentials, apperrors.AppError) {
	accessStrategy, err := s.strategyFactory.NewSecretAccessStrategy(&credentials)
	if err != nil {
		return model.Credentials{}, err.Append("Failed to initialize strategy")
	}

	data, err := s.repository.Get(application, credentials.SecretName)
	if err != nil {
		return model.Credentials{}, err
	}

	return accessStrategy.ToCredentials(data, &credentials), nil
}

func (s *service) Upsert(application, serviceID string, credentials *model.Credentials) (applications.Credentials, apperrors.AppError) {
	return s.modifySecret(application, serviceID, credentials, s.upsertSecret)
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

	err = modFunction(modStrategy, application, name, serviceID, secretData)
	if err != nil {
		return applications.Credentials{}, err
	}

	return modStrategy.ToCredentialsInfo(credentials, name), nil
}

func (s *service) upsertSecret(modStrategy strategy.ModificationStrategy, application, name, serviceID string, newData strategy.SecretData) apperrors.AppError {
	currentData, err := s.repository.Get(application, name)
	if err != nil {
		if err.Code() == apperrors.CodeNotFound {
			return s.repository.Create(application, name, serviceID, newData)
		}

		return err
	}

	if modStrategy.ShouldUpdate(currentData, newData) {
		return s.repository.Upsert(application, name, serviceID, newData)
	}

	return nil
}

func (s *service) createSecret(_ strategy.ModificationStrategy, application, name, serviceID string, newData strategy.SecretData) apperrors.AppError {
	return s.repository.Create(application, name, serviceID, newData)
}
