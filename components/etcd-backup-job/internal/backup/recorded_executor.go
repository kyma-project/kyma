package backup

import (
	"github.com/pkg/errors"
	coreTypes "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const absBackupCfgMapKeyName = "abs-backup-file-path-from-last-success"

//go:generate mockery -name=configMapClient -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=singleBackupExecutor -output=automock -outpkg=automock -case=underscore

type (
	configMapClient interface {
		Create(*coreTypes.ConfigMap) (*coreTypes.ConfigMap, error)
		Update(*coreTypes.ConfigMap) (*coreTypes.ConfigMap, error)
		Get(name string, options metav1.GetOptions) (*coreTypes.ConfigMap, error)
	}

	singleBackupExecutor interface {
		SingleBackup(stopCh <-chan struct{}, blobPrefix string) (*SingleBackupOutput, error)
	}
)

// RecordedExecutor saves in k8s ConfigMap path to the backup file when backup process ends without errors.
type RecordedExecutor struct {
	underlying singleBackupExecutor
	cfgMapCli  configMapClient
	cfgMapName string
}

// NewRecordedExecutor returns new instance of RecordedExecutor
func NewRecordedExecutor(underlying singleBackupExecutor, cfgMapName string, cfgMapCli configMapClient) *RecordedExecutor {
	return &RecordedExecutor{
		underlying: underlying,
		cfgMapCli:  cfgMapCli,
		cfgMapName: cfgMapName,
	}
}

// SingleBackup executes underlying SingleBackup method and if process ends without errors then
// path to the backup file is save in k8s ConfigMap
// BEWARE: It can return own error when it cannot save information to ConfigMap.
func (r *RecordedExecutor) SingleBackup(stopCh <-chan struct{}, blobPrefix string) (*SingleBackupOutput, error) {
	out, err := r.underlying.SingleBackup(stopCh, blobPrefix)
	if err != nil {
		return out, err
	}

	if err := r.upsertABSBackupPathToCfgMap(out.ABSBackupPath); err != nil {
		return out, errors.Wrap(err, "while upserting the newest ABS backup path to config map")
	}

	return out, nil
}

func (r *RecordedExecutor) upsertABSBackupPathToCfgMap(path string) error {
	err := r.createCfgMapWithABSBackupPath(path)
	switch {
	case err == nil:
	case apiErrors.IsAlreadyExists(err):
		r.updateCfgMapWithABSBackupPath(path)
	default:
		return errors.Wrapf(err, "while creating %s config map", r.cfgMapName)
	}

	return nil
}

func (r *RecordedExecutor) createCfgMapWithABSBackupPath(path string) error {
	_, err := r.cfgMapCli.Create(&coreTypes.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: r.cfgMapName,
		},
		Data: map[string]string{
			absBackupCfgMapKeyName: path,
		},
	})

	return err
}

func (r *RecordedExecutor) updateCfgMapWithABSBackupPath(path string) error {
	oldCfg, err := r.cfgMapCli.Get(r.cfgMapName, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "while getting %s config map", r.cfgMapName)
	}

	cfgCopy := oldCfg.DeepCopy()
	cfgCopy.Data = r.ensureMapIsInitiated(cfgCopy.Data)
	cfgCopy.Data[absBackupCfgMapKeyName] = path

	if _, err := r.cfgMapCli.Update(cfgCopy); err != nil {
		return errors.Wrapf(err, "while updating %s config map", r.cfgMapName)
	}

	return nil
}

// ensureMapIsInitiated ensures that given map is initiated.
// - returns given map if it's already allocated
// - otherwise returns empty map
func (r *RecordedExecutor) ensureMapIsInitiated(m map[string]string) map[string]string {
	if m == nil {
		empty := make(map[string]string)
		return empty
	}

	return m
}
