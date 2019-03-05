package utils

import (
	"fmt"
	"io/ioutil"
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

	"github.com/ghodss/yaml"
)

type backupClient struct {
	backupClient *backup.Clientset
	coreClient   *kubernetes.Clientset
}

type BackupClient interface {
	CreateBackup(backupName, backupSpecPath string) error
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

func (c *backupClient) CreateBackup(backupName, specPath string) error {
	var backupSpec backupv1.BackupSpec
	fileBytes, err := ioutil.ReadFile(specPath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(fileBytes, &backupSpec)
	if err != nil {
		return err
	}

	backup := &backupv1.Backup{
		ObjectMeta: metav1.ObjectMeta{
			Name: backupName,
		},
		Spec: backupSpec,
	}
	_, err = c.backupClient.ArkV1().Backups("heptio-ark").Create(backup)
	return err
}

func (c *backupClient) WaitForBackupToBeCreated(backupName string, waitmax time.Duration) error {
	timeout := time.After(waitmax)
	tick := time.Tick(2 * time.Second)

	for {
		select {
		case <-timeout:
			return fmt.Errorf("Backup %v could not be created within given time  %v", backupName, waitmax)
		case <-tick:
			backup, err := c.backupClient.ArkV1().Backups("heptio-ark").Get(backupName, metav1.GetOptions{})
			if err != nil {
				return err
			}
			if backup.Status.Phase == backupv1.BackupPhaseCompleted {
				return nil
			}
			if backup.Status.Phase == backupv1.BackupPhaseFailed || backup.Status.Phase == backupv1.BackupPhaseFailedValidation {
				return fmt.Errorf("Backup %v Failed with status %v :\n%+v", backupName, backup.Status.Phase, backup)
			}
		}
	}
}

func (c *backupClient) WaitForBackupToBeRestored(backupName string, waitmax time.Duration) error {
	timeout := time.After(waitmax)
	tick := time.Tick(2 * time.Second)

	for {
		select {
		case <-timeout:
			return fmt.Errorf("Backup %v could not be created within given time  %v", backupName, waitmax)
		case <-tick:
			restore, err := c.backupClient.ArkV1().Restores("heptio-ark").Get(backupName, metav1.GetOptions{})
			if err != nil {
				return err
			}
			if restore.Status.Phase == backupv1.RestorePhaseCompleted {
				return nil
			}
			if restore.Status.Phase == backupv1.RestorePhaseFailed || restore.Status.Phase == backupv1.RestorePhaseFailedValidation {
				return fmt.Errorf("Restore %v Failed with status %v :\n%+v", backupName, restore.Status.Phase, restore)
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
