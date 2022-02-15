package mapping_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	sc_fake "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/fake"
	externalversions2 "github.com/kubernetes-sigs/service-catalog/pkg/client/informers_generated/externalversions"
	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/internal/broker"
	"github.com/kyma-project/kyma/components/application-broker/internal/mapping"
	"github.com/kyma-project/kyma/components/application-broker/internal/mapping/automock"
	"github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/fake"
	"github.com/kyma-project/kyma/components/application-broker/pkg/client/informers/externalversions"
	"github.com/kyma-project/kyma/components/application-broker/platform/logger/spy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/informers/core/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
)

const (
	fixNSName  = "production"
	fixAPPName = "ec-prod"
)

func TestControllerRunSuccess(t *testing.T) {
	// given
	fixEM := fixApplicationMappingCR(fixAPPName, fixNSName)
	fixLivenessCheckStatus := broker.LivenessCheckStatus{Succeeded: false}

	expectations := &sync.WaitGroup{}
	expectations.Add(2)
	fulfillExpectation := func(mock.Arguments) {
		expectations.Done()
	}

	emInformer := newEmInformerFromFakeClientset(fixEM)
	instInformer := newInstanceInformerFromFakeClientset()

	brokerFacade := &automock.NsBrokerFacade{}
	defer brokerFacade.AssertExpectations(t)
	brokerFacade.On("Exist", fixNSName).Return(false, nil).
		Run(fulfillExpectation)
	brokerFacade.On("Create", fixNSName).Return(nil).
		Run(fulfillExpectation)

	svc := mapping.New(emInformer, instInformer, brokerFacade, nil, fakeInstanceChecker(false), spy.NewLogDummy(), &fixLivenessCheckStatus)
	awaitInformerStartAtMost(t, time.Second, emInformer)

	ctx, cancel := context.WithCancel(context.Background())

	// when
	go svc.Run(ctx.Done())

	// then
	awaitForSyncGroupAtMost(t, expectations, time.Second)

	// clean-up - release controller
	cancel()
}

type tcNsBrokersEnabled struct {
	name                  string
	prepareNsBrokerFacade func() *automock.NsBrokerFacade
	prepareMappingSvc     func() *automock.MappingLister
	prepareNsBrokerSyncer func() *automock.NsBrokerSyncer
	errorMsg              string
}

func TestControllerProcessItemOnEMCreationWhenNsBrokersEnabled(t *testing.T) {
	for _, tc := range []tcNsBrokersEnabled{
		{
			name: "happy path",
			prepareNsBrokerFacade: func() *automock.NsBrokerFacade {
				nsBrokerFacade := &automock.NsBrokerFacade{}
				nsBrokerFacade.On("Exist", fixNSName).Return(false, nil)
				nsBrokerFacade.On("Create", fixNSName).Return(nil)
				return nsBrokerFacade
			},
			prepareNsBrokerSyncer: func() *automock.NsBrokerSyncer {
				return &automock.NsBrokerSyncer{}
			},
		},
		{
			name: "error on checking if namespaced broker exist",
			prepareNsBrokerFacade: func() *automock.NsBrokerFacade {
				nsBrokerFacade := &automock.NsBrokerFacade{}
				nsBrokerFacade.On("Exist", fixNSName).Return(false, errors.New("some error"))
				return nsBrokerFacade
			},
			prepareNsBrokerSyncer: func() *automock.NsBrokerSyncer {
				return &automock.NsBrokerSyncer{}
			},
			errorMsg: "while checking if namespaced broker exist in namespace [production]: some error",
		},
		{
			name: "error on creation of namespaced broker",
			prepareNsBrokerFacade: func() *automock.NsBrokerFacade {
				nsBrokerFacade := &automock.NsBrokerFacade{}
				nsBrokerFacade.On("Exist", fixNSName).Return(false, nil)
				nsBrokerFacade.On("Create", fixNSName).Return(errors.New("some error"))
				return nsBrokerFacade
			},
			prepareNsBrokerSyncer: func() *automock.NsBrokerSyncer {
				return &automock.NsBrokerSyncer{}
			},
			errorMsg: "while creating namespaced broker in namespace [production]: some error",
		},
		{
			name: "sync namespaced broker if it already exist",
			prepareNsBrokerFacade: func() *automock.NsBrokerFacade {
				nsBrokerFacade := &automock.NsBrokerFacade{}
				nsBrokerFacade.On("Exist", fixNSName).Return(true, nil)
				return nsBrokerFacade
			},
			prepareNsBrokerSyncer: func() *automock.NsBrokerSyncer {
				nsBrokerSyncer := &automock.NsBrokerSyncer{}
				nsBrokerSyncer.On("SyncBroker", fixNSName).Return(nil)
				return nsBrokerSyncer
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// GIVEN
			fixEM := fixApplicationMappingCR(fixAPPName, fixNSName)
			fixLivenessCheckStatus := broker.LivenessCheckStatus{Succeeded: false}

			emInformer := newEmInformerFromFakeClientset(fixEM)
			instInformer := newInstanceInformerFromFakeClientset()

			awaitInformerStartAtMost(t, time.Second, emInformer)

			nsBrokerFacade := tc.prepareNsBrokerFacade()
			defer nsBrokerFacade.AssertExpectations(t)

			nsBrokerSyncer := tc.prepareNsBrokerSyncer()
			defer nsBrokerSyncer.AssertExpectations(t)

			svc := mapping.New(emInformer, instInformer, nsBrokerFacade, nsBrokerSyncer, fakeInstanceChecker(false), spy.NewLogDummy(), &fixLivenessCheckStatus)

			err := svc.ProcessItem(fmt.Sprintf("%s/%s", fixNSName, fixAPPName))
			if tc.errorMsg == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.errorMsg)
			}
		})
	}
}

