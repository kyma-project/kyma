package k8s

import (
	"testing"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/k8s/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apps "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	rqSvc := newResourceQuotaStatusService(rqLister, rsLister, ssLister, lrLister)

	// WHEN
	status, err := rqSvc.CheckResourceQuotaStatus(fixNamespaceName())
	require.NoError(t, err)

	// THEN
	assert.True(t, status.Exceeded)
	assert.Len(t, status.ExceededQuotas, 1)
	assert.Len(t, status.ExceededQuotas[0].AffectedResources, 1)
	assert.Equal(t, fixName(), status.ExceededQuotas[0].QuotaName)
	assert.Equal(t, v1.ResourceLimitsMemory, v1.ResourceName(status.ExceededQuotas[0].ResourceName))
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

	rqSvc := newResourceQuotaStatusService(rqLister, rsLister, ssLister, lrLister)

	// WHEN
	status, err := rqSvc.CheckResourceQuotaStatus(fixNamespaceName())
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

	rqSvc := newResourceQuotaStatusService(rqLister, rsLister, ssLister, lrLister)

	// WHEN
	status, err := rqSvc.CheckResourceQuotaStatus(fixNamespaceName())
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

	rqSvc := newResourceQuotaStatusService(rqLister, rsLister, ssLister, lrLister)

	// WHEN
	status, err := rqSvc.CheckResourceQuotaStatus(fixNamespaceName())
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

	rqSvc := newResourceQuotaStatusService(rqLister, rsLister, ssLister, lrLister)

	// WHEN
	status, err := rqSvc.CheckResourceQuotaStatus(fixNamespaceName())
	require.NoError(t, err)

	// THEN
	assert.True(t, status.Exceeded)
	assert.Len(t, status.ExceededQuotas, 1)
	assert.Len(t, status.ExceededQuotas[0].AffectedResources, 1)
}

func TestResourceQuotaStatusService_CheckResourceQuotaStatus_WithoutResourceQuotasAndLimitRanges(t *testing.T) {
	// GIVEN
	rqLister := automock.NewResourceQuotaLister()
	rqLister.On("ListResourceQuotas", fixNamespaceName()).Return([]*v1.ResourceQuota{}, nil)
	rsLister := automock.NewReplicaSetLister()
	rsLister.On("ListReplicaSets", fixNamespaceName()).Return(fixReplicaSetExceeded(), nil)
	ssLister := automock.NewStatefulSetLister()
	ssLister.On("ListStatefulSets", fixNamespaceName()).Return(fixStatefulSet(), nil)
	lrLister := automock.NewLimitRangeLister()
	lrLister.On("List", fixNamespaceName()).Return([]*v1.LimitRange{}, nil)
	defer func() {
		rqLister.AssertExpectations(t)
		rsLister.AssertExpectations(t)
		ssLister.AssertExpectations(t)
		lrLister.AssertExpectations(t)
	}()

	rqSvc := newResourceQuotaStatusService(rqLister, rsLister, ssLister, lrLister)

	// WHEN
	status, err := rqSvc.CheckResourceQuotaStatus(fixNamespaceName())
	require.NoError(t, err)

	// THEN
	assert.False(t, status.Exceeded)
	assert.Len(t, status.ExceededQuotas, 0)
}

func TestResourceQuotaStatusService_CheckResourceQuotaStatus_WithoutDefaultLimits(t *testing.T) {
	// GIVEN
	rqLister := automock.NewResourceQuotaLister()
	rqLister.On("ListResourceQuotas", fixNamespaceName()).Return(fixResourceQuota(), nil)
	rsLister := automock.NewReplicaSetLister()
	rsLister.On("ListReplicaSets", fixNamespaceName()).Return(fixReplicaSetExceededWithNoLimits(), nil)
	ssLister := automock.NewStatefulSetLister()
	ssLister.On("ListStatefulSets", fixNamespaceName()).Return(fixStatefulSet(), nil)
	lrLister := automock.NewLimitRangeLister()
	lrLister.On("List", fixNamespaceName()).Return([]*v1.LimitRange{}, nil)
	defer func() {
		rqLister.AssertExpectations(t)
		rsLister.AssertExpectations(t)
		ssLister.AssertExpectations(t)
		lrLister.AssertExpectations(t)
	}()

	rqSvc := newResourceQuotaStatusService(rqLister, rsLister, ssLister, lrLister)

	// WHEN
	status, err := rqSvc.CheckResourceQuotaStatus(fixNamespaceName())
	require.NoError(t, err)

	// THEN
	assert.False(t, status.Exceeded)
	assert.Len(t, status.ExceededQuotas, 0)
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
			Status: v1.ResourceQuotaStatus{
				Used: v1.ResourceList{
					v1.ResourceLimitsMemory:   resource.MustParse("0"),
					v1.ResourceRequestsMemory: resource.MustParse("0"),
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

func ptrInt32(int int32) *int32 {
	return &int
}
