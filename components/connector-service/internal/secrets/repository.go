package secrets

import (
	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Manager interface {
	Get(name string, options metav1.GetOptions) (*v1.Secret, error)
}

type Repository interface {
	Get(name string) (crt []byte, key []byte, appError apperrors.AppError)
}

type repository struct {
	secretsManager Manager
}

func NewRepository(secretsManager Manager) Repository {
	return &repository{secretsManager: secretsManager}
}

func (r *repository) Get(name string) (crt []byte, key []byte, appError apperrors.AppError) {
	secret, err := r.secretsManager.Get(name, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, nil, apperrors.NotFound("secret %s not found", name)
		}
		return nil, nil, apperrors.Internal("failed to get %s secret, %s", name, err)
	}

	return secret.Data["ca.crt"], secret.Data["ca.key"], nil
}