func TestControllerProcessItemOnEMDeletionWhenNsBrokersEnabled(t *testing.T) {
	for _, tc := range []tcNsBrokersEnabled{
		{
			name: "broker will be removed if there is no application mappings in the namespace",
			prepareNsBrokerFacade: func() *automock.NsBrokerFacade {
				nsBrokerFacade := &automock.NsBrokerFacade{}
				nsBrokerFacade.On("Exist", fixNSName).Return(true, nil)
				nsBrokerFacade.On("Delete", fixNSName).Return(nil)
				return nsBrokerFacade
			},
			prepareMappingSvc: func() *automock.MappingLister {
				mappingSvc := &automock.MappingLister{}
				mappingSvc.On("ListApplicationMappings", fixNSName).Return([]*v1alpha1.ApplicationMapping{}, nil)
				return mappingSvc
			},
			prepareNsBrokerSyncer: func() *automock.NsBrokerSyncer {
				return &automock.NsBrokerSyncer{}
			},
		},
		{
			name: "error on checking if namespaced broker exist",
			prepareNsBrokerFacade: func() *automock.NsBrokerFacade {
				nsBrokerFacade := &automock.NsBrokerFacade{}
				nsBrokerFacade.On("Exist", fixNSName).Return(false, errors.New("some error"))
				return nsBrokerFacade
			},
			prepareMappingSvc: func() *automock.MappingLister {
				return &automock.MappingLister{}
			},
			prepareNsBrokerSyncer: func() *automock.NsBrokerSyncer {
				return &automock.NsBrokerSyncer{}
			},
			errorMsg: "while checking if namespaced broker exist in namespace [production]: some error",
		},
		{
			name: "error on removing namespaced broker",
			prepareNsBrokerFacade: func() *automock.NsBrokerFacade {
				nsBrokerFacade := &automock.NsBrokerFacade{}
				nsBrokerFacade.On("Exist", fixNSName).Return(true, nil)
				nsBrokerFacade.On("Delete", fixNSName).Return(errors.New("some error"))
				return nsBrokerFacade
			},
			prepareMappingSvc: func() *automock.MappingLister {
				mappingSvc := &automock.MappingLister{}
				mappingSvc.On("ListApplicationMappings", fixNSName).Return([]*v1alpha1.ApplicationMapping{}, nil)
				return mappingSvc
			},
			prepareNsBrokerSyncer: func() *automock.NsBrokerSyncer {
				return &automock.NsBrokerSyncer{}
			},
			errorMsg: "while removing namespaced broker from namespace [production]: some error",
		},
		{
			name: "namespace broker already removed",
			prepareNsBrokerFacade: func() *automock.NsBrokerFacade {
				nsBrokerFacade := &automock.NsBrokerFacade{}
				nsBrokerFacade.On("Exist", fixNSName).Return(false, nil)
				return nsBrokerFacade
			},
			prepareMappingSvc: func() *automock.MappingLister {
				return &automock.MappingLister{}
			},
			prepareNsBrokerSyncer: func() *automock.NsBrokerSyncer {
				return &automock.NsBrokerSyncer{}
			},
		},
		{
			name: "sync broker if at least 1 mapping exists instead of deleting it",
			prepareNsBrokerFacade: func() *automock.NsBrokerFacade {
				nsBrokerFacade := &automock.NsBrokerFacade{}
				nsBrokerFacade.On("Exist", fixNSName).Return(true, nil)
				return nsBrokerFacade
			},
			prepareMappingSvc: func() *automock.MappingLister {
				mappingSvc := &automock.MappingLister{}
				mappingSvc.On("ListApplicationMappings", fixNSName).Return([]*v1alpha1.ApplicationMapping{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "fix",
						},
					},
				}, nil)
				return mappingSvc
			},
			prepareNsBrokerSyncer: func() *automock.NsBrokerSyncer {
				nsBrokerSyncer := &automock.NsBrokerSyncer{}
				nsBrokerSyncer.On("SyncBroker", fixNSName).Return(nil)
				return nsBrokerSyncer
			},
		},
		{
			name: "error while listing application mappings",
			prepareNsBrokerFacade: func() *automock.NsBrokerFacade {
				nsBrokerFacade := &automock.NsBrokerFacade{}
				nsBrokerFacade.On("Exist", fixNSName).Return(true, nil)
				return nsBrokerFacade
			},
			prepareMappingSvc: func() *automock.MappingLister {
				mappingSvc := &automock.MappingLister{}
				mappingSvc.On("ListApplicationMappings", fixNSName).Return(nil, errors.New("some error"))
				return mappingSvc
			},
			prepareNsBrokerSyncer: func() *automock.NsBrokerSyncer {
				return &automock.NsBrokerSyncer{}
			},
			errorMsg: fmt.Sprintf("while listing application mappings from namespace [%s]: some error", fixNSName),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// GIVEN
			fixLivenessCheckStatus := broker.LivenessCheckStatus{Succeeded: false}

			emInformer := newEmInformerFromFakeClientset(nil)
			instInformer := newInstanceInformerFromFakeClientset()

			awaitInformerStartAtMost(t, time.Second, emInformer)

			nsBrokerFacade := tc.prepareNsBrokerFacade()
			defer nsBrokerFacade.AssertExpectations(t)

			mappingSvc := tc.prepareMappingSvc()
			defer mappingSvc.AssertExpectations(t)

			nsBrokerSyncer := tc.prepareNsBrokerSyncer()
			defer nsBrokerSyncer.AssertExpectations(t)

			svc := mapping.New(emInformer, instInformer, nsBrokerFacade, nsBrokerSyncer, fakeInstanceChecker(false), spy.NewLogDummy(), &fixLivenessCheckStatus).
				WithMappingLister(mappingSvc)
			// WHEN
			err := svc.ProcessItem(fmt.Sprintf("%s/%s", fixNSName, fixAPPName))
			// THEN
			if tc.errorMsg == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.errorMsg)
			}
		})
	}

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

