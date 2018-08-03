package backup

import (
	"fmt"
	"strings"

	etcdTypes "github.com/coreos/etcd-operator/pkg/apis/etcd/v1beta2"
	etcdOpClient "github.com/coreos/etcd-operator/pkg/generated/clientset/versioned/typed/etcd/v1beta2"
	etcdOpLister "github.com/coreos/etcd-operator/pkg/generated/listers/etcd/v1beta2"
	"github.com/kyma-project/kyma/tools/etcd-backup/internal/platform/idprovider"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const etcdBackupGenerateName = "automated-etcd-backup-job-"

// Executor executes the backup process of etcd database
type Executor struct {
	log              logrus.FieldLogger
	etcdBackupCli    etcdOpClient.EtcdBackupInterface
	etcdBackupLister etcdOpLister.EtcdBackupNamespaceLister

	absContainerName string
	etcdEndpoints    []string
	absSecretName    string

	idProvider idprovider.Fn
}

// NewExecutor returns new instance of Executor
func NewExecutor(cfg Config, absSecretName, absContainerName string, etcdBackupCli etcdOpClient.EtcdBackupInterface, nsScopedLister etcdOpLister.EtcdBackupNamespaceLister, log logrus.FieldLogger) *Executor {
	return &Executor{
		log:              log,
		etcdBackupCli:    etcdBackupCli,
		etcdBackupLister: nsScopedLister,
		absSecretName:    absSecretName,
		etcdEndpoints:    cfg.EtcdEndpoints,
		idProvider:       idprovider.New(),
		absContainerName: removeSlashSuffixIfNeeded(absContainerName),
	}
}

// SingleBackupOutput contains all information which will be returned from SingleBackup function
type SingleBackupOutput struct {
	ABSBackupPath string
}

// SingleBackup executes backup process for given etcd-cluster and waits for its status
func (e *Executor) SingleBackup(stopCh <-chan struct{}, blobPrefix string) (*SingleBackupOutput, error) {
	e.log.Debug("Starting backup process")
	defer e.log.Debug("Backup process completed")

	etcdBackupTmpl, err := e.etcdBackup(blobPrefix)
	if err != nil {
		return nil, errors.Wrap(err, "while creating the EtcdBackup CR tmpl")
	}

	etcdBackupCR, err := e.etcdBackupCli.Create(etcdBackupTmpl)
	if err != nil {
		return nil, errors.Wrap(err, "while creating the EtcdBackup custom resource in k8s")
	}
	defer func() {
		e.log.Debugf("Deleting EtcdBackup CR %q", etcdBackupCR.Name)
		// TODO: consider to return this error from function instead of logging it here (can be achieved by named error)
		if err := e.etcdBackupCli.Delete(etcdBackupCR.Name, nil); err != nil {
			e.log.Errorf("Cannot delete EtcdBackup %q, got err: %v", etcdBackupCR.Name, err)
		}
	}()

	e.log.Debugf("EtcdBackup CR %q created, waiting for status", etcdBackupCR.Name)
	if err := e.waitForEtcdBackupStatus(etcdBackupCR.Name, stopCh); err != nil {
		return nil, errors.Wrap(err, "while getting EtcdBackup status")
	}

	return &SingleBackupOutput{
		ABSBackupPath: etcdBackupCR.Spec.ABS.Path,
	}, nil
}

func (e *Executor) etcdBackup(blobPrefix string) (*etcdTypes.EtcdBackup, error) {
	id, err := e.idProvider()
	if err != nil {
		return nil, errors.Wrap(err, "while creating ID")
	}
	blobPrefix = removeSlashSuffixIfNeeded(blobPrefix)

	return &etcdTypes.EtcdBackup{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: etcdBackupGenerateName, // used to generate a unique name
		},
		Spec: etcdTypes.BackupSpec{
			EtcdEndpoints: e.etcdEndpoints,

			StorageType: etcdTypes.BackupStorageTypeABS,
			BackupSource: etcdTypes.BackupSource{
				ABS: &etcdTypes.ABSBackupSource{
					ABSSecret: e.absSecretName,
					Path:      fmt.Sprintf("%s/%s/%s", e.absContainerName, blobPrefix, id),
				},
			},
		},
	}, nil
}

func (e *Executor) waitForEtcdBackupStatus(name string, stopCh <-chan struct{}) error {
	for {
		if e.shouldExit(stopCh) {
			return errors.New("stop channel was called when waiting for EtcdBackup status")
		}

		backup, err := e.etcdBackupLister.Get(name)
		if apiErrors.IsNotFound(err) { // It's possible that cache was not synced yet
			continue
		}
		if err != nil {
			return errors.Wrap(err, "while getting EtcdBackup custom resource")
		}

		if backup.Status.Succeeded {
			return nil
		}

		if backup.Status.Reason != "" {
			return errors.Errorf("EtcdBackup failed with reason: %s", backup.Status.Reason)
		}
	}
}

// removeSlashSuffixIfNeeded ensures that path does not contain the slash suffix
func removeSlashSuffixIfNeeded(path string) string {
	if strings.HasSuffix(path, "/") {
		return path[:len(path)-1]
	}
	return path
}

func (e *Executor) shouldExit(stopCh <-chan struct{}) bool {
	select {
	case <-stopCh:
		return true
	default:
	}
	return false
}
