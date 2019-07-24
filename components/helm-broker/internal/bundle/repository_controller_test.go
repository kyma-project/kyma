package bundle_test

import (
	"testing"

	"context"
	"time"

	"errors"
	"fmt"

	"github.com/kyma-project/kyma/components/helm-broker/internal/bundle"
	"github.com/kyma-project/kyma/components/helm-broker/internal/bundle/automock"
	"github.com/kyma-project/kyma/components/helm-broker/platform/logger/spy"
	"github.com/sirupsen/logrus"
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
	mapTestData   = "https://example.com"
	mapLabelKey   = "repo"
	mapLabelValue = "true"
	brokerName    = "broker"
)

func TestNewRepositoryController_HappyPath(t *testing.T) {
	// GIVEN
	bundleSyncer := &automock.BundleSyncer{}
	bundleLoader := &automock.BundleLoader{}
	brokerSyncer := &automock.BrokerSyncer{}

	defer func() {
		bundleSyncer.AssertExpectations(t)
		bundleLoader.AssertExpectations(t)
		brokerSyncer.AssertExpectations(t)
	}()

	bundleSyncer.On("AddProvider", mapTestData, mock.Anything).Return(nil).Once()
	bundleSyncer.On("Execute").Return(nil).Once()
	bundleSyncer.On("CleanProviders").Once()
	brokerSyncer.On("Sync", brokerName, 5).Return(nil).Once()

	logSink := spy.NewLogSink()
	client := fake.NewSimpleClientset(fixConfigMap())
	cfgMapInformer := corev1.NewConfigMapInformer(client, mapNamespace, time.Minute, cache.Indexers{})

	asyncOpDone := make(chan struct{})
	hookAsyncOp := func() {
		asyncOpDone <- struct{}{}
	}

	controller := bundle.NewRepositoryController(bundleSyncer, bundleLoader, brokerSyncer, brokerName, cfgMapInformer, logSink.Logger, false).
		WithTestHookOnAsyncOpDone(hookAsyncOp)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// WHEN
	go cfgMapInformer.Run(ctx.Done())
	go controller.Run(ctx.Done())

	// THEN
	awaitForChanAtMost(t, asyncOpDone, time.Second)
}

func TestNewRepositoryController_ManyProviders(t *testing.T) {
	// GIVEN
	bundleSyncer := &automock.BundleSyncer{}
	bundleLoader := &automock.BundleLoader{}
	brokerSyncer := &automock.BrokerSyncer{}

	defer func() {
		bundleSyncer.AssertExpectations(t)
		bundleLoader.AssertExpectations(t)
		brokerSyncer.AssertExpectations(t)
	}()

	bundleSyncer.On("AddProvider", mapTestData, mock.Anything).Return(nil).Times(2)
	bundleSyncer.On("CleanProviders").Once()
	bundleSyncer.On("Execute").Return(nil).Once()
	brokerSyncer.On("Sync", brokerName, 5).Return(nil).Once()

	cfgMap := fixConfigMap()
	cfgMap.Data = map[string]string{
		"URLs": mapTestData + "\n" + mapTestData + "\n" + "http://example.com",
	}

	logSink := spy.NewLogSink()
	client := fake.NewSimpleClientset(cfgMap)
	cfgMapInformer := corev1.NewConfigMapInformer(client, mapNamespace, time.Minute, cache.Indexers{})

	asyncOpDone := make(chan struct{})
	hookAsyncOp := func() {
		asyncOpDone <- struct{}{}
	}

	controller := bundle.NewRepositoryController(bundleSyncer, bundleLoader, brokerSyncer, brokerName, cfgMapInformer, logSink.Logger, false).
		WithTestHookOnAsyncOpDone(hookAsyncOp)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// WHEN
	go cfgMapInformer.Run(ctx.Done())
	go controller.Run(ctx.Done())

	// THEN
	awaitForChanAtMost(t, asyncOpDone, time.Second)
}

