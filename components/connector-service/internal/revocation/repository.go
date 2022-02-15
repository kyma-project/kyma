package revocation

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

type Manager interface {
	Get(ctx context.Context, name string, options metav1.GetOptions) (*v1.ConfigMap, error)
	Update(ctx context.Context, configmap *v1.ConfigMap, options metav1.UpdateOptions) (*v1.ConfigMap, error)
}

type RevocationListRepository interface {
	Insert(ctx context.Context, hash string) error
	Contains(ctx context.Context, hash string) (bool, error)
}

type revocationListRepository struct {
	configListManager Manager
	configMapName     string
}

func NewRepository(configListManager Manager, configMapName string) RevocationListRepository {
	return &revocationListRepository{
		configListManager: configListManager,
		configMapName:     configMapName,
	}
}

func (r *revocationListRepository) Insert(ctx context.Context, hash string) error {
	configMap, err := r.configListManager.Get(ctx, r.configMapName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	revokedCerts := configMap.Data
	if revokedCerts == nil {
		revokedCerts = map[string]string{}
	}
	revokedCerts[hash] = hash

	updatedConfigMap := configMap
	updatedConfigMap.Data = revokedCerts

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		_, err = r.configListManager.Update(ctx, updatedConfigMap, metav1.UpdateOptions{})
		return err
	})

	return err
}

func (r *revocationListRepository) Contains(ctx context.Context, hash string) (bool, error) {
	configMap, err := r.configListManager.Get(ctx, r.configMapName, metav1.GetOptions{})
	if err != nil {
		return false, err
	}

	found := false
	if configMap.Data != nil {
		_, found = configMap.Data[hash]
	}

	return found, nil
}
