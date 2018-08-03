package backup_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	etcdTypes "github.com/coreos/etcd-operator/pkg/apis/etcd/v1beta2"
	etcdOpFakeClientset "github.com/coreos/etcd-operator/pkg/generated/clientset/versioned/fake"
	etcdOpClient "github.com/coreos/etcd-operator/pkg/generated/clientset/versioned/typed/etcd/v1beta2"
	etcdOpInformers "github.com/coreos/etcd-operator/pkg/generated/informers/externalversions"
	"github.com/coreos/etcd-operator/pkg/generated/informers/externalversions/etcd/v1beta2"
	etcdOpLister "github.com/coreos/etcd-operator/pkg/generated/listers/etcd/v1beta2"
	"github.com/kyma-project/kyma/tools/etcd-backup/internal/backup"
	"github.com/kyma-project/kyma/tools/etcd-backup/internal/platform/idprovider"
	"github.com/kyma-project/kyma/tools/etcd-backup/internal/platform/logger/spy"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

func TestExecutorSingleBackupSuccess(t *testing.T) {
	// given
	tc := newExecutorTestCase(t)
	tc.StartAndWaitForInformers()

	sut := backup.NewExecutor(tc.NewExecutorParams()).
		WithIdProvider(fixIDProvider(tc.id))

	tc.MarkEtcdBackupAsSucceededAfterCreation()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// when
	out, err := sut.SingleBackup(ctx.Done(), tc.fixBlobPrefix)

	// then
	require.NoError(t, err)
	assert.Equal(t, tc.FixSingleBackupOutput(), out)

	performedActions := filterOutInformerActions(tc.etcdOpCli.Actions())
	assert.True(t, containsAction(createEtcdBackupAction(tc.FixEtcdBackupCR()), performedActions), "")
	assert.True(t, containsAction(deleteEtcdBackupAction(tc.FixEtcdBackupCR()), performedActions), "")

	assert.Empty(t, tc.logErrSink.DumpAll())
}

func TestExecutorSingleBackupFailure(t *testing.T) {
	t.Run("when creating backup name", func(t *testing.T) {
		// given
		tc := newExecutorTestCase(t)

		failingIDProvider := func() (string, error) { return "", errors.New("fix Err") }

		sut := backup.NewExecutor(tc.NewExecutorParams()).
			WithIdProvider(failingIDProvider)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// when
		out, err := sut.SingleBackup(ctx.Done(), tc.fixBlobPrefix)

		// then
		assert.EqualError(t, err, "while creating the EtcdBackup CR tmpl: while creating ID: fix Err")
		assert.Nil(t, out)

		assert.Empty(t, tc.logErrSink.DumpAll())
	})

	t.Run("when creating EtcdBackup CR", func(t *testing.T) {
		// given
		tc := newExecutorTestCase(t)

		sut := backup.NewExecutor(tc.NewExecutorParams()).
			WithIdProvider(fixIDProvider(tc.id))

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		tc.etcdOpCli.PrependReactor(failingRector(onCreateEtcdBackup()))

		// when
		out, err := sut.SingleBackup(ctx.Done(), tc.fixBlobPrefix)

		// then
		assertErrorContainsStatement(t, err, "while creating the EtcdBackup custom resource in k8s")
		assert.Nil(t, out)

		assert.Empty(t, tc.logErrSink.DumpAll())
	})

	t.Run("when deleting EtcdBackup CR", func(t *testing.T) {
		// given
		tc := newExecutorTestCase(t)
		tc.StartAndWaitForInformers()

		sut := backup.NewExecutor(tc.NewExecutorParams()).
			WithIdProvider(fixIDProvider(tc.id))

		tc.MarkEtcdBackupAsSucceededAfterCreation()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		tc.etcdOpCli.PrependReactor(failingRector(onDeleteEtcdBackup()))

		// when
		out, err := sut.SingleBackup(ctx.Done(), tc.fixBlobPrefix)

		// then
		assert.NoError(t, err)
		assert.Equal(t, tc.FixSingleBackupOutput(), out)

		// EtcdBackup "" - quotes are empty because we are using GenerateName functionality but it's not
		// covered by fake client, so the Name for created EtcdBackup CR is empty in real scenario it should
		// be populated.
		tc.logErrSink.AssertErrorLogged(t, `Cannot delete EtcdBackup "", got err: custom error`)
	})
}

