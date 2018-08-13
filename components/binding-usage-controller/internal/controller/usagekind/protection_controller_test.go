package usagekind_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/components/binding-usage-controller/internal/controller/usagekind"
	"github.com/kyma-project/kyma/components/binding-usage-controller/internal/platform/logger/spy"
	"github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/kyma-project/kyma/components/binding-usage-controller/pkg/client/clientset/versioned/fake"
	informers "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/client/informers/externalversions"
)

func TestProtectionControllerAddsFinalizer(t *testing.T) {
	// GIVEN
	tc := newProtectionControllerTestCase()
	defer tc.TearDown()
	tc.RunController()

	// WHEN
	tc.AddUsageKind(t, fixUsageKind())
	tc.WaitForAddFinalizer(t, 3*time.Second)

	// THEN
	uk := tc.GetUsageKind(t, fixUsageKind().Name)
	assert.Contains(t, uk.Finalizers, usagekind.FinalizerName)

	// WHEN
	tc.MarkRemoval(t, uk.Name)
	tc.WaitForProcessDeletionDone(t, 3*time.Second)

	// THEN
	uk = tc.GetUsageKind(t, fixUsageKind().Name)
	assert.NotContains(t, uk.Finalizers, usagekind.FinalizerName)
}

func TestProtectionControllerProtectsUsageKindUsedBySBU(t *testing.T) {
	// GIVEN
	tc := newProtectionControllerTestCase()
	defer tc.TearDown()
	tc.RunController()

	tc.AddSBU(t, fixSBU(fixUsageKind().Name))
	tc.AddUsageKind(t, fixUsageKind())
	tc.WaitForAddFinalizer(t, 3*time.Second)

	// WHEN
	tc.MarkRemoval(t, fixUsageKind().Name)
	tc.WaitForProcessDeletionDone(t, 3*time.Second)

	// THEN
	uk := tc.GetUsageKind(t, fixUsageKind().Name)
	assert.Contains(t, uk.Finalizers, usagekind.FinalizerName)

	// WHEN
	tc.RemoveSBU(t, fixSBU(fixUsageKind().Name))
	tc.WaitForProcessDeletionDone(t, 3*time.Second)

	// THEN

	// Caches may be not up to date and update of UsageKind triggers another processing, so just after removal the UsageKind could contain the finalizer.
	// After a short period of time the uk.Finalizers must be empty
	tc.WaitForTrue(t, 100*time.Millisecond, func() bool {
		uk := tc.GetUsageKind(t, fixUsageKind().Name)
		return len(uk.Finalizers) == 0
	}, "expected empty finalizer list")
}

func fixSBU(kind string) *v1alpha1.ServiceBindingUsage {
	return &v1alpha1.ServiceBindingUsage{
		ObjectMeta: metaV1.ObjectMeta{
			Namespace: "production",
			Name:      "sbu-",
			UID:       "uid-123",
		},
		Spec: v1alpha1.ServiceBindingUsageSpec{
			UsedBy: v1alpha1.LocalReferenceByKindAndName{
				Name: "redis-client",
				Kind: kind,
			},
			ServiceBindingRef: v1alpha1.LocalReferenceByName{
				Name: "redis-client",
			},
		},
	}
}

type protectionTestCase struct {
	*testCase

	addFinalizerOpDone    chan struct{}
	processDeletionOpDone chan struct{}
}

func newProtectionControllerTestCase() *protectionTestCase {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

	addFinalizerDone := make(chan struct{})
	addFinalizerOpHook := func() {
		addFinalizerDone <- struct{}{}
	}

	processDeletionDone := make(chan struct{})
	processDeletionOpHook := func() {
		processDeletionDone <- struct{}{}
	}

	cs := fake.NewSimpleClientset()
	informersFactory := informers.NewSharedInformerFactory(cs, 0)

	ctrl := usagekind.NewProtectionController(informersFactory.Servicecatalog().V1alpha1().UsageKinds(),
		informersFactory.Servicecatalog().V1alpha1().ServiceBindingUsages(),
		cs.ServicecatalogV1alpha1(),
		spy.NewLogSink().Logger).
		WithTestHookOnAsyncOpDone(addFinalizerOpHook, processDeletionOpHook)

	informersFactory.Start(ctx.Done())

	return &protectionTestCase{
		testCase: &testCase{
			ctrl:           ctrl,
			cancel:         cancel,
			timeoutContext: ctx,
			clientSet:      cs,
		},
		addFinalizerOpDone:    addFinalizerDone,
		processDeletionOpDone: processDeletionDone,
	}
}

func (tc *protectionTestCase) MarkRemoval(t *testing.T, name string) {
	uk := tc.GetUsageKind(t, name)
	copy := uk.DeepCopy()
	copy.DeletionTimestamp = &metaV1.Time{Time: time.Now()}
	tc.UpdateUsageKind(t, copy)
}

func (tc *protectionTestCase) WaitForAddFinalizer(t *testing.T, timeout time.Duration) {
	tc.WaitForChan(t, tc.addFinalizerOpDone, timeout)
}

func (tc *protectionTestCase) WaitForProcessDeletionDone(t *testing.T, timeout time.Duration) {
	tc.WaitForChan(t, tc.processDeletionOpDone, timeout)
}
