package backup

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"

	backupv1 "github.com/heptio/ark/pkg/apis/ark/v1"
	arkbackuppkg "github.com/heptio/ark/pkg/backup"
	backup "github.com/heptio/ark/pkg/generated/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"github.com/ghodss/yaml"
	"k8s.io/apimachinery/pkg/api/errors"
	"github.com/heptio/ark/pkg/cmd/util/output"
	"github.com/heptio/ark/pkg/restic"
	"github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/utils/config"
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
	DescribeBackup(name string) error
	DescribeRestore(name string) error
}

func NewBackupClient() (BackupClient, error) {
	config, err := config.NewRestClientConfig()
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

func (c *backupClient) DescribeBackup(backupName string) error {
	backup, err := c.backupClient.ArkV1().Backups("heptio-ark").Get(backupName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	deleteRequestListOptions := arkbackuppkg.NewDeleteBackupRequestListOptions(backup.Name, string(backup.UID))
	deleteRequestList, err := c.backupClient.ArkV1().DeleteBackupRequests("heptio-ark").List(deleteRequestListOptions)
	if err != nil {
		return err
	}

	opts := restic.NewPodVolumeBackupListOptions(backup.Name, string(backup.UID))
	podVolumeBackupList, err := c.backupClient.ArkV1().PodVolumeBackups("heptio-ark").List(opts)
	if err != nil {
		return err
	}

	s := output.DescribeBackup(backup, deleteRequestList.Items, podVolumeBackupList.Items, true, c.backupClient)
	log.Printf("========================== Begin Backup: %v ==========================\n", backupName)
	log.Println(s)
	log.Printf("========================== End Backup: %v ==========================\n", backupName)
	return nil
}

func (c *backupClient) DescribeRestore(backupName string) error {
	restore, err := c.backupClient.ArkV1().Restores("heptio-ark").Get(backupName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	opts := restic.NewPodVolumeRestoreListOptions(restore.Name, string(restore.UID))
	podvolumeRestoreList, err := c.backupClient.ArkV1().PodVolumeRestores("heptio-ark").List(opts)
	if err != nil {
		return err
	}

	s := output.DescribeRestore(restore, podvolumeRestoreList.Items, true, c.backupClient)
	log.Printf("========================== Begin Restore: %v ==========================\n", backupName)
	log.Println(s)
	log.Printf("========================== End Restore: %v ==========================\n", backupName)
	return nil
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
			IncludeClusterResources: c.ptrBool(true),
			RestorePVs: c.ptrBool(true),
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
	timeout := time.After(waitmax)
	tick := time.Tick(1 * time.Second)

	for {
		select {
		case <-timeout:
			return fmt.Errorf("Namespace not deleted within given time  %v", waitmax)
		case <-tick:
			if _, err := c.coreClient.CoreV1().Namespaces().Get(name, metav1.GetOptions{}); errors.IsNotFound(err) {
				return nil
			}
		}
	}
}

func (*backupClient) ptrBool(b bool) *bool {
	return &b
}