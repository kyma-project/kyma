package backup_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kyma-project/kyma/tools/etcd-backup/internal/backup"
	"github.com/kyma-project/kyma/tools/etcd-backup/internal/backup/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	coreTypes "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRecordedExecutorSingleBackupSuccess(t *testing.T) {
	// given
	fixCfgMapName := "recorded-etcd-backup-data"
	fixBlobPrefix := "blob-prefix"
	fixBackupOut := &backup.SingleBackupOutput{
		ABSBackupPath: "abs/backup/path",
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	backupExecutorMock := automock.NewSingleBackupExecutor()
	defer backupExecutorMock.AssertExpectations(t)
	backupExecutorMock.ExpectOnSingleBackup(ctx.Done(), fixBlobPrefix, fixBackupOut)

	cfgMapCliMock := automock.NewConfigMapClient()
	defer cfgMapCliMock.AssertExpectations(t)
	cfgMapCliMock.ExpectOnCreate(&coreTypes.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: fixCfgMapName,
		},
		Data: map[string]string{
			"abs-backup-file-path-from-last-success": fixBackupOut.ABSBackupPath,
		},
	})

	sut := backup.NewRecordedExecutor(backupExecutorMock, fixCfgMapName, cfgMapCliMock)

	// when
	out, err := sut.SingleBackup(ctx.Done(), fixBlobPrefix)

	// then
	require.NoError(t, err)
	assert.Equal(t, fixBackupOut, out)
}

func TestRecordedExecutorSingleBackupFailure(t *testing.T) {
	// given
	var (
		fixBlobPrefix = "blob-prefix"
		fixErr        = errors.New("fix ERR")
		fixBackupOut  = func() *backup.SingleBackupOutput {
			return &backup.SingleBackupOutput{
				ABSBackupPath: "abs/backup/path",
			}
		}
	)

	t.Run("Error returned from underlying component", func(t *testing.T) {
		fixBOut := fixBackupOut()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		backupExecutorMock := automock.NewSingleBackupExecutor()
		defer backupExecutorMock.AssertExpectations(t)
		backupExecutorMock.ExpectErrorOnSingleBackup(fixBOut, fixErr)

		sut := backup.NewRecordedExecutor(backupExecutorMock, "", nil)

		// when
		out, err := sut.SingleBackup(ctx.Done(), fixBlobPrefix)

		// then
		assert.Error(t, err, fixErr)
		assert.Equal(t, fixBOut, out)
	})

	t.Run("Error when saving to config map", func(t *testing.T) {
		fixBOut := fixBackupOut()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		backupExecutorMock := automock.NewSingleBackupExecutor()
		defer backupExecutorMock.AssertExpectations(t)
		backupExecutorMock.ExpectOnSingleBackup(ctx.Done(), fixBlobPrefix, fixBOut)

		cfgMapCliMock := automock.NewConfigMapClient()
		defer cfgMapCliMock.AssertExpectations(t)
		cfgMapCliMock.ExpectErrorOnCreate(fixErr)

		sut := backup.NewRecordedExecutor(backupExecutorMock, "fix-cfg-map-name", cfgMapCliMock)

		// when
		out, err := sut.SingleBackup(ctx.Done(), fixBlobPrefix)

		// then
		assert.Error(t, err, fixErr)
		assert.Equal(t, fixBOut, out)
	})
}
