package testkit

import (
	"fmt"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type DataStore struct {
	namespace  string
	coreClient kubernetes.Interface
	name       string
}

func NewDataStore(coreClient *kubernetes.Clientset, namespace string) *DataStore {
	return &DataStore{
		namespace:  namespace,
		coreClient: coreClient,
		name:       fmt.Sprintf("state-%v", namespace),
	}
}

func (ds DataStore) Store(key, val string) error {
	found := true
	cm, err := ds.coreClient.CoreV1().ConfigMaps(ds.namespace).Get(ds.name, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		found = false
		cm = &core.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: ds.name,
			},
			Data: map[string]string{},
		}
	}

	cm.Data[key] = val
	if found {
		_, err = ds.coreClient.CoreV1().ConfigMaps(ds.namespace).Update(cm)
	} else {
		_, err = ds.coreClient.CoreV1().ConfigMaps(ds.namespace).Create(cm)
	}
	return err
}

func (ds DataStore) Load(key string) (string, error) {
	cm, err := ds.coreClient.CoreV1().ConfigMaps(ds.namespace).Get(ds.name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	if val, ok := cm.Data[key]; ok {
		return val, nil
	}
	return "", fmt.Errorf("key not found: %v", key)
}

func (ds DataStore) Destroy() error {
	return ds.coreClient.CoreV1().ConfigMaps(ds.namespace).Delete(ds.name, &metav1.DeleteOptions{})
}