func fixApplicationMappingCR(name, ns string) *v1alpha1.ApplicationMapping {
	return &v1alpha1.ApplicationMapping{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		},
	}
}

func fixApplication(fixAPPName string) internal.Application {
	return internal.Application{
		Name:        internal.ApplicationName(fixAPPName),
		AccessLabel: "fix-access-1",
	}
}

func fixNamespace(fixNSName string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: fixNSName,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
	}
}

func fixNamespaceWithAccessLabel(fixNSName string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: fixNSName,
			Labels: map[string]string{
				"accessLabel": "fix-access-1",
			},
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
	}
}

func newEmInformerFromFakeClientset(fixEM *v1alpha1.ApplicationMapping) cache.SharedIndexInformer {
	var client *fake.Clientset
	if fixEM != nil {
		client = fake.NewSimpleClientset(fixEM)
	} else {
		client = fake.NewSimpleClientset()
	}
	informerFactory := externalversions.NewSharedInformerFactory(client, 0)
	applicationSharedInformers := informerFactory.Applicationconnector().V1alpha1()
	emInformer := applicationSharedInformers.ApplicationMappings().Informer()
	return emInformer
}

func newNsInformerFromFakeClientset(fixNs *corev1.Namespace) cache.SharedIndexInformer {
	var client *k8sfake.Clientset
	if fixNs != nil {
		client = k8sfake.NewSimpleClientset(fixNs)
	} else {
		client = k8sfake.NewSimpleClientset()
	}
	i := v1.NewNamespaceInformer(client, 0, cache.Indexers{})
	return i
}

type instanceChecker interface {
	AnyServiceInstanceExists(namespace string) (bool, error)
}

type fakeInstChecker struct {
	val bool
}

func (f *fakeInstChecker) AnyServiceInstanceExists(ns string) (bool, error) {
	return f.val, nil
}

func fakeInstanceChecker(val bool) instanceChecker {
	return &fakeInstChecker{
		val: val,
	}
}
func newInstanceInformerFromFakeClientset() cache.SharedIndexInformer {
	client := sc_fake.NewSimpleClientset()
	informerFactory := externalversions2.NewSharedInformerFactory(client, 0)
	return informerFactory.Servicecatalog().V1beta1().ServiceInstances().Informer()
}
