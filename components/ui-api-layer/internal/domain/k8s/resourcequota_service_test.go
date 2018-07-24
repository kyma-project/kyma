package k8s_test

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/k8s"
	testingUtils "github.com/kyma-project/kyma/components/ui-api-layer/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
)

func TestResourceQuotaResolver_ListSuccess(t *testing.T) {
	// GIVEN
	rq1 := fixResourceQuota("rq1", "prod")
	rq2 := fixResourceQuota("rq2", "prod")
	rqQa := fixResourceQuota("rq", "qa")
	informer := fixResourceQuotaInformer(rq1, rq2, rqQa)

	svc := k8s.NewResourceQuotaService(informer)
	testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

	// WHEN
	result, err := svc.List("prod")

	// THEN
	require.NoError(t, err)
	assert.Contains(t, result, rq1)
	assert.Contains(t, result, rq2)
	assert.Len(t, result, 2)
}

func fixResourceQuotaInformer(objects ...runtime.Object) cache.SharedIndexInformer {
	client := fake.NewSimpleClientset(objects...)
	informerFactory := informers.NewSharedInformerFactory(client, 0)

	return informerFactory.Core().V1().ResourceQuotas().Informer()
}

func fixResourceQuota(name, environment string) *v1.ResourceQuota {
	return &v1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: environment,
		},
	}
}