func TestNewRepositoryController_OutOfRetries(t *testing.T) {
	// GIVEN
	bundleSyncer := &automock.BundleSyncer{}
	bundleLoader := &automock.BundleLoader{}
	brokerSyncer := &automock.BrokerSyncer{}

	defer func() {
		bundleSyncer.AssertExpectations(t)
		bundleLoader.AssertExpectations(t)
		brokerSyncer.AssertExpectations(t)
	}()

	fixErr := "fix"
	bundleSyncer.On("AddProvider", mapTestData, mock.Anything).Return(nil).Once()
	bundleSyncer.On("CleanProviders").Once()
	bundleSyncer.On("Execute").Return(errors.New(fixErr)).Once()

	logSink := spy.NewLogSink()
	client := fake.NewSimpleClientset(fixConfigMap())
	cfgMapInformer := corev1.NewConfigMapInformer(client, mapNamespace, time.Minute, cache.Indexers{})

	asyncOpDone := make(chan struct{})
	hookAsyncOp := func() {
		asyncOpDone <- struct{}{}
	}

	controller := bundle.NewRepositoryController(bundleSyncer, bundleLoader, brokerSyncer, brokerName, cfgMapInformer, logSink.Logger, false).
		WithTestHookOnAsyncOpDone(hookAsyncOp).
		WithoutRetries()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// WHEN
	go cfgMapInformer.Run(ctx.Done())
	go controller.Run(ctx.Done())

	// THEN
	awaitForChanAtMost(t, asyncOpDone, time.Second)
	logSink.AssertLogged(t, logrus.ErrorLevel, fmt.Sprintf("Error processing %q (giving up - to many retires): while syncing bundles: %v", mapName+"/"+mapNamespace, fixErr))
}

func TestNewRepositoryController_UnsecuredRepositories(t *testing.T) {
	// GIVEN
	bundleSyncer := &automock.BundleSyncer{}
	bundleLoader := &automock.BundleLoader{}
	brokerSyncer := &automock.BrokerSyncer{}

	defer func() {
		bundleSyncer.AssertExpectations(t)
		bundleLoader.AssertExpectations(t)
		brokerSyncer.AssertExpectations(t)
	}()

	bundleSyncer.On("CleanProviders").Once()

	cfgMap := fixConfigMap()
	cfgMap.Data = map[string]string{
		"URLs": "http://example.com",
	}

	logSink := spy.NewLogSink()
	client := fake.NewSimpleClientset(cfgMap)
	cfgMapInformer := corev1.NewConfigMapInformer(client, mapNamespace, time.Minute, cache.Indexers{})

	asyncOpDone := make(chan struct{})
	hookAsyncOp := func() {
		asyncOpDone <- struct{}{}
	}

	controller := bundle.NewRepositoryController(bundleSyncer, bundleLoader, brokerSyncer, brokerName, cfgMapInformer, logSink.Logger, false).
		WithTestHookOnAsyncOpDone(hookAsyncOp).
		WithoutRetries()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// WHEN
	go cfgMapInformer.Run(ctx.Done())
	go controller.Run(ctx.Done())

	// THEN
	awaitForChanAtMost(t, asyncOpDone, time.Second)
	logSink.AssertLogged(t, logrus.WarnLevel, fmt.Sprintf("Cannot create repositories for %s/%s: All Repository URLs are incorrect or unsecured", mapName, mapNamespace))
}

func TestNewRepositoryController_UnsecuredRepositoriesWithDevelopMode(t *testing.T) {
	// GIVEN
	bundleSyncer := &automock.BundleSyncer{}
	bundleLoader := &automock.BundleLoader{}
	brokerSyncer := &automock.BrokerSyncer{}

	defer func() {
		bundleSyncer.AssertExpectations(t)
		bundleLoader.AssertExpectations(t)
		brokerSyncer.AssertExpectations(t)
	}()
	unsecureMapTestData := "http://example.com"

	bundleSyncer.On("AddProvider", unsecureMapTestData, mock.Anything).Return(nil).Once()
	bundleSyncer.On("CleanProviders").Once()
	bundleSyncer.On("Execute").Return(nil).Once()
	brokerSyncer.On("Sync", brokerName, 5).Return(nil).Once()

	cfgMap := fixConfigMap()
	cfgMap.Data = map[string]string{
		"URLs": unsecureMapTestData,
	}

	logSink := spy.NewLogSink()
	client := fake.NewSimpleClientset(cfgMap)
	cfgMapInformer := corev1.NewConfigMapInformer(client, mapNamespace, time.Minute, cache.Indexers{})

	asyncOpDone := make(chan struct{})
	hookAsyncOp := func() {
		asyncOpDone <- struct{}{}
	}

	controller := bundle.NewRepositoryController(bundleSyncer, bundleLoader, brokerSyncer, brokerName, cfgMapInformer, logSink.Logger, true).
		WithTestHookOnAsyncOpDone(hookAsyncOp).
		WithoutRetries()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// WHEN
	go cfgMapInformer.Run(ctx.Done())
	go controller.Run(ctx.Done())

	// THEN
	awaitForChanAtMost(t, asyncOpDone, time.Second)
}

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
