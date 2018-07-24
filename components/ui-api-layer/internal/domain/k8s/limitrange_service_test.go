package k8s

import (
	"testing"
	"time"

	testingUtils "github.com/kyma-project/kyma/components/ui-api-layer/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
)

func TestLimitRangeService_List(t *testing.T) {
	// GIVEN
	informer := fixLimitRangeInformer(fixLimitRange())
	svc := newLimitRangeService(informer)
	testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

	// WHEN
	result, err := svc.List(fixLimitRangeNamespace())

	// THEN
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Contains(t, result, fixLimitRange())
}

func TestLimitRangeService_List_NotFound(t *testing.T) {
	// GIVEN
	informer := fixLimitRangeInformer()
	svc := newLimitRangeService(informer)
	testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

	// WHEN
	result, err := svc.List("env")

	// THEN
	require.NoError(t, err)
	assert.Len(t, result, 0)
}

func fixLimitRangeInformer(objects ...runtime.Object) cache.SharedIndexInformer {
	client := fake.NewSimpleClientset(objects...)
	informerFactory := informers.NewSharedInformerFactory(client, 0)

	return informerFactory.Core().V1().LimitRanges().Informer()
}
