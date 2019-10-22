package revocation

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

type Manager interface {
	Get(name string, options metav1.GetOptions) (*v1.ConfigMap, error)
	Update(configmap *v1.ConfigMap) (*v1.ConfigMap, error)
}

type RevocationListRepository interface {
	Insert(hash string) error
	Contains(hash string) (bool, error)
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

func (r *revocationListRepository) Insert(hash string) error {
	configMap, err := r.configListManager.Get(r.configMapName, metav1.GetOptions{})
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
		_, err = r.configListManager.Update(updatedConfigMap)
		return err
	})

	return err
}

func (r *revocationListRepository) Contains(hash string) (bool, error) {
	configMap, err := r.configListManager.Get(r.configMapName, metav1.GetOptions{})
	if err != nil {
		return false, err
	}

	found := false
	if configMap.Data != nil {
		_, found = configMap.Data[hash]
	}

	return found, nil
}
