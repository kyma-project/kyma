package k8s

import (
	"testing"
	"time"

	gqlschema "github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	testingUtils "github.com/kyma-project/kyma/components/console-backend-service/internal/testing"
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
	svc := newLimitRangeService(informer, nil)
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
	svc := newLimitRangeService(informer, nil)
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

func TestLimitRange_Create(t *testing.T) {
	fakeClientSet := fake.NewSimpleClientset().CoreV1()

	informer := fixLimitRangeInformer()
	svc := newLimitRangeService(informer, fakeClientSet)

	namespace := "examplenamespace"
	name := "limitrangeexample"

	t.Run("Limit Range creation successful", func(t *testing.T) {
		limitRangeGQL := fixLimitRangeFromProperties("512Mi", "512Mi", "512Mi", "Container")
		_, err := svc.Create(namespace, name, limitRangeGQL)
		require.NoError(t, err)
	})

	t.Run("Limit Range creation failed, wrong unit used for memory", func(t *testing.T) {
		limitRangeGQL := fixLimitRangeFromProperties("512MGi", "512Mi", "512Mi", "Container")
		_, err := svc.Create(namespace, name, limitRangeGQL)
		require.Error(t, err)
	})

	t.Run("Limit Range creation failed, wrong limit range type", func(t *testing.T) {
		limitRangeGQL := fixLimitRangeFromProperties("512Mi", "512Mi", "512Mi", "RANDOM")
		_, err := svc.Create(namespace, name, limitRangeGQL)
		require.Error(t, err)
	})
}

func fixLimitRangeFromProperties(defaultMem string, defaultRequestMem string, maxMem string, lrType string) gqlschema.LimitRangeInput {
	return gqlschema.LimitRangeInput{
		Default: gqlschema.ResourceValuesInput{
			Memory: &defaultMem,
		},
		DefaultRequest: gqlschema.ResourceValuesInput{
			Memory: &defaultRequestMem,
		},
		Max: gqlschema.ResourceValuesInput{
			Memory: &maxMem,
		},
		Type: lrType,
	}
}
