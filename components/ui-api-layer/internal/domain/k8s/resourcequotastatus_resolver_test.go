package k8s

import (
	"context"
	"testing"

	"time"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/k8s/automock"
	testingUtils "github.com/kyma-project/kyma/components/ui-api-layer/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
)

func TestResourceQuotaStatusResolver_ResourceQuotaStatus_HappyPath(t *testing.T) {
	// GIVEN
	rqLister := automock.NewResourceQuotaLister()
	rqLister.On("ListResourceQuotas", fixNamespaceName()).Return(fixResourceQuota(), nil)
	rsLister := automock.NewReplicaSetLister()
	rsLister.On("ListReplicaSets", fixNamespaceName()).Return(fixReplicaSet(), nil)
	ssLister := automock.NewStatefulSetLister()
	ssLister.On("ListStatefulSets", fixNamespaceName()).Return(fixStatefulSet(), nil)
	podLister := automock.NewPodsLister()
	podLister.On("ListPods", fixNamespaceName(), fixStatefulSetMatchLabels()).Return(fixReplicas(fixStatefulSetMatchLabels()), nil)

	defer func() {
		rqLister.AssertExpectations(t)
		rsLister.AssertExpectations(t)
		ssLister.AssertExpectations(t)
		podLister.AssertExpectations(t)
	}()

	client := fake.NewSimpleClientset()
	informerFactory := informers.NewSharedInformerFactory(client, 0)
	informer := informerFactory.Apps().V1beta2().Deployments().Informer()
	deploySvc := newDeploymentService(informer)
	rqStatusSvc := newResourceQuotaStatusService(rqLister, rsLister, ssLister, podLister, deploySvc)

	testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

	resolver := newResourceQuotaStatusResolver(rqStatusSvc)

	// WHEN
	status, err := resolver.ResourceQuotasStatus(context.Background(), fixNamespaceName())
	require.NoError(t, err)

	assert.False(t, status.Exceeded)
}
