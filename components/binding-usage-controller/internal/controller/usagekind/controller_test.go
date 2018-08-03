package usagekind_test

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/binding-usage-controller/internal/controller"
	"github.com/kyma-project/kyma/components/binding-usage-controller/internal/controller/usagekind"
	"github.com/kyma-project/kyma/components/binding-usage-controller/internal/controller/usagekind/automock"
	"github.com/kyma-project/kyma/components/binding-usage-controller/internal/platform/logger/spy"
	"github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/kyma-project/kyma/components/binding-usage-controller/pkg/client/clientset/versioned/fake"
	informers "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/client/informers/externalversions"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	dynamicFake "k8s.io/client-go/dynamic/fake"
)

func TestControllerRunAddAndDelete(t *testing.T) {
	// GIVEN
	registry := &automock.SupervisorRegistry{}
	tc := newTestCase(registry)
	defer tc.TearDown()

	// the goal of controller work is to register/unregister supervisor
	registry.On("Register", controller.Kind(fixUsageKind().Name), mock.MatchedBy(tc.SupervisorMatcher())).Return(nil).Once()

	tc.RunController()

	// WHEN
	tc.AddUsageKind(t, fixUsageKind())

	// THEN
	tc.WaitForItemProcessed(t, 5*time.Second)
	registry.AssertExpectations(t)

	// GIVEN
	registry.On("Unregister", controller.Kind(fixUsageKind().Name)).Return(nil).Once()

	// WHEN
	tc.RemoveUsageKind(t, fixUsageKind().Name)

	// THEN
	tc.WaitForItemProcessed(t, 5*time.Second)
	registry.AssertExpectations(t)
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

type testCase struct {
	kindController *usagekind.Controller
	cancel         context.CancelFunc
	timeoutContext context.Context
	asyncOpDone    chan struct{}
	clientSet      *fake.Clientset
}

func newTestCase(registry *automock.SupervisorRegistry) *testCase {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	asyncOpDone := make(chan struct{})
	hookAsyncOp := func() {
		asyncOpDone <- struct{}{}
	}

	cs := fake.NewSimpleClientset()
	informersFactory := informers.NewSharedInformerFactory(cs, 0)
	kindInformer := informersFactory.Servicecatalog().V1alpha1().UsageKinds()
	cp := &dynamicFake.FakeClientPool{}

	ctr := usagekind.NewKindController(kindInformer, registry, cp, spy.NewLogSink().Logger).WithTestHookOnAsyncOpDone(hookAsyncOp)

	informersFactory.Start(ctx.Done())

	return &testCase{
		kindController: ctr,
		cancel:         cancel,
		timeoutContext: ctx,
		asyncOpDone:    asyncOpDone,
		clientSet:      cs,
	}
}

func (tc *testCase) SupervisorMatcher() interface{} {
	return func(s *controller.GenericSupervisor) bool {
		// todo (pluggable SBU cleanup): check generic supervisor
		return true
	}
}

func (tc *testCase) RunController() {
	go tc.kindController.Run(tc.timeoutContext.Done())
}

func (tc *testCase) TearDown() {
	defer tc.cancel()
}

func (tc *testCase) WaitForItemProcessed(t *testing.T, timeout time.Duration) {
	select {
	case <-tc.asyncOpDone:
	case <-time.After(timeout):
		t.Fatalf("timeout occured when waiting for channel")
	}
}

func (tc *testCase) RemoveUsageKind(t *testing.T, name string) {
	err := tc.clientSet.Servicecatalog().UsageKinds().Delete(name, &v1.DeleteOptions{})
	require.NoError(t, err)
}

func (tc *testCase) AddUsageKind(t *testing.T, uk *v1alpha1.UsageKind) {
	_, err := tc.clientSet.Servicecatalog().UsageKinds().Create(uk)
	require.NoError(t, err)
}
