package utils

import (
	"fmt"
	"os"
	"time"

	backupv1 "github.com/heptio/ark/pkg/apis/ark/v1"
	backup "github.com/heptio/ark/pkg/generated/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	api "k8s.io/kubernetes/pkg/apis/core"
)

type backupClient struct {
	backupClient *backup.Clientset
	coreClient   *kubernetes.Clientset
}

type BackupClient interface {
	CreateBackup(backupName string, includedNamespaces, excludedNamespaces, includedResources, excludedResources []string) error
	RestoreBackup(backupName string) error
	GetBackupStatus(backupName string) string
	CreateNamespace(name string) error
	DeleteNamespace(name string) error
	WaitForNamespaceToBeDeleted(name string, waitmax time.Duration) error
	WaitForBackupToBeCreated(name string, waitmax time.Duration) error
	WaitForBackupToBeRestored(name string, waitmax time.Duration) error
}

func NewBackupClient() (BackupClient, error) {

	kubeconfig := os.Getenv("KUBECONFIG")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	backupClientSet, err := backup.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	coreClientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &backupClient{
		coreClient:   coreClientSet,
		backupClient: backupClientSet,
	}, nil

}

func (c *backupClient) CreateBackup(backupName string, includedNamespaces, excludedNamespaces, includedResources, excludedResources []string) error {
	backup := &backupv1.Backup{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Backup",
			APIVersion: backupv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: backupName,
		},
		Spec: backupv1.BackupSpec{
			IncludedNamespaces: includedNamespaces,
			ExcludedNamespaces: excludedNamespaces,
			IncludedResources:  includedResources,
			ExcludedResources:  excludedResources,
		},
	}
	_, err := c.backupClient.ArkV1().Backups("heptio-ark").Create(backup)
	return err
}

func (c *backupClient) WaitForBackupToBeCreated(backupName string, waitmax time.Duration) error {
	backupWatch, err := c.backupClient.ArkV1().Backups("heptio-ark").Watch(metav1.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(api.ObjectNameField, backupName).String(),
	})
	if err != nil {
		return err
	}
	timeout := time.After(waitmax)

	for {
		select {
		case <-timeout:
			return fmt.Errorf("Backup could not be created within given time  %v", waitmax)
		case event := <-backupWatch.ResultChan():
			if event.Type == "ERROR" {
				return fmt.Errorf("%+v", event)
			}
			if event.Type == "MODIFIED" {
				backup, ok := event.Object.(*backupv1.Backup)
				if !ok {
					return fmt.Errorf("%v", event)
				}
				if backup.Status.Phase == "Completed" {
					return nil
				}
			}
		}
	}
}

func (c *backupClient) WaitForBackupToBeRestored(backupName string, waitmax time.Duration) error {
	restoreWatch, err := c.backupClient.ArkV1().Restores("heptio-ark").Watch(metav1.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(api.ObjectNameField, backupName).String(),
	})
	if err != nil {
		return err
	}
	timeout := time.After(waitmax)

	for {
		select {
		case <-timeout:
			return fmt.Errorf("Backup could not be restored within given time  %v", waitmax)
		case event := <-restoreWatch.ResultChan():
			if event.Type == "ERROR" {
				return fmt.Errorf("%+v", event)
			}
			if event.Type == "MODIFIED" {
				restore, ok := event.Object.(*backupv1.Restore)
				if !ok {
					fmt.Errorf("%v", event)
				}
				if restore.Status.Phase == "Completed" {
					return nil
				}
			}
		}
	}
}

func (c *backupClient) GetBackupStatus(backupName string) string {
	backup, err := c.backupClient.ArkV1().Backups("heptio-ark").Get(backupName, metav1.GetOptions{})
	if err != nil {
		return ""
	}
	return string(backup.Status.Phase)
}

func (c *backupClient) RestoreBackup(backupName string) error {
	restore := &backupv1.Restore{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Restore",
			APIVersion: backupv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: backupName,
		},
		Spec: backupv1.RestoreSpec{
			BackupName: backupName,
		},
	}
	_, err := c.backupClient.ArkV1().Restores("heptio-ark").Create(restore)
	return err
}

func (c *backupClient) CreateNamespace(name string) error {
	namespace := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namepspace",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"env":             "true",
				"istio-injection": "enabled",
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
	mywatch, err := c.coreClient.CoreV1().Namespaces().Watch(metav1.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(api.ObjectNameField, name).String(),
	})
	if err != nil {
		return err
	}

	timeout := time.After(waitmax)

	for {
		select {
		case <-timeout:
			return fmt.Errorf("Namespace not deleted within given time  %v", waitmax)
		case event := <-mywatch.ResultChan():
			if event.Type == "ERROR" {
				return fmt.Errorf("Could not delete namespace")
			}
			if event.Type == "DELETED" {
				return nil
			}
		}
	}
}
