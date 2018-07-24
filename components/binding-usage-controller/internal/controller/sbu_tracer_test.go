package controller_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/binding-usage-controller/internal/controller"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	fixUsageName = "test-usage"
)

func TestUsageAnnotationTracerInjectedLabels(t *testing.T) {
	// given
	tracer := controller.NewUsageAnnotationTracer()
	testedObjMeta := metaV1.ObjectMeta{}

	fixLabels := map[string]string{
		"k": "v",
	}

	// when
	tracer.SetAnnotationAboutBindingUsage(&testedObjMeta, fixUsageName, fixLabels)

	// then
	got, err := tracer.GetInjectedLabels(testedObjMeta, fixUsageName)
	require.NoError(t, err)
	assertEqualMaps(t, fixLabels, got)

	// when
	tracer.DeleteAnnotationAboutBindingUsage(&testedObjMeta, fixUsageName)

	// then
	got, err = tracer.GetInjectedLabels(testedObjMeta, fixUsageName)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestUsageAnnotationTracerWithCorruptedAnnotation(t *testing.T) {
	// given
	tracer := controller.NewUsageAnnotationTracer()
	testedObjMeta := metaV1.ObjectMeta{
		Annotations: map[string]string{
			controller.TracingAnnotationKey: "corrupted",
		},
	}
	fixLabels := map[string]string{
		"lab1": "val1",
	}

	// when
	gotKeys, err := tracer.GetInjectedLabels(testedObjMeta, fixUsageName)

	// then
	assert.Empty(t, gotKeys)
	assert.NotNil(t, err)

	// when
	err = tracer.SetAnnotationAboutBindingUsage(&testedObjMeta, fixUsageName, fixLabels)

	// then
	assert.Error(t, err)
}