type executorTestCase struct {
	t *testing.T

	id               string
	fixBlobPrefix    string
	fixContainerName string
	fixEtcdEndpoints []string
	fixABSSecretName string

	fixNsName             string
	etcdOpCli             *etcdOpFakeClientset.Clientset
	etcdOpInformerFactory etcdOpInformers.SharedInformerFactory
	etcdOpInformersGroup  v1beta2.Interface
	etcdOpNsScopedLister  etcdOpLister.EtcdBackupNamespaceLister
	etcdOpNsScopedCli     etcdOpClient.EtcdBackupInterface

	logErrSink *spy.LogSink
}

func newExecutorTestCase(t *testing.T) *executorTestCase {
	var (
		fixNsName             = "ns-name"
		etcdOpCli             = etcdOpFakeClientset.NewSimpleClientset()
		etcdOpInformerFactory = etcdOpInformers.NewFilteredSharedInformerFactory(etcdOpCli, 0, fixNsName, nil)
		etcdOpInformersGroup  = etcdOpInformerFactory.Etcd().V1beta2()
		etcdOpNsScopedLister  = etcdOpInformersGroup.EtcdBackups().Lister().EtcdBackups(fixNsName)
		etcdOpNsScopedCli     = etcdOpCli.EtcdV1beta2().EtcdBackups(fixNsName)
	)

	return &executorTestCase{
		t: t,

		id:               "123",
		fixBlobPrefix:    "fix-name-cnt",
		fixContainerName: "fix-blob-prefix",
		fixEtcdEndpoints: []string{"etcd-fix-endpoints"},
		fixABSSecretName: "dummy-fix-sec-name",

		fixNsName:             fixNsName,
		etcdOpCli:             etcdOpCli,
		etcdOpInformerFactory: etcdOpInformerFactory,
		etcdOpInformersGroup:  etcdOpInformersGroup,
		etcdOpNsScopedLister:  etcdOpNsScopedLister,
		etcdOpNsScopedCli:     etcdOpNsScopedCli,
		logErrSink:            newLogSinkForErrors(),
	}
}

func (tc *executorTestCase) StartAndWaitForInformers() {
	stopCh := make(<-chan struct{})
	tc.etcdOpInformerFactory.Start(stopCh)
	tc.etcdOpInformerFactory.WaitForCacheSync(stopCh)
}

func (tc *executorTestCase) MarkEtcdBackupAsSucceededAfterCreation() {
	tc.etcdOpInformersGroup.EtcdBackups().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			casted := obj.(*etcdTypes.EtcdBackup)

			cpy := casted.DeepCopy()
			cpy.Status.Succeeded = true
			_, err := tc.etcdOpNsScopedCli.Update(cpy)
			require.NoError(tc.t, err)
		},
	})

}

func (tc *executorTestCase) NewExecutorParams() (backup.Config, string, string, etcdOpClient.EtcdBackupInterface, etcdOpLister.EtcdBackupNamespaceLister, logrus.FieldLogger) {
	cfg := backup.Config{EtcdEndpoints: tc.fixEtcdEndpoints}

	return cfg, tc.fixABSSecretName, tc.fixContainerName, tc.etcdOpNsScopedCli, tc.etcdOpNsScopedLister, tc.logErrSink.Logger
}

func (tc *executorTestCase) FixSingleBackupOutput() *backup.SingleBackupOutput {
	return &backup.SingleBackupOutput{
		ABSBackupPath: fmt.Sprintf("%s/%s/%s", tc.fixContainerName, tc.fixBlobPrefix, tc.id),
	}
}

func (tc *executorTestCase) FixEtcdBackupCR() *etcdTypes.EtcdBackup {
	return &etcdTypes.EtcdBackup{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "automated-etcd-backup-job-", // used to generate a unique name
		},
		Spec: etcdTypes.BackupSpec{
			EtcdEndpoints: tc.fixEtcdEndpoints,

			StorageType: etcdTypes.BackupStorageTypeABS,
			BackupSource: etcdTypes.BackupSource{
				ABS: &etcdTypes.ABSBackupSource{
					ABSSecret: tc.fixABSSecretName,
					Path:      fmt.Sprintf("%s/%s/%s", tc.fixContainerName, tc.fixBlobPrefix, tc.id),
				},
			},
		},
	}

}

func fixIDProvider(id string) idprovider.Fn {
	return func() (string, error) {
		return id, nil
	}
}

func onCreateEtcdBackup() (string, string) {
	return "create", "etcdbackups"
}

func onDeleteEtcdBackup() (string, string) {
	return "delete", "etcdbackups"
}
