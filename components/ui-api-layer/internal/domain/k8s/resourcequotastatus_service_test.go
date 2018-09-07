package k8s

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/k8s/automock"
	testingUtils "github.com/kyma-project/kyma/components/ui-api-layer/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	apps "k8s.io/api/apps/v1"
	api "k8s.io/api/apps/v1beta2"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
)

func TestResourceQuotaStatusService_CheckResourceQuotaStatus_HardEqualsUsed(t *testing.T) {
	// GIVEN
	rqLister := automock.NewResourceQuotaLister()
	rqLister.On("ListResourceQuotas", fixNamespaceName()).Return(fixResourceQuotaExceeded(), nil)
	defer rqLister.AssertExpectations(t)

	rqSvc := newResourceQuotaStatusService(rqLister, nil, nil, nil, nil)

	// WHEN
	status, err := rqSvc.CheckResourceQuotaStatus(fixNamespaceName(), fixResourceNames())
	require.NoError(t, err)

	// THEN
	assert.True(t, status.Exceeded)
	assert.Len(t, status.ExceededQuotas, 1)
	assert.Len(t, status.ExceededQuotas[0].ResourcesRequests, 2)
}

func TestResourceQuotaStatusService_CheckResourceQuotaStatus_ReplicaSetExceeded(t *testing.T) {
	// GIVEN
	rqLister := automock.NewResourceQuotaLister()
	rqLister.On("ListResourceQuotas", fixNamespaceName()).Return(fixResourceQuota(), nil)
	rsLister := automock.NewReplicaSetLister()
	rsLister.On("ListReplicaSets", fixNamespaceName()).Return(fixReplicaSetExceeded(), nil)
	podLister := automock.NewPodsLister()
	podLister.On("ListPods", fixNamespaceName(), fixReplicaSetMatchLabels()).Return(fixReplicasExceeding(fixReplicaSetMatchLabels()), nil)
	ssLister := automock.NewStatefulSetLister()
	ssLister.On("ListStatefulSets", fixNamespaceName()).Return([]*apps.StatefulSet{}, nil)
	defer func() {
		rqLister.AssertExpectations(t)
		rsLister.AssertExpectations(t)
		podLister.AssertExpectations(t)
		ssLister.AssertExpectations(t)
	}()

	rqSvc := newResourceQuotaStatusService(rqLister, rsLister, ssLister, podLister, nil)

	// WHEN
	status, err := rqSvc.CheckResourceQuotaStatus(fixNamespaceName(), fixResourceNames())
	require.NoError(t, err)

	// THEN
	assert.True(t, status.Exceeded)
	assert.Len(t, status.ExceededQuotas, 1)
	assert.Len(t, status.ExceededQuotas[0].ResourcesRequests, 2)
}

func TestResourceQuotaStatusService_CheckResourceQuotaStatus_ReplicaSetExceeded_MaxUnavailable(t *testing.T) {
	// GIVEN
	rqLister := automock.NewResourceQuotaLister()
	rqLister.On("ListResourceQuotas", fixNamespaceName()).Return(fixResourceQuota(), nil)
	rsLister := automock.NewReplicaSetLister()
	rsLister.On("ListReplicaSets", fixNamespaceName()).Return(fixReplicaSetWithOwnerReference(), nil)
	ssLister := automock.NewStatefulSetLister()
	ssLister.On("ListStatefulSets", fixNamespaceName()).Return([]*apps.StatefulSet{}, nil)
	defer func() {
		rqLister.AssertExpectations(t)
		rsLister.AssertExpectations(t)
		ssLister.AssertExpectations(t)
	}()

	client := fake.NewSimpleClientset(fixDeploy())
	informerFactory := informers.NewSharedInformerFactory(client, 0)
	informer := informerFactory.Apps().V1beta2().Deployments().Informer()
	deploySvc := newDeploymentService(informer)

	testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

	rqSvc := newResourceQuotaStatusService(rqLister, rsLister, ssLister, nil, deploySvc)

	// WHEN
	status, err := rqSvc.CheckResourceQuotaStatus(fixNamespaceName(), fixResourceNames())
	require.NoError(t, err)

	// THEN
	assert.False(t, status.Exceeded)
}

