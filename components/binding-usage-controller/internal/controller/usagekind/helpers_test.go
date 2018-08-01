package usagekind_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/components/binding-usage-controller/internal/controller"
	"github.com/kyma-project/kyma/components/binding-usage-controller/internal/controller/usagekind"
	"github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/kyma-project/kyma/components/binding-usage-controller/pkg/client/clientset/versioned/fake"
)

type testCase struct {
	ctrl           runnable
	cancel         context.CancelFunc
	timeoutContext context.Context
	clientSet      *fake.Clientset
}

type runnable interface {
	Run(stopCh <-chan struct{})
}

func (tc *testCase) RunController() {
	go tc.ctrl.Run(tc.timeoutContext.Done())
}

func (tc *testCase) TearDown() {
	defer tc.cancel()
}

func (tc *testCase) AddSBU(t *testing.T, sbu *v1alpha1.ServiceBindingUsage) {
	_, err := tc.clientSet.Servicecatalog().ServiceBindingUsages(sbu.Namespace).Create(sbu)
	require.NoError(t, err)
}

func (tc *testCase) RemoveSBU(t *testing.T, sbu *v1alpha1.ServiceBindingUsage) {
	err := tc.clientSet.Servicecatalog().ServiceBindingUsages(sbu.Namespace).Delete(sbu.Name, &v1.DeleteOptions{})
	require.NoError(t, err)
	tc.ctrl.(*usagekind.ProtectionController).OnDeleteSBU(&controller.SBUDeletedEvent{
		Name:       sbu.Name,
		Namespace:  sbu.Namespace,
		UsedByKind: sbu.Spec.UsedBy.Kind,
	})
}

func (tc *testCase) AddUsageKind(t *testing.T, uk *v1alpha1.UsageKind) {
	_, err := tc.clientSet.Servicecatalog().UsageKinds().Create(uk)
	require.NoError(t, err)
}

func (tc *testCase) UpdateUsageKind(t *testing.T, uk *v1alpha1.UsageKind) {
	_, err := tc.clientSet.Servicecatalog().UsageKinds().Update(uk)
	require.NoError(t, err)
}

func (tc *testCase) GetUsageKind(t *testing.T, name string) *v1alpha1.UsageKind {
	obj, err := tc.clientSet.Servicecatalog().UsageKinds().Get(name, v1.GetOptions{})
	require.NoError(t, err)
	return obj
}

func (tc *testCase) RemoveUsageKind(t *testing.T, name string) {
	err := tc.clientSet.Servicecatalog().UsageKinds().Delete(name, &v1.DeleteOptions{})
	require.NoError(t, err)
}

func (tc *testCase) WaitForChan(t *testing.T, ch chan struct{}, timeout time.Duration) {
	select {
	case <-ch:
	case <-time.After(timeout):
		t.Fatalf("timeout occured when waiting for channel")
	}
}
