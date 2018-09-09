package k8s

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/k8s/automock"
	testingUtils "github.com/kyma-project/kyma/components/ui-api-layer/internal/testing"
	"github.com/stretchr/testify/assert"
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

func TestResourceQuotaStatusService_CheckResourceQuotaStatus_ReplicaSetExceeded(t *testing.T) {
	// GIVEN
	rqLister := automock.NewResourceQuotaLister()
	rqLister.On("ListResourceQuotas", fixNamespaceName()).Return(fixResourceQuota(), nil)
	rsLister := automock.NewReplicaSetLister()
	rsLister.On("ListReplicaSets", fixNamespaceName()).Return(fixReplicaSetExceeded(), nil)
	ssLister := automock.NewStatefulSetLister()
	ssLister.On("ListStatefulSets", fixNamespaceName()).Return([]*apps.StatefulSet{}, nil)
	lrLister := automock.NewLimitRangeLister()
	lrLister.On("List", fixNamespaceName()).Return(fixLimitRangeList(), nil)
	defer func() {
		rqLister.AssertExpectations(t)
		rsLister.AssertExpectations(t)
		ssLister.AssertExpectations(t)
		lrLister.AssertExpectations(t)
	}()

	rqSvc := newResourceQuotaStatusService(rqLister, rsLister, ssLister, lrLister, nil)

	// WHEN
	status, err := rqSvc.CheckResourceQuotaStatus(fixNamespaceName(), fixResourceNames())
	require.NoError(t, err)

	// THEN
	assert.True(t, status.Exceeded)
	assert.Len(t, status.ExceededQuotas, 1)
	assert.Len(t, status.ExceededQuotas[0].AffectedResources, 1)
	assert.Equal(t, fixName(), status.ExceededQuotas[0].QuotaName)
	assert.Equal(t, v1.ResourceLimitsMemory, v1.ResourceName(status.ExceededQuotas[0].ResourceName))
}

func TestResourceQuotaStatusService_CheckResourceQuotaStatus_ReplicaSetExceeded_MaxUnavailable(t *testing.T) {
	// GIVEN
	rqLister := automock.NewResourceQuotaLister()
	rqLister.On("ListResourceQuotas", fixNamespaceName()).Return(fixResourceQuota(), nil)
	rsLister := automock.NewReplicaSetLister()
	rsLister.On("ListReplicaSets", fixNamespaceName()).Return(fixReplicaSetExceededWithOwnerReference(), nil)
	ssLister := automock.NewStatefulSetLister()
	ssLister.On("ListStatefulSets", fixNamespaceName()).Return([]*apps.StatefulSet{}, nil)
	lrLister := automock.NewLimitRangeLister()
	lrLister.On("List", fixNamespaceName()).Return(fixLimitRangeList(), nil)
	defer func() {
		rqLister.AssertExpectations(t)
		rsLister.AssertExpectations(t)
		ssLister.AssertExpectations(t)
		lrLister.AssertExpectations(t)
	}()

	client := fake.NewSimpleClientset(fixDeploy())
	informerFactory := informers.NewSharedInformerFactory(client, 0)
	informer := informerFactory.Apps().V1beta2().Deployments().Informer()
	deploySvc := newDeploymentService(informer)

	testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

	rqSvc := newResourceQuotaStatusService(rqLister, rsLister, ssLister, lrLister, deploySvc)

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
	rsLister.On("ListReplicaSets", fixNamespaceName()).Return(fixReplicaSetExceededWithOwnerReference(), nil)
	ssLister := automock.NewStatefulSetLister()
	ssLister.On("ListStatefulSets", fixNamespaceName()).Return([]*apps.StatefulSet{}, nil)
	lrLister := automock.NewLimitRangeLister()
	lrLister.On("List", fixNamespaceName()).Return(fixLimitRangeList(), nil)
	defer func() {
		rqLister.AssertExpectations(t)
		rsLister.AssertExpectations(t)
		ssLister.AssertExpectations(t)
		lrLister.AssertExpectations(t)
	}()

	client := fake.NewSimpleClientset(fixDeployWithPercentageMaxUnavailable())
	informerFactory := informers.NewSharedInformerFactory(client, 0)
	informer := informerFactory.Apps().V1beta2().Deployments().Informer()
	deploySvc := newDeploymentService(informer)
	rqSvc := newResourceQuotaStatusService(rqLister, rsLister, ssLister, lrLister, deploySvc)

	testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

	// WHEN
	status, err := rqSvc.CheckResourceQuotaStatus(fixNamespaceName(), fixResourceNames())
	require.NoError(t, err)

	// THEN
	assert.True(t, status.Exceeded)
	assert.Len(t, status.ExceededQuotas, 1)
	assert.Len(t, status.ExceededQuotas[0].AffectedResources, 1)
}

