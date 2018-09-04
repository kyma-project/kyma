package k8s

import (
	"context"
	"testing"

	"time"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/k8s/automock"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	testingUtils "github.com/kyma-project/kyma/components/ui-api-layer/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
)

func TestResourceQuotaResolver_ResourceQuotasQuery(t *testing.T) {
	// GIVEN
	env := "production"
	lister := automock.NewResourceQuotaLister()
	lister.On("ListResourceQuotas", env).Return([]*v1.ResourceQuota{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "mem-default",
				Namespace: "production",
			},
		},
	}, nil)
	defer lister.AssertExpectations(t)

	resolver := newResourceQuotaResolver(lister, nil)

	// WHEN
	result, err := resolver.ResourceQuotasQuery(context.Background(), env)

	// THEN
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, gqlschema.ResourceQuota{Name: "mem-default"}, result[0])
}

func TestResourceQuotaResolver_ResourceQuotaStatus_HappyPath(t *testing.T) {
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

	resolver := newResourceQuotaResolver(rqLister, rqStatusSvc)

	// WHEN
	status, err := resolver.ResourceQuotaStatus(context.Background(), fixNamespaceName())
	require.NoError(t, err)

	assert.False(t, status.Exceeded)
}
