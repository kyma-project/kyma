package client

import (
	"fmt"
	"time"

	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/config"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type backupClient struct {
	coreClient *kubernetes.Clientset
}

type BackupClient interface {
	CreateNamespace(name string) error
	DeleteNamespace(name string) error
	WaitForNamespaceToBeDeleted(name string, waitmax time.Duration) error
}

func NewBackupClient() (BackupClient, error) {
	config, err := config.NewRestClientConfig()
	if err != nil {
		return nil, err
	}

	coreClientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &backupClient{
		coreClient: coreClientSet,
	}, nil

}

func (c *backupClient) CreateNamespace(name string) error {
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"env":             "true",
				"istio-injection": "enabled",
				"test":            "backup-restore",
			},
		},
		Spec: corev1.NamespaceSpec{},
	}
	_, err := c.coreClient.CoreV1().Namespaces().Create(namespace)
	return err
}

func (c *backupClient) DeleteNamespace(name string) error {
	return c.coreClient.CoreV1().Namespaces().Delete(name, &metav1.DeleteOptions{})
}

func (c *backupClient) WaitForNamespaceToBeDeleted(name string, waitmax time.Duration) error {
	timeout := time.After(waitmax)
	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-timeout:
			return fmt.Errorf("Namespace not deleted within given time  %v", waitmax)
		case <-ticker.C:
			if _, err := c.coreClient.CoreV1().Namespaces().Get(name, metav1.GetOptions{}); errors.IsNotFound(err) {
				return nil
			}
		}
	}
}
