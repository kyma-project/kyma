package controller_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/binding-usage-controller/internal/controller"
	"github.com/kyma-project/kyma/components/binding-usage-controller/internal/controller/automock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	fakeDynamic "k8s.io/client-go/dynamic/fake"
)

func TestGenericSupervisor_EnsureLabelsCreated(t *testing.T) {
	// GIVEN
	labelOp := &automock.LabelsSvc{}
	tracerOp := &automock.GenericUsageBindingAnnotationTracer{}
	defer labelOp.AssertExpectations(t)
	defer tracerOp.AssertExpectations(t)

	usageKind := newUnstructured("servicecatalog.kyma.cx/v1alpha1", "UsageKind", "test", "test")
	labels := fixLabels()

	labelOp.On("EnsureLabelsAreApplied", usageKind, labels).Return(nil)
	labelOp.On("DetectLabelsConflicts", usageKind, labels).Return([]string{}, nil)
	tracerOp.On("SetAnnotationAboutBindingUsage", usageKind, "test", labels).Return(nil)

	scheme := runtime.NewScheme()
	client := fakeDynamic.NewSimpleDynamicClient(scheme, usageKind)

	resourceInterface := client.Resource(schema.GroupVersionResource{Group: "servicecatalog.kyma.cx", Version: "v1alpha1", Resource: "usagekinds"})

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

	usageKind := newUnstructured("servicecatalog.kyma.cx/v1alpha1", "UsageKind", "test", "test")
	labels := fixLabels()

	labelOp.On("EnsureLabelsAreDeleted", usageKind, labels).Return(nil)
	tracerOp.On("GetInjectedLabels", usageKind, "test").Return(labels, nil)
	tracerOp.On("DeleteAnnotationAboutBindingUsage", usageKind, "test").Return(nil)

	scheme := runtime.NewScheme()
	client := fakeDynamic.NewSimpleDynamicClient(scheme, usageKind)

	resourceInterface := client.Resource(schema.GroupVersionResource{Group: "servicecatalog.kyma.cx", Version: "v1alpha1", Resource: "usagekinds"})

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

	usageKind := newUnstructured("servicecatalog.kyma.cx/v1alpha1", "UsageKind", "test", "test")
	labels := fixLabels()

	tracerOp.On("GetInjectedLabels", usageKind, "test").Return(labels, nil)

	scheme := runtime.NewScheme()
	client := fakeDynamic.NewSimpleDynamicClient(scheme, usageKind)

	resourceInterface := client.Resource(schema.GroupVersionResource{Group: "servicecatalog.kyma.cx", Version: "v1alpha1", Resource: "usagekinds"})

	logErrSink := newLogSinkForErrors()
	ctrl := controller.NewGenericSupervisor(resourceInterface, nil, logErrSink.Logger).WithUsageAnnotationTracer(tracerOp)

	// WHEN
	result, err := ctrl.GetInjectedLabels("test", "test", "test")

	// THEN
	require.NoError(t, err)
	require.Equal(t, labels, result)
}

func TestGenericSupervisor_DetectLabelsConflict_Err(t *testing.T) {
	// GIVEN
	labelOp := &automock.LabelsSvc{}
	defer labelOp.AssertExpectations(t)

	labels := fixLabels()
	usageKind := newUnstructured("servicecatalog.kyma.cx/v1alpha1", "UsageKind", "test", "test")
	usageKind.SetLabels(labels)

	labelOp.On("DetectLabelsConflicts", usageKind, labels).Return([]string{"label"}, errors.New("fix"))

	scheme := runtime.NewScheme()
	client := fakeDynamic.NewSimpleDynamicClient(scheme, usageKind)

	resourceInterface := client.Resource(schema.GroupVersionResource{Group: "servicecatalog.kyma.cx", Version: "v1alpha1", Resource: "usagekinds"})

	logErrSink := newLogSinkForErrors()
	ctrl := controller.NewGenericSupervisor(resourceInterface, labelOp, logErrSink.Logger)

	// WHEN
	err := ctrl.EnsureLabelsCreated("test", "test", "test", labels)

	// THEN
	assert.Error(t, err)
}

func newUnstructured(apiVersion, kind, namespace, name string) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": apiVersion,
			"kind":       kind,
			"metadata": map[string]interface{}{
				"namespace": namespace,
				"name":      name,
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