func TestResourceQuotaStatusService_CheckResourceQuotaStatus_ReplicaSetExceeded_MaxUnavailable_Percent(t *testing.T) {
	// GIVEN
	rqLister := automock.NewResourceQuotaLister()
	rqLister.On("ListResourceQuotas", fixNamespaceName()).Return(fixResourceQuota(), nil)
	rsLister := automock.NewReplicaSetLister()
	rsLister.On("ListReplicaSets", fixNamespaceName()).Return(fixReplicaSetWithOwnerReference(), nil)
	podLister := automock.NewPodsLister()
	podLister.On("ListPods", fixNamespaceName(), fixReplicaSetMatchLabels()).Return(fixReplicasExceeding(fixReplicaSetMatchLabels()), nil)
	ssLister := automock.NewStatefulSetLister()
	ssLister.On("ListStatefulSets", fixNamespaceName()).Return([]*apps.StatefulSet{}, nil)
	defer func() {
		rqLister.AssertExpectations(t)
		rsLister.AssertExpectations(t)
		podLister.AssertExpectations(t)
		ssLister.AssertExpectations(t)
	}()

	client := fake.NewSimpleClientset(fixDeployWithPercentageMaxUnavailable())
	informerFactory := informers.NewSharedInformerFactory(client, 0)
	informer := informerFactory.Apps().V1beta2().Deployments().Informer()
	deploySvc := newDeploymentService(informer)
	rqSvc := newResourceQuotaStatusService(rqLister, rsLister, ssLister, podLister, deploySvc)

	testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

	// WHEN
	status, err := rqSvc.CheckResourceQuotaStatus(fixNamespaceName(), fixResourceNames())
	require.NoError(t, err)

	// THEN
	assert.True(t, status.Exceeded)
	assert.Len(t, status.ExceededQuotas, 1)
	assert.Len(t, status.ExceededQuotas[0].ResourcesRequests, 2)
}

func TestResourceQuotaStatusService_CheckResourceQuotaStatus_StatefulSetExceed(t *testing.T) {
	// GIVEN
	rqLister := automock.NewResourceQuotaLister()
	rqLister.On("ListResourceQuotas", fixNamespaceName()).Return(fixResourceQuota(), nil)
	rsLister := automock.NewReplicaSetLister()
	rsLister.On("ListReplicaSets", fixNamespaceName()).Return([]*apps.ReplicaSet{}, nil)
	ssLister := automock.NewStatefulSetLister()
	ssLister.On("ListStatefulSets", fixNamespaceName()).Return(fixStatefulSet(), nil)
	podLister := automock.NewPodsLister()
	podLister.On("ListPods", fixNamespaceName(), fixStatefulSetMatchLabels()).Return(fixReplicasExceeding(fixStatefulSetMatchLabels()), nil)
	defer func() {
		rqLister.AssertExpectations(t)
		rsLister.AssertExpectations(t)
		podLister.AssertExpectations(t)
	}()

	rqSvc := newResourceQuotaStatusService(rqLister, rsLister, ssLister, podLister, nil)

	// WHEN
	status, err := rqSvc.CheckResourceQuotaStatus(fixNamespaceName(), fixResourceNames())
	require.NoError(t, err)

	// THEN
	assert.True(t, status.Exceeded)
	assert.Len(t, status.ExceededQuotas, 1)
	assert.Len(t, status.ExceededQuotas[0].ResourcesRequests, 2)
}

func TestResourceQuotaStatusService_CheckResourceQuotaStatus_ManyResourcesExceed(t *testing.T) {
	// GIVEN
	rqLister := automock.NewResourceQuotaLister()
	rqLister.On("ListResourceQuotas", fixNamespaceName()).Return(fixResourceQuota(), nil)
	rsLister := automock.NewReplicaSetLister()
	rsLister.On("ListReplicaSets", fixNamespaceName()).Return(fixReplicaSetExceeded(), nil)
	ssLister := automock.NewStatefulSetLister()
	ssLister.On("ListStatefulSets", fixNamespaceName()).Return(fixStatefulSet(), nil)
	podLister := automock.NewPodsLister()
	podLister.On("ListPods", fixNamespaceName(), fixReplicaSetMatchLabels()).Return(fixReplicasExceeding(fixReplicaSetMatchLabels()), nil).Run(func(args mock.Arguments) {
		podLister.On("ListPods", fixNamespaceName(), fixStatefulSetMatchLabels()).Return(fixReplicasExceeding(fixStatefulSetMatchLabels()), nil)
	})
	defer func() {
		rqLister.AssertExpectations(t)
		rsLister.AssertExpectations(t)
		podLister.AssertExpectations(t)
	}()

	rqSvc := newResourceQuotaStatusService(rqLister, rsLister, ssLister, podLister, nil)

	// WHEN
	status, err := rqSvc.CheckResourceQuotaStatus(fixNamespaceName(), fixResourceNames())
	require.NoError(t, err)

	// THEN
	assert.True(t, status.Exceeded)
	assert.Len(t, status.ExceededQuotas, 1)
	assert.Len(t, status.ExceededQuotas[0].ResourcesRequests, 2)
	assert.Len(t, status.ExceededQuotas[0].ResourcesRequests[0].DemandingResources, 2)
	assert.Len(t, status.ExceededQuotas[0].ResourcesRequests[1].DemandingResources, 2)
}

func fixResourceNames() []v1.ResourceName {
	return []v1.ResourceName{
		v1.ResourceRequestsMemory,
		v1.ResourceLimitsMemory,
		v1.ResourceRequestsCPU,
		v1.ResourceLimitsCPU,
		v1.ResourcePods,
	}
}

func fixNamespaceName() string {
	return "test-ns"
}

func fixName() string {
	return "fix-name"
}

