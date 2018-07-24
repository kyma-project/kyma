package labeler_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/remote-environment-broker/internal"
	"github.com/kyma-project/kyma/components/remote-environment-broker/internal/labeler"
	"github.com/kyma-project/kyma/components/remote-environment-broker/internal/labeler/automock"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/remoteenvironment/v1alpha1"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/client/clientset/versioned/fake"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/client/informers/externalversions"
	"github.com/kyma-project/kyma/components/remote-environment-broker/platform/logger/spy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/informers/core/v1"
	fake2 "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
)

const (
	fixNSName = "production"
	fixREName = "ec-prod"
)

func TestControllerRunSuccess(t *testing.T) {

	// given
	fixEM := fixEnvironmentMappingCR(fixREName, fixNSName)
	fixRE := fixRemoteEnvironment(fixREName)
	fixNS := fixNamespace(fixNSName)

	expectations := &sync.WaitGroup{}
	expectations.Add(2)
	fulfillExpectation := func(mock.Arguments) {
		expectations.Done()
	}

	expectedPatchNS := fmt.Sprintf(`{"metadata":{"labels":{"accessLabel":"%s"}}}`, fixRE.AccessLabel)

	emInformer := fakeEMInformer(fixEM)
	nsInformer := fakeNSInformer(fixNS)

	nsClientMock := &automock.NsPatcher{}
	defer nsClientMock.AssertExpectations(t)
	nsClientMock.On("Patch", fixNSName, types.StrategicMergePatchType, []byte(expectedPatchNS)).
		Return(&corev1.Namespace{}, nil).
		Run(fulfillExpectation).
		Once()

	reGetterMock := &automock.ReGetter{}
	defer reGetterMock.AssertExpectations(t)
	reGetterMock.On("Get", internal.RemoteEnvironmentName(fixREName)).
		Return(&fixRE, nil).
		Run(fulfillExpectation).
		Once()

	svc := labeler.New(emInformer, nsInformer, nsClientMock, reGetterMock, spy.NewLogDummy())

	awaitInformerStartAtMost(t, time.Second, emInformer)
	awaitInformerStartAtMost(t, time.Second, nsInformer)

	ctx, close := context.WithCancel(context.Background())
	// when
	go svc.Run(ctx.Done())

	// then
	awaitForSyncGroupAtMost(t, expectations, time.Second)

	// clean-up - release controller
	close()
}

func TestControllerRunSuccessLabelRemove(t *testing.T) {
	// given
	fixEM := fixEnvironmentMappingCR(fixREName, fixNSName)
	fixNS := fixNamespaceWithAccessLabel(fixNSName)
	fixExpectedNS := fixNamespace(fixNSName)

	emInformer := fakeEMInformer(fixEM)
	nsClientMock := &automock.NsPatcher{}
	defer nsClientMock.AssertExpectations(t)

	deletedLabelNS := `{"metadata":{"labels":null}}`
	nsClientMock.On("Patch", fixNSName, types.StrategicMergePatchType, []byte(deletedLabelNS)).
		Return(&fixExpectedNS, nil).
		Once()

	svc := labeler.New(emInformer, nil, nsClientMock, nil, spy.NewLogDummy())

	awaitInformerStartAtMost(t, time.Second, emInformer)

	// when
	err := svc.DeleteAccessLabelFromNamespace(&fixNS)

	// then
	assert.NoError(t, err)
}

func TestControllerRunFailure(t *testing.T) {
	// given
	fixEM := fixEnvironmentMappingCR(fixREName, fixNSName)
	fixNS := fixNamespace(fixNSName)
	fixErr := errors.New("fix get err")
	fixPatchErr := errors.New("fix patch err")

	emInformer := fakeEMInformer(fixEM)

	expectations := &sync.WaitGroup{}
	expectations.Add(2)
	fulfillExpectation := func(mock.Arguments) {
		expectations.Done()
	}

	nsClientMock := &automock.NsPatcher{}
	defer nsClientMock.AssertExpectations(t)
	nsClientMock.On("Patch", fixNSName, types.StrategicMergePatchType, []byte("{}")).
		Return(nil, fixPatchErr).
		Run(fulfillExpectation).
		Once()

	reGetter := &automock.ReGetter{}
	defer reGetter.AssertExpectations(t)
	reGetter.On("Get", internal.RemoteEnvironmentName(fixREName)).
		Return(nil, fixErr).
		Run(fulfillExpectation).
		Once()

	svc := labeler.New(emInformer, nil, nsClientMock, reGetter, spy.NewLogDummy())

	awaitInformerStartAtMost(t, time.Second, emInformer)

	// when
	err2 := svc.DeleteAccessLabelFromNamespace(&fixNS)
	_, err3 := svc.GetAccessLabelFromRE(fixREName)

	// then
	awaitForSyncGroupAtMost(t, expectations, time.Second)
	assert.EqualError(t, err2, fmt.Sprintf("failed to delete AccessLabel from the namespace: %q, failed to patch namespace: %q: %v", fixNSName, fixNSName, fixPatchErr.Error()))
	assert.EqualError(t, err3, fmt.Sprintf("while getting remote environment with name: %q: %v", fixREName, fixErr.Error()))
}

func awaitForSyncGroupAtMost(t *testing.T, wg *sync.WaitGroup, timeout time.Duration) {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()

	select {
	case <-c:
	case <-time.After(timeout):
		t.Fatalf("timeout occurred when waiting for sync group")
	}
}

func awaitInformerStartAtMost(t *testing.T, timeout time.Duration, informer cache.SharedIndexInformer) {
	stop := make(chan struct{})
	syncedDone := make(chan struct{})

	go func() {
		if !cache.WaitForCacheSync(stop, informer.HasSynced) {
			t.Fatalf("timeout occurred when waiting to sync informer")
		}
		close(syncedDone)
	}()

	go informer.Run(stop)

	select {
	case <-time.After(timeout):
		close(stop)
	case <-syncedDone:
	}
}

func fixEnvironmentMappingCR(name, ns string) v1alpha1.EnvironmentMapping {
	return v1alpha1.EnvironmentMapping{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
	}
}

func fixRemoteEnvironment(fixREName string) internal.RemoteEnvironment {
	return internal.RemoteEnvironment{
		Name:        internal.RemoteEnvironmentName(fixREName),
		AccessLabel: "fix-access-1",
	}
}

func fixNamespace(fixNSName string) corev1.Namespace {
	return corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: fixNSName,
		},
	}
}
func fixNamespaceWithAccessLabel(fixNSName string) corev1.Namespace {
	return corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: fixNSName,
			Labels: map[string]string{
				"accessLabel": "fix-access-1",
			},
		},
	}
}
func fakeEMInformer(fixEM v1alpha1.EnvironmentMapping) cache.SharedIndexInformer {
	client := fake.NewSimpleClientset(&fixEM)
	informerFactory := externalversions.NewSharedInformerFactory(client, 0)
	remoteEnvironmentSharedInformers := informerFactory.Remoteenvironment().V1alpha1()
	emInformer := remoteEnvironmentSharedInformers.EnvironmentMappings().Informer()
	return emInformer
}

func fakeNSInformer(fixNS corev1.Namespace) cache.SharedIndexInformer {
	client := fake2.NewSimpleClientset(&fixNS)
	return v1.NewNamespaceInformer(client, 0, cache.Indexers{})
}
