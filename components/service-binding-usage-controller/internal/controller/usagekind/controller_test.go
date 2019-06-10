package usagekind_test

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/service-binding-usage-controller/internal/controller"
	metricAutomock "github.com/kyma-project/kyma/components/service-binding-usage-controller/internal/controller/automock"
	"github.com/kyma-project/kyma/components/service-binding-usage-controller/internal/controller/metric"
	"github.com/kyma-project/kyma/components/service-binding-usage-controller/internal/controller/usagekind"
	"github.com/kyma-project/kyma/components/service-binding-usage-controller/internal/controller/usagekind/automock"
	"github.com/kyma-project/kyma/components/service-binding-usage-controller/internal/platform/logger/spy"
	"github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned/fake"
	informers "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/informers/externalversions"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	dynamicFake "k8s.io/client-go/dynamic/fake"
)

func TestControllerRunAddAndDelete(t *testing.T) {
	// GIVEN
	registry := &automock.SupervisorRegistry{}
	metrics := &metricAutomock.BusinessMetric{}

	tc := newControllerTestCase(registry, metrics)
	defer tc.TearDown()

	// the goal of controller work is to register/unregister supervisor
	registry.On("Register", controller.Kind(fixUsageKind().Name), mock.MatchedBy(tc.SupervisorMatcher())).Return(nil).Once()
	metrics.ExpectOnIncrementQueueLength(metric.UkController).Once()
	metrics.ExpectOnDecrementQueueLength(metric.UkController).Once()
	metrics.ExpectOnRecordLatency(metric.UkController).Once()

	tc.RunController()

	// WHEN
	tc.AddUsageKind(t, fixUsageKind())

	// THEN
	tc.WaitForItemProcessed(t, 5*time.Second)
	registry.AssertExpectations(t)
	metrics.AssertExpectations(t)

	// GIVEN
	registry.On("Unregister", controller.Kind(fixUsageKind().Name)).Return(nil).Once()
	metrics.ExpectOnIncrementQueueLength(metric.UkController).Once()
	metrics.ExpectOnDecrementQueueLength(metric.UkController).Once()
	metrics.ExpectOnRecordLatency(metric.UkController).Once()

	// WHEN
	tc.RemoveUsageKind(t, fixUsageKind().Name)

	// THEN
	tc.WaitForItemProcessed(t, 5*time.Second)
	registry.AssertExpectations(t)
	metrics.AssertExpectations(t)
}

func fixUsageKind() *v1alpha1.UsageKind {
	return &v1alpha1.UsageKind{
		ObjectMeta: v1.ObjectMeta{
			Name: "deployment",
		},
		Spec: v1alpha1.UsageKindSpec{
			DisplayName: "Simple deployment",
			Resource: &v1alpha1.ResourceReference{
				Kind:    "deployment",
				Group:   "apps",
				Version: "v1",
			},
			LabelsPath: "spec.template.metadata.labels",
		},
	}
}

type controllerTestCase struct {
	*testCase

	asyncOpDone chan struct{}
}

func newControllerTestCase(registry *automock.SupervisorRegistry, metric *metricAutomock.BusinessMetric) *controllerTestCase {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	asyncOpDone := make(chan struct{})
	hookAsyncOp := func() {
		asyncOpDone <- struct{}{}
	}

	cs := fake.NewSimpleClientset()
	informersFactory := informers.NewSharedInformerFactory(cs, 0)
	kindInformer := informersFactory.Servicecatalog().V1alpha1().UsageKinds()
	cp := &dynamicFake.FakeDynamicClient{}

	ctr := usagekind.NewKindController(kindInformer, registry, cp, spy.NewLogSink().Logger, metric).WithTestHookOnAsyncOpDone(hookAsyncOp)

	informersFactory.Start(ctx.Done())
	return &controllerTestCase{
		testCase: &testCase{
			ctrl:           ctr,
			cancel:         cancel,
			timeoutContext: ctx,
			clientSet:      cs,
		},
		asyncOpDone: asyncOpDone,
	}
}

func (tc *controllerTestCase) SupervisorMatcher() interface{} {
	return func(s *controller.GenericSupervisor) bool {
		// todo (pluggable SBU cleanup): check generic supervisor
		return true
	}
}

func (tc *controllerTestCase) WaitForItemProcessed(t *testing.T, timeout time.Duration) {
	tc.WaitForChan(t, tc.asyncOpDone, timeout)
}
