package testkit

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	cm, err := ds.coreClient.CoreV1().ConfigMaps(ds.namespace).Get(ds.name, meta.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		found = false
		cm = &core.ConfigMap{
			ObjectMeta: meta.ObjectMeta{
				Name: ds.name,
			},
			Data: map[string]string{},
		}
	}
	data := &cm.Data
	(*data)[key] = val
	if found {
		_, err = ds.coreClient.CoreV1().ConfigMaps(ds.namespace).Update(cm)
	} else {
		_, err = ds.coreClient.CoreV1().ConfigMaps(ds.namespace).Create(cm)
	}
	return err
}

func (ds DataStore) Load(key string) (string, error) {
	cm, err := ds.coreClient.CoreV1().ConfigMaps(ds.namespace).Get(ds.name, meta.GetOptions{})
	if err != nil {
		return "", err
	}
	if val, ok := cm.Data[key]; ok {
		return val, nil
	}
	return "", fmt.Errorf("key not found: %v", key)
}
