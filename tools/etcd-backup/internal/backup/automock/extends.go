package automock

import (
	"github.com/stretchr/testify/mock"
	"github.com/kyma-project/kyma/tools/etcd-backup/internal/backup"
	"k8s.io/api/core/v1"
)

// Single Backup Executor
func (_m *singleBackupExecutor) ExpectOnSingleBackup(stopCh <-chan struct{}, blobPrefix string, out *backup.SingleBackupOutput) *mock.Call {
	return _m.On("SingleBackup", stopCh, blobPrefix).Return(out, nil)
}

func (_m *singleBackupExecutor) ExpectErrorOnSingleBackup(out *backup.SingleBackupOutput, err error) *mock.Call {
	return _m.On("SingleBackup", mock.Anything, mock.Anything).Return(out, err)
}

// Config Map Client
func (_m *configMapClient) ExpectOnCreate(cfgMap *v1.ConfigMap) *mock.Call {
	return _m.On("Create", cfgMap).Return(cfgMap, nil)
}

func (_m *configMapClient) ExpectErrorOnCreate(err error) *mock.Call {
	return _m.On("Create", mock.Anything).Return(nil, err)
}