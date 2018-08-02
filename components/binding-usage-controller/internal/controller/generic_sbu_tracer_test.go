package controller_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/binding-usage-controller/internal/controller"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestGenericUsageAnnotationTracerInjectedLabels(t *testing.T) {
	// given
	tracer := controller.NewGenericUsageAnnotationTracer()
	obj := &unstructured.Unstructured{Object: map[string]interface{}{
		"metadata": map[string]interface{}{
			"name": "deploy-test",
		},
	}}

	fixLabels := map[string]string{
		"k": "v",
	}

	// when
	tracer.SetAnnotationAboutBindingUsage(obj, fixUsageName, fixLabels)

	// then
	got, err := tracer.GetInjectedLabels(obj, fixUsageName)
	require.NoError(t, err)
	assertEqualMaps(t, fixLabels, got)

	// when
	err = tracer.DeleteAnnotationAboutBindingUsage(obj, fixUsageName)

	// then
	require.NoError(t, err)
	got, err = tracer.GetInjectedLabels(obj, fixUsageName)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestGenericUsageAnnotationTracerWithCorruptedAnnotation(t *testing.T) {
	// given
	tracer := controller.NewGenericUsageAnnotationTracer()
	obj := &unstructured.Unstructured{Object: map[string]interface{}{
		"metadata": map[string]interface{}{
			"name": "deploy-test",
			"annotations": map[string]interface{}{
				controller.TracingAnnotationKey: "corrupted",
			},
		},
	}}
	fixLabels := map[string]string{
		"lab1": "val1",
	}

	// when
	gotKeys, err := tracer.GetInjectedLabels(obj, fixUsageName)

	// then
	require.Error(t, err)
	assert.Empty(t, gotKeys)

	// when
	err = tracer.SetAnnotationAboutBindingUsage(obj, fixUsageName, fixLabels)

	// then
	assert.Error(t, err)
}

func assertEqualMaps(t *testing.T, expected, got map[string]string) {
	assert.Len(t, got, len(expected))
	for k, v := range expected {
		gotValue, found := got[k]
		assert.True(t, found)
		assert.Equal(t, v, gotValue)
	}
}