func fixResourceQuotaExceeded() []*v1.ResourceQuota {
	return []*v1.ResourceQuota{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fixName(),
				Namespace: fixNamespaceName(),
			},
			Spec: v1.ResourceQuotaSpec{
				Hard: v1.ResourceList{
					v1.ResourceLimitsMemory:   resource.MustParse("1Gi"),
					v1.ResourceRequestsMemory: resource.MustParse("1Gi"),
				},
			},
			Status: v1.ResourceQuotaStatus{
				Used: v1.ResourceList{
					v1.ResourceLimitsMemory:   resource.MustParse("1Gi"),
					v1.ResourceRequestsMemory: resource.MustParse("1Gi"),
				},
			},
		},
	}
}

func fixResourceQuota() []*v1.ResourceQuota {
	return []*v1.ResourceQuota{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fixName(),
				Namespace: fixNamespaceName(),
			},
			Spec: v1.ResourceQuotaSpec{
				Hard: v1.ResourceList{
					v1.ResourceLimitsMemory:   resource.MustParse("1Gi"),
					v1.ResourceRequestsMemory: resource.MustParse("1Gi"),
				},
			},
		},
	}
}

func fixStatefulSet() []*apps.StatefulSet {
	return []*apps.StatefulSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fixName(),
				Namespace: fixNamespaceName(),
			},
			Spec: apps.StatefulSetSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: fixStatefulSetMatchLabels(),
				},
				Replicas: ptrInt32(3),
			},
			Status: apps.StatefulSetStatus{
				Replicas: 2,
			},
		},
	}
}

func fixReplicaSet() []*apps.ReplicaSet {
	return []*apps.ReplicaSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: fixNamespaceName(),
			},
			Spec: apps.ReplicaSetSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: fixReplicaSetMatchLabels(),
				},
				Replicas: ptrInt32(2),
			},
			Status: apps.ReplicaSetStatus{
				Replicas: 2,
			},
		},
	}
}

func fixReplicaSetExceeded() []*apps.ReplicaSet {
	return []*apps.ReplicaSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fixName(),
				Namespace: fixNamespaceName(),
			},
			Spec: apps.ReplicaSetSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: fixReplicaSetMatchLabels(),
				},
				Replicas: ptrInt32(3),
			},
			Status: apps.ReplicaSetStatus{
				Replicas: 2,
			},
		},
	}
}

func fixReplicaSetWithOwnerReference() []*apps.ReplicaSet {
	return []*apps.ReplicaSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fixName(),
				Namespace: fixNamespaceName(),
				OwnerReferences: []metav1.OwnerReference{
					{
						Name: fixDeployName(),
						Kind: "Deployment",
						UID:  "123",
					},
				},
			},
			Spec: apps.ReplicaSetSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: fixReplicaSetMatchLabels(),
				},
				Replicas: ptrInt32(3),
			},
			Status: apps.ReplicaSetStatus{
				Replicas: 2,
			},
		},
	}
}

func fixReplicaSetMatchLabels() map[string]string {
	return map[string]string{
		"label": "replica",
	}
}

func fixStatefulSetMatchLabels() map[string]string {
	return map[string]string{
		"label": "stateful",
	}
}

func fixReplicasExceeding(labels map[string]string) []v1.Pod {
	return []v1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Resources: v1.ResourceRequirements{
							Limits: v1.ResourceList{
								v1.ResourceMemory: resource.MustParse("2Gi"),
							},
							Requests: v1.ResourceList{
								v1.ResourceMemory: resource.MustParse("2Gi"),
							},
						},
					},
				},
			},
		},
	}
}

func fixReplicas(labels map[string]string) []v1.Pod {
	return []v1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Resources: v1.ResourceRequirements{
							Limits: v1.ResourceList{
								v1.ResourceMemory: resource.MustParse("100Mi"),
							},
							Requests: v1.ResourceList{
								v1.ResourceMemory: resource.MustParse("100Mi"),
							},
						},
					},
				},
			},
		},
	}
}

func fixDeploy() *api.Deployment {
	return &api.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fixDeployName(),
			Namespace: fixNamespaceName(),
		},
		Spec: api.DeploymentSpec{
			Strategy: api.DeploymentStrategy{
				RollingUpdate: &api.RollingUpdateDeployment{
					MaxUnavailable: &intstr.IntOrString{IntVal: 1, StrVal: "1"},
				},
			},
		},
	}
}

func fixDeployWithPercentageMaxUnavailable() *api.Deployment {
	return &api.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fixDeployName(),
			Namespace: fixNamespaceName(),
		},
		Spec: api.DeploymentSpec{
			Strategy: api.DeploymentStrategy{
				RollingUpdate: &api.RollingUpdateDeployment{
					MaxUnavailable: &intstr.IntOrString{StrVal: "20%"},
				},
			},
		},
	}
}

func fixDeployName() string {
	return "deploy-name"
}

func ptrInt32(int int32) *int32 {
	return &int
}
