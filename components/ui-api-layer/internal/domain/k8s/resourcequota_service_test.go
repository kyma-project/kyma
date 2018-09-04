package k8s_test

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/k8s"
	testingUtils "github.com/kyma-project/kyma/components/ui-api-layer/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apps "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
)

func TestResourceQuotaResolver_ListSuccess(t *testing.T) {
	// GIVEN
	rq1 := fixResourceQuota("rq1", "prod")
	rq2 := fixResourceQuota("rq2", "prod")
	rqQa := fixResourceQuota("rq", "qa")
	informer := fixInformer(rq1, rq2, rqQa).Core().V1().ResourceQuotas().Informer()

	svc := k8s.NewResourceQuotaService(informer, nil, nil, nil)
	testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

	// WHEN
	result, err := svc.ListResourceQuotas("prod")

	// THEN
	require.NoError(t, err)
	assert.Contains(t, result, rq1)
	assert.Contains(t, result, rq2)
	assert.Len(t, result, 2)
}

func TestResourceQuotaService_ListReplicaSets(t *testing.T) {
	// GIVEN
	rs1 := fixReplicaSet("rs1", "prod")
	rs2 := fixReplicaSet("rs2", "prod")
	rsQa := fixReplicaSet("rs", "qa")
	informer := fixInformer(rs1, rs2, rsQa).Apps().V1().ReplicaSets().Informer()

	svc := k8s.NewResourceQuotaService(nil, informer, nil, nil)
	testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

	// WHEN
	result, err := svc.ListReplicaSets("prod")

	// THEN
	require.NoError(t, err)
	assert.Contains(t, result, rs1)
	assert.Contains(t, result, rs2)
	assert.Len(t, result, 2)
}

func TestResourceQuotaService_ListStatefulSets(t *testing.T) {
	// GIVEN
	rs1 := fixStatefulSet("rs1", "prod")
	rs2 := fixStatefulSet("rs2", "prod")
	rsQa := fixStatefulSet("rs", "qa")
	informer := fixInformer(rs1, rs2, rsQa).Apps().V1().StatefulSets().Informer()

	svc := k8s.NewResourceQuotaService(nil, nil, informer, nil)
	testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

	// WHEN
	result, err := svc.ListStatefulSets("prod")

	// THEN
	require.NoError(t, err)
	assert.Equal(t, result[0], rs1)
	assert.Equal(t, result[1], rs2)
	assert.Len(t, result, 2)
}

func TestResourceQuotaService_ListPods(t *testing.T) {
	// GIVEN
	labels := map[string]string{"label": "true"}

	po1 := fixPod("po1", "prod", labels)
	po2 := fixPod("po2", "prod", labels)
	po3 := fixPod("po3", "prod", map[string]string{})
	po4 := fixPod("po4", "xd", map[string]string{})

	client := fake.NewSimpleClientset(po1, po2, po3, po4)

	svc := k8s.NewResourceQuotaService(nil, nil, nil, client.Core())

	// WHEN
	result, err := svc.ListPods("prod", labels)

	// THEN
	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func fixInformer(objects ...runtime.Object) informers.SharedInformerFactory {
	client := fake.NewSimpleClientset(objects...)
	informerFactory := informers.NewSharedInformerFactory(client, 0)

	return informerFactory
}

func fixResourceQuota(name, environment string) *v1.ResourceQuota {
	return &v1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: environment,
		},
	}
}

func fixReplicaSet(name, environment string) *apps.ReplicaSet {
	return &apps.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: environment,
		},
	}
}

func fixStatefulSet(name, environment string) *apps.StatefulSet {
	return &apps.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: environment,
		},
	}
}

func fixPod(name, environment string, labels map[string]string) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: environment,
			Labels:    labels,
		},
	}
}
