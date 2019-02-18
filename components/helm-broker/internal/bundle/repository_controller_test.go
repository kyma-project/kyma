package bundle_test

import (
	"testing"

	"context"
	"time"

	"github.com/kyma-project/kyma/components/helm-broker/internal/bundle"
	"github.com/kyma-project/kyma/components/helm-broker/internal/bundle/automock"
	"github.com/kyma-project/kyma/components/helm-broker/platform/logger/spy"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
)

const (
	mapNamespace  = "test"
	mapName       = "test"
	mapKey        = "URLs"
	mapTestData   = "data"
	mapLabelKey   = "repo"
	mapLabelValue = "true"
	brokerName    = "broker"
)

func TestNewRepositoryController_HappyPath(t *testing.T) {
	// GIVEN
	bundleRemover := &automock.BundleRemover{}
	bundleSyncer := &automock.BundleSyncer{}
	bundleLoader := &automock.BundleLoader{}
	brokerSyncer := &automock.BrokerSyncer{}

	defer func() {
		bundleSyncer.AssertExpectations(t)
		bundleRemover.AssertExpectations(t)
		bundleLoader.AssertExpectations(t)
		brokerSyncer.AssertExpectations(t)
	}()

	bundleRemover.On("RemoveAll").Return(nil).Once()
	bundleSyncer.On("CleanProviders").Return(nil).Once()
	bundleSyncer.On("AddProvider", mapTestData, mock.Anything).Return(nil).Once()
	bundleSyncer.On("Execute").Once()
	brokerSyncer.On("Sync", brokerName, 5).Return(nil).Once()

	logSink := spy.NewLogSink()
	client := fake.NewSimpleClientset(fixConfigMap())
	cfgMapInformer := corev1.NewConfigMapInformer(client, mapNamespace, time.Minute, cache.Indexers{})

	asyncOpDone := make(chan struct{})
	hookAsyncOp := func() {
		asyncOpDone <- struct{}{}
	}

	controller := bundle.NewRepositoryController(bundleRemover, bundleSyncer, bundleLoader, brokerSyncer, brokerName, cfgMapInformer, logSink.Logger).
		WithTestHookOnAsyncOpDone(hookAsyncOp)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// WHEN
	go cfgMapInformer.Run(ctx.Done())
	go controller.Run(ctx.Done())

	// THEN
	awaitForChanAtMost(t, asyncOpDone, 5*time.Second)
}

//func TestNewRepositoryController_HappyPath(t *testing.T) {
//	// GIVEN
//	bundleRemover := &automock.BundleRemover{}
//	bundleSyncer := &automock.BundleSyncer{}
//	bundleLoader := &automock.BundleLoader{}
//	brokerSyncer := &automock.BrokerSyncer{}
//
//	defer func() {
//		bundleSyncer.AssertExpectations(t)
//		bundleRemover.AssertExpectations(t)
//		bundleLoader.AssertExpectations(t)
//		brokerSyncer.AssertExpectations(t)
//	}()
//
//	bundleRemover.On("RemoveAll").Return(nil).Once()
//	bundleSyncer.On("CleanProviders").Return(nil).Once()
//	bundleSyncer.On("AddProvider", mapTestData, mock.Anything).Return(nil).Once()
//	bundleSyncer.On("Execute").Once()
//	brokerSyncer.On("Sync", brokerName, 5).Return(nil).Once()
//
//	logSink := spy.NewLogSink()
//	client := fake.NewSimpleClientset(fixConfigMap())
//	cfgMapInformer := corev1.NewConfigMapInformer(client, mapNamespace, time.Minute, cache.Indexers{})
//
//	asyncOpDone := make(chan struct{})
//	hookAsyncOp := func() {
//		asyncOpDone <- struct{}{}
//	}
//
//	controller := bundle.NewRepositoryController(bundleRemover, bundleSyncer, bundleLoader, brokerSyncer, brokerName, cfgMapInformer, logSink.Logger).
//		WithTestHookOnAsyncOpDone(hookAsyncOp)
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	// WHEN
//	go cfgMapInformer.Run(ctx.Done())
//	go controller.Run(ctx.Done())
//
//	// THEN
//	awaitForChanAtMost(t, asyncOpDone, 5*time.Second)
//}

func fixConfigMap() *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mapName,
			Namespace: mapNamespace,
			Labels: map[string]string{
				mapLabelKey: mapLabelValue,
			},
		},
		Data: map[string]string{
			mapKey: mapTestData,
		},
	}
}

func awaitForChanAtMost(t *testing.T, ch <-chan struct{}, timeout time.Duration) {
	select {
	case <-ch:
	case <-time.After(timeout):
		t.Fatalf("timeout occured when waiting for channel")
	}
}
