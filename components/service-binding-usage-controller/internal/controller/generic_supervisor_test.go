package controller_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/service-binding-usage-controller/internal/controller"
	"github.com/kyma-project/kyma/components/service-binding-usage-controller/internal/controller/automock"
	"github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	fakeDynamic "k8s.io/client-go/dynamic/fake"
)

func TestGenericSupervisor_EnsureLabelsCreated(t *testing.T) {
	// GIVEN
	labelOp := &automock.LabelsSvc{}
	tracerOp := &automock.GenericUsageBindingAnnotationTracer{}
	defer labelOp.AssertExpectations(t)
	defer tracerOp.AssertExpectations(t)

	usageKind := fixUnstructuredUK()
	labels := fixLabels()

	labelOp.On("EnsureLabelsAreApplied", usageKind, labels).Return(nil)
	tracerOp.On("SetAnnotationAboutBindingUsage", usageKind, "test", labels).Return(nil)

	scheme := runtime.NewScheme()
	client := fakeDynamic.NewSimpleDynamicClient(scheme, usageKind)

	resourceInterface := client.Resource(v1alpha1.SchemeGroupVersion.WithResource("usagekinds"))

	logErrSink := newLogSinkForErrors()
	ctrl := controller.NewGenericSupervisor(resourceInterface, labelOp, logErrSink.Logger).WithUsageAnnotationTracer(tracerOp)

	// WHEN
	err := ctrl.EnsureLabelsCreated("test", "test", "test", labels)

	// THEN
	require.NoError(t, err)
}

func TestGenericSupervisor_EnsureLabelsDeleted(t *testing.T) {
	// GIVEN
	labelOp := &automock.LabelsSvc{}
	tracerOp := &automock.GenericUsageBindingAnnotationTracer{}
	defer labelOp.AssertExpectations(t)
	defer tracerOp.AssertExpectations(t)

	usageKind := fixUnstructuredUK()
	labels := fixLabels()

	labelOp.On("EnsureLabelsAreDeleted", usageKind, labels).Return(nil)
	tracerOp.On("GetInjectedLabels", usageKind, "test").Return(labels, nil)
	tracerOp.On("DeleteAnnotationAboutBindingUsage", usageKind, "test").Return(nil)

	scheme := runtime.NewScheme()
	client := fakeDynamic.NewSimpleDynamicClient(scheme, usageKind)

	resourceInterface := client.Resource(v1alpha1.SchemeGroupVersion.WithResource("usagekinds"))

	logErrSink := newLogSinkForErrors()
	ctrl := controller.NewGenericSupervisor(resourceInterface, labelOp, logErrSink.Logger).WithUsageAnnotationTracer(tracerOp)

	// WHEN
	err := ctrl.EnsureLabelsDeleted("test", "test", "test")

	// THEN
	require.NoError(t, err)
}

func TestGenericSupervisor_GetInjectedLabels(t *testing.T) {
	// GIVEN
	tracerOp := &automock.GenericUsageBindingAnnotationTracer{}
	defer tracerOp.AssertExpectations(t)

	usageKind := fixUnstructuredUK()
	labels := fixLabels()

	tracerOp.On("GetInjectedLabels", usageKind, "test").Return(labels, nil)

	scheme := runtime.NewScheme()
	client := fakeDynamic.NewSimpleDynamicClient(scheme, usageKind)

	resourceInterface := client.Resource(v1alpha1.SchemeGroupVersion.WithResource("usagekinds"))

	logErrSink := newLogSinkForErrors()
	ctrl := controller.NewGenericSupervisor(resourceInterface, nil, logErrSink.Logger).WithUsageAnnotationTracer(tracerOp)

	// WHEN
	result, err := ctrl.GetInjectedLabels("test", "test", "test")

	// THEN
	require.NoError(t, err)
	require.Equal(t, labels, result)
}

func fixUnstructuredUK() *unstructured.Unstructured {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "servicecatalog.kyma-project.io/v1alpha1",
			"kind":       "UsageKind",
			"metadata": map[string]interface{}{
				"namespace": "test",
				"name":      "test",
			},
		},
	}
	return obj
}

func fixLabels() map[string]string {
	return map[string]string{
		"label": "label",
	}
}