func TestResourceQuotaStatusService_CheckResourceQuotaStatus_StatefulSetExceed(t *testing.T) {
	// GIVEN
	rqLister := automock.NewResourceQuotaLister()
	rqLister.On("ListResourceQuotas", fixNamespaceName()).Return(fixResourceQuota(), nil)
	rsLister := automock.NewReplicaSetLister()
	rsLister.On("ListReplicaSets", fixNamespaceName()).Return([]*apps.ReplicaSet{}, nil)
	ssLister := automock.NewStatefulSetLister()
	ssLister.On("ListStatefulSets", fixNamespaceName()).Return(fixStatefulSetExceeded(), nil)
	lrLister := automock.NewLimitRangeLister()
	lrLister.On("List", fixNamespaceName()).Return(fixLimitRangeList(), nil)
	defer func() {
		rqLister.AssertExpectations(t)
		rsLister.AssertExpectations(t)
		ssLister.AssertExpectations(t)
		lrLister.AssertExpectations(t)
	}()

	rqSvc := newResourceQuotaStatusService(rqLister, rsLister, ssLister, lrLister, nil)

	// WHEN
	status, err := rqSvc.CheckResourceQuotaStatus(fixNamespaceName(), fixResourceNames())
	require.NoError(t, err)

	// THEN
	assert.True(t, status.Exceeded)
	assert.Len(t, status.ExceededQuotas, 1)
	assert.Len(t, status.ExceededQuotas[0].AffectedResources, 1)
}

func TestResourceQuotaStatusService_CheckResourceQuotaStatus_ManyResourcesExceed(t *testing.T) {
	// GIVEN
	rqLister := automock.NewResourceQuotaLister()
	rqLister.On("ListResourceQuotas", fixNamespaceName()).Return(fixResourceQuota(), nil)
	rsLister := automock.NewReplicaSetLister()
	rsLister.On("ListReplicaSets", fixNamespaceName()).Return(fixReplicaSetExceeded(), nil)
	ssLister := automock.NewStatefulSetLister()
	ssLister.On("ListStatefulSets", fixNamespaceName()).Return(fixStatefulSetExceeded(), nil)
	lrLister := automock.NewLimitRangeLister()
	lrLister.On("List", fixNamespaceName()).Return(fixLimitRangeList(), nil)
	defer func() {
		rqLister.AssertExpectations(t)
		rsLister.AssertExpectations(t)
		ssLister.AssertExpectations(t)
		lrLister.AssertExpectations(t)
	}()

	rqSvc := newResourceQuotaStatusService(rqLister, rsLister, ssLister, lrLister, nil)

	// WHEN
	status, err := rqSvc.CheckResourceQuotaStatus(fixNamespaceName(), fixResourceNames())
	require.NoError(t, err)

	// THEN
	assert.True(t, status.Exceeded)
	assert.Len(t, status.ExceededQuotas, 1)
	assert.Len(t, status.ExceededQuotas[0].AffectedResources, 2)
}

func TestResourceQuotaStatusService_CheckResourceQuotaStatus_LimitRangeExceedQuota(t *testing.T) {
	// GIVEN
	rqLister := automock.NewResourceQuotaLister()
	rqLister.On("ListResourceQuotas", fixNamespaceName()).Return(fixResourceQuota(), nil)
	rsLister := automock.NewReplicaSetLister()
	rsLister.On("ListReplicaSets", fixNamespaceName()).Return(fixReplicaSetExceededWithNoLimits(), nil)
	ssLister := automock.NewStatefulSetLister()
	ssLister.On("ListStatefulSets", fixNamespaceName()).Return(fixStatefulSet(), nil)
	lrLister := automock.NewLimitRangeLister()
	lrLister.On("List", fixNamespaceName()).Return(fixLimitRangeExceedingList(), nil)
	defer func() {
		rqLister.AssertExpectations(t)
		rsLister.AssertExpectations(t)
		ssLister.AssertExpectations(t)
		lrLister.AssertExpectations(t)
	}()

	rqSvc := newResourceQuotaStatusService(rqLister, rsLister, ssLister, lrLister, nil)

	// WHEN
	status, err := rqSvc.CheckResourceQuotaStatus(fixNamespaceName(), fixResourceNames())
	require.NoError(t, err)

	// THEN
	assert.True(t, status.Exceeded)
	assert.Len(t, status.ExceededQuotas, 2)
	assert.Len(t, status.ExceededQuotas[0].AffectedResources, 2)
	assert.Len(t, status.ExceededQuotas[1].AffectedResources, 2)
}

