package appsecrets

import (
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/k8sconsts"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/applications"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/model"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/secrets/strategy"
	"k8s.io/apimachinery/pkg/types"
)

type modificationFunction func(modStrategy strategy.ModificationStrategy, application string, appUID types.UID, name, packageID string, newData strategy.SecretData) apperrors.AppError

//go:generate mockery --name CredentialsService
type CredentialsService interface {
	Get(application string, credentials applications.Credentials) (model.Credentials, apperrors.AppError)
	Create(application string, appUID types.UID, packageID string, credentials *model.Credentials) (applications.Credentials, apperrors.AppError)
	Upsert(application string, appUID types.UID, packageID string, credentials *model.Credentials) (applications.Credentials, apperrors.AppError)
	Delete(name string) apperrors.AppError
}

type credentialsService struct {
	repository      Repository
	strategyFactory strategy.Factory
	nameResolver    k8sconsts.NameResolver
}

func NewCredentialsService(repository Repository, strategyFactory strategy.Factory, nameResolver k8sconsts.NameResolver) CredentialsService {
	return &credentialsService{
		repository:      repository,
		strategyFactory: strategyFactory,
		nameResolver:    nameResolver,
	}
}

func (s *credentialsService) Create(application string, appUID types.UID, packageID string, credentials *model.Credentials) (applications.Credentials, apperrors.AppError) {
	return s.modifySecret(application, appUID, packageID, credentials, s.createSecret)
}

func (s *credentialsService) Get(application string, credentials applications.Credentials) (model.Credentials, apperrors.AppError) {
	accessStrategy, err := s.strategyFactory.NewSecretAccessStrategy(&credentials)
	if err != nil {
		return model.Credentials{}, err.Append("Failed to initialize strategy")
	}

	data, err := s.repository.Get(credentials.SecretName)
	if err != nil {
		return model.Credentials{}, err
	}

	return accessStrategy.ToCredentials(data, &credentials)
}

func (s *credentialsService) Upsert(application string, appUID types.UID, packageID string, credentials *model.Credentials) (applications.Credentials, apperrors.AppError) {
	return s.modifySecret(application, appUID, packageID, credentials, s.upsertSecret)
}

func (s *credentialsService) Delete(name string) apperrors.AppError {
	return s.repository.Delete(name)
}

func (s *credentialsService) modifySecret(application string, appUID types.UID, packageID string, credentials *model.Credentials, modFunction modificationFunction) (applications.Credentials, apperrors.AppError) {
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

	name := s.nameResolver.GetCredentialsSecretName(application, packageID)

	secretData, err := modStrategy.CreateSecretData(credentials)
	if err != nil {
		return applications.Credentials{}, err.Append("Failed to create secret data")
	}

	err = modFunction(modStrategy, application, appUID, name, packageID, secretData)
	if err != nil {
		return applications.Credentials{}, err
	}

	return modStrategy.ToCredentialsInfo(credentials, name), nil
}

func (s *credentialsService) upsertSecret(modStrategy strategy.ModificationStrategy, application string, appUID types.UID, name, packageID string, newData strategy.SecretData) apperrors.AppError {
	currentData, err := s.repository.Get(name)
	if err != nil {
		if err.Code() == apperrors.CodeNotFound {
			return s.repository.Create(application, appUID, name, packageID, newData)
		}

		return err
	}

	if modStrategy.ShouldUpdate(currentData, newData) {
		return s.repository.Upsert(application, appUID, name, packageID, newData)
	}

	return nil
}

func (s *credentialsService) createSecret(_ strategy.ModificationStrategy, application string, appUID types.UID, name, packageID string, newData strategy.SecretData) apperrors.AppError {
	return s.repository.Create(application, appUID, name, packageID, newData)
}
