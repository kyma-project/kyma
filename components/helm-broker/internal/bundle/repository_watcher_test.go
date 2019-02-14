package bundle_test

import (
	"testing"

	"errors"

	"github.com/knative/pkg/configmap"
	"github.com/kyma-project/kyma/components/helm-broker/internal/bundle"
	"github.com/kyma-project/kyma/components/helm-broker/internal/bundle/automock"
	"github.com/kyma-project/kyma/components/helm-broker/platform/logger/spy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

const (
	mapNamespace = "test"
	mapName      = "test"
	mapKey       = "test"
	mapTestData  = "data"
	brokerName   = "broker"
)

func TestRepositoryWatcher_StartWatchMapData(t *testing.T) {
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
	cfgMapInformer := configmap.NewInformedWatcher(client, mapNamespace)

	stopCh := make(chan struct{})
	watcher := bundle.NewRepositoryWatcher(bundleRemover, bundleSyncer, bundleLoader, brokerSyncer, brokerName, cfgMapInformer, logSink.Logger)

	// WHEN
	err := watcher.StartWatchMapData(mapName, mapKey, stopCh)
	// THEN
	assert.NoError(t, err)

}

func TestRepositoryWatcher_WhenUrlIsNotChanged(t *testing.T) {
	client := fake.NewSimpleClientset(fixConfigMap())
	cfgMapInformer := configmap.NewInformedWatcher(client, mapNamespace)
	logSink := spy.NewLogSink()
	stopCh := make(chan struct{})

	watcher := bundle.NewRepositoryWatcher(nil, nil, nil, nil, brokerName, cfgMapInformer, logSink.Logger)
	watcher.LastURL = mapTestData

	err := watcher.StartWatchMapData(mapName, mapKey, stopCh)
	assert.NoError(t, err)
}

func TestRepositoryWatcher_OnErrorWhenRemovingBundles(t *testing.T) {
	client := fake.NewSimpleClientset(fixConfigMap())
	cfgMapInformer := configmap.NewInformedWatcher(client, mapNamespace)
	logSink := spy.NewLogSink()
	stopCh := make(chan struct{})

	bundleRemover := &automock.BundleRemover{}
	defer bundleRemover.AssertExpectations(t)
	bundleRemover.On("RemoveAll").Return(errors.New("fix")).Once()

	watcher := bundle.NewRepositoryWatcher(bundleRemover, nil, nil, nil, brokerName, cfgMapInformer, logSink.Logger)

	err := watcher.StartWatchMapData(mapName, mapKey, stopCh)
	assert.NoError(t, err)
}

func fixConfigMap() *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: v12.ObjectMeta{
			Name:      mapName,
			Namespace: mapNamespace,
		},
		Data: map[string]string{
			mapKey: mapTestData,
		},
	}
}