func TestResourceQuotaStatusService_CheckResourceQuotaStatus_MultiContainers(t *testing.T) {
	// GIVEN
	rqLister := automock.NewResourceQuotaLister()
	rqLister.On("ListResourceQuotas", fixNamespaceName()).Return(fixResourceQuota(), nil)
	rsLister := automock.NewReplicaSetLister()
	rsLister.On("ListReplicaSets", fixNamespaceName()).Return(fixReplicaSetMultiContainer(), nil)
	ssLister := automock.NewStatefulSetLister()
	ssLister.On("ListStatefulSets", fixNamespaceName()).Return([]*apps.StatefulSet{}, nil)
	lrLister := automock.NewLimitRangeLister()
	lrLister.On("List", fixNamespaceName()).Return(fixLimitRangeList(), nil)
	defer func() {
		rqLister.AssertExpectations(t)
		rsLister.AssertExpectations(t)
		ssLister.AssertExpectations(t)
		lrLister.AssertExpectations(t)
	}()

	rqSvc := newResourceQuotaStatusService(rqLister, rsLister, ssLister, lrLister, nil)

	// WHEN
	status, err := rqSvc.CheckResourceQuotaStatus(fixNamespaceName(), fixResourceNames())
	require.NoError(t, err)

	// THEN
	assert.True(t, status.Exceeded)
	assert.Len(t, status.ExceededQuotas, 1)
	assert.Len(t, status.ExceededQuotas[0].AffectedResources, 1)
}

func fixResourceNames() []v1.ResourceName {
	return []v1.ResourceName{
		v1.ResourceRequestsMemory,
		v1.ResourceLimitsMemory,
		v1.ResourceRequestsCPU,
		v1.ResourceLimitsCPU,
	}
}

func fixNamespaceName() string {
	return "test-ns"
}

func fixName() string {
	return "fix-name"
}

func fixLimitRangeList() []*v1.LimitRange {
	return []*v1.LimitRange{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fixName(),
				Namespace: fixNamespaceName(),
			},
			Spec: v1.LimitRangeSpec{
				Limits: []v1.LimitRangeItem{
					{
						Type: v1.LimitTypeContainer,
						DefaultRequest: v1.ResourceList{
							v1.ResourceMemory: resource.MustParse("500Mi"),
							v1.ResourceCPU:    resource.MustParse("500Mi"),
						},
					},
				},
			},
		},
	}
}

func fixReplicaSetMultiContainer() []*apps.ReplicaSet {
	return []*apps.ReplicaSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: fixNamespaceName(),
			},
			Spec: apps.ReplicaSetSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: fixReplicaSetMatchLabels(),
				},
				Replicas: ptrInt32(4),
				Template: v1.PodTemplateSpec{
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{Name: "test1"},
							{Name: "test2"},
							{Name: "test3"},
						},
					},
				},
			},
			Status: apps.ReplicaSetStatus{
				Replicas: 2,
			},
		},
	}
}

func fixLimitRangeExceedingList() []*v1.LimitRange {
	return []*v1.LimitRange{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fixName(),
				Namespace: fixNamespaceName(),
			},
			Spec: v1.LimitRangeSpec{
				Limits: []v1.LimitRangeItem{
					{
						Type: v1.LimitTypeContainer,
						DefaultRequest: v1.ResourceList{
							v1.ResourceMemory: resource.MustParse("2Gi"),
							v1.ResourceCPU:    resource.MustParse("2Gi"),
						},
						Default: v1.ResourceList{
							v1.ResourceMemory: resource.MustParse("2Gi"),
							v1.ResourceCPU:    resource.MustParse("2Gi"),
						},
					},
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
				Template: v1.PodTemplateSpec{
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{Name: "test1"},
						},
					},
				},
			},
			Status: apps.StatefulSetStatus{
				Replicas: 2,
			},
		},
	}
}

func fixStatefulSetExceeded() []*apps.StatefulSet {
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
				Template: v1.PodTemplateSpec{
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Resources: v1.ResourceRequirements{
									Limits: v1.ResourceList{
										v1.ResourceMemory: resource.MustParse("2Gi"),
									},
								},
							},
						},
					},
				},
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
				Replicas: ptrInt32(3),
				Template: v1.PodTemplateSpec{
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Resources: v1.ResourceRequirements{
									Limits: v1.ResourceList{
										v1.ResourceMemory: resource.MustParse("2Gi"),
										v1.ResourceCPU:    resource.MustParse("10"),
									},
								},
							},
						},
					},
				},
			},
			Status: apps.ReplicaSetStatus{
				Replicas: 1,
			},
		},
	}
}

func fixReplicaSetExceededWithNoLimits() []*apps.ReplicaSet {
	return []*apps.ReplicaSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fixName(),
				Namespace: fixNamespaceName(),
			},
			Spec: apps.ReplicaSetSpec{
				Replicas: ptrInt32(3),
				Template: v1.PodTemplateSpec{
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{Name: "test1"},
						},
					},
				},
			},
			Status: apps.ReplicaSetStatus{
				Replicas: 1,
			},
		},
	}
}

func fixReplicaSetExceededWithOwnerReference() []*apps.ReplicaSet {
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
				Replicas: ptrInt32(3),
				Template: v1.PodTemplateSpec{
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Resources: v1.ResourceRequirements{
									Limits: v1.ResourceList{
										v1.ResourceMemory: resource.MustParse("2Gi"),
									},
								},
							},
						},
					},
				},
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

func fixDeploy() *api.Deployment {
	return &api.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fixDeployName(),
			Namespace: fixNamespaceName(),
		},
		Spec: api.DeploymentSpec{
			Strategy: api.DeploymentStrategy{
				Type: api.RollingUpdateDeploymentStrategyType,
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
