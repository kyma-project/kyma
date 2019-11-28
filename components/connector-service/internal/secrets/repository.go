package secrets

import (
	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type ManagerConstructor func(namespace string) Manager

type Manager interface {
	Get(name string, options metav1.GetOptions) (*v1.Secret, error)
}

type Repository interface {
	Get(name types.NamespacedName) (secretData map[string][]byte, appError apperrors.AppError)
}

type repository struct {
	secretsManagerConstructor ManagerConstructor
}

// NewRepository creates a new secrets repository
func NewRepository(secretsManagerConstructor ManagerConstructor) Repository {
	return &repository{
		secretsManagerConstructor: secretsManagerConstructor,
	}
}

func (r *repository) Get(name types.NamespacedName) (secretData map[string][]byte, appError apperrors.AppError) {
	secretsManager := r.secretsManagerConstructor(name.Namespace)
	secret, err := secretsManager.Get(name.Name, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, apperrors.NotFound("secret %s not found", name)
		}
		return nil, apperrors.Internal("failed to get %s secret, %s", name, err)
	}

	return secret.Data, nil
}
