package target

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-project/kyma/components/binding/internal/storage"
	"github.com/kyma-project/kyma/components/binding/pkg/apis/v1alpha1"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicFake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/scheme"
	k8sTesting "k8s.io/client-go/testing"
)

const (
	namespace   = "test"
	bindingName = "binding-test"
	secretName  = "secret-test"
	deployName  = "deploy"
	targetKind  = "Deployment"

	labelsPath         = "spec.template.metadata.labels"
	existingLabelKey   = "existing-label"
	existingLabelValue = "should-not-be-changed"
	bindingLabelValue  = "2e95db41-2075-4e94-aa61-3f2cf6d705d5"
)

func TestHandler_AddLabel(t *testing.T) {
	for tn, tc := range map[string]struct {
		Deployment *v1.Deployment
	}{
		"when labels are empty": {
			Deployment: fixDeployment("withEmptyLabels"),
		},
		"not empty labels": {
			Deployment: fixDeployment("withNotEmptyLabels"),
		},
		"not existing labels": {
			Deployment: fixDeployment("withNotExistingLabels"),
		},
	} {
		t.Run(tn, func(t *testing.T) {
			// given
			sch, err := v1alpha1.SchemeBuilder.Build()
			require.NoError(t, err)
			err = scheme.AddToScheme(sch)
			require.NoError(t, err)

			fixedBinding := fixBinding()

			kindStorage := storage.NewKindStorage()
			err = kindStorage.Register(*fixTargetKind())
			require.NoError(t, err)

			dc := dynamicFake.NewSimpleDynamicClient(sch, tc.Deployment, fixSecret())
			handler := NewHandler(dc, kindStorage)

			// when
			err = handler.AddLabel(fixedBinding)
			require.NoError(t, err)

			// then
			resource, err := dc.Resource(schema.GroupVersionResource{
				Group:    "apps",
				Version:  "v1",
				Resource: "deployments",
			}).Namespace(namespace).Get(context.Background(), fixedBinding.Spec.Target.Name, metav1.GetOptions{})
			require.NoError(t, err)

			unstructured := resource.UnstructuredContent()
			var deployment v1.Deployment
			err = runtime.DefaultUnstructuredConverter.
				FromUnstructured(unstructured, &deployment)
			require.NoError(t, err)

			assert.Contains(t, deployment.Spec.Template.Labels, handler.labelKey(fixedBinding))
		})
	}

	t.Run("should return error when Deployment does not exist", func(t *testing.T) {
		// given
		sch, err := v1alpha1.SchemeBuilder.Build()
		require.NoError(t, err)
		err = scheme.AddToScheme(sch)
		require.NoError(t, err)

		kindStorage := storage.NewKindStorage()
		err = kindStorage.Register(*fixTargetKind())
		require.NoError(t, err)

		dc := dynamicFake.NewSimpleDynamicClient(sch, fixSecret())
		handler := NewHandler(dc, kindStorage)

		// when
		err = handler.AddLabel(fixBinding())
		assert.Error(t, err)
	})

	t.Run("should return error when couldn't get resource", func(t *testing.T) {
		// given
		sch, err := v1alpha1.SchemeBuilder.Build()
		require.NoError(t, err)
		err = scheme.AddToScheme(sch)
		require.NoError(t, err)

		fixedBinding := fixBinding()
		fixedTargetKind := fixTargetKind()

		kindStorage := storage.NewKindStorage()
		err = kindStorage.Register(*fixedTargetKind)
		require.NoError(t, err)

		dc := dynamicFake.NewSimpleDynamicClient(sch, fixSecret())
		handler := NewHandler(dc, kindStorage)

		// when
		err = handler.AddLabel(fixedBinding)
		assert.Error(t, err)
	})

	t.Run("should return error when couldn't update resource", func(t *testing.T) {
		// given
		sch, err := v1alpha1.SchemeBuilder.Build()
		require.NoError(t, err)
		err = scheme.AddToScheme(sch)
		require.NoError(t, err)

		fixedBinding := fixBinding()
		fixedDeploy := fixDeployment("withEmptyLabels")
		fixedTargetKind := fixTargetKind()

		kindStorage := storage.NewKindStorage()
		err = kindStorage.Register(*fixedTargetKind)
		require.NoError(t, err)

		dc := dynamicFake.NewSimpleDynamicClient(sch, fixedDeploy, fixSecret())
		dc.PrependReactor("update", "deployments", func(action k8sTesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, nil, errors.New("update error")
		})
		handler := NewHandler(dc, kindStorage)

		// when
		err = handler.AddLabel(fixedBinding)
		assert.Error(t, err)
	})
}

func TestHandler_LabelExist(t *testing.T) {
	for tn, tc := range map[string]struct {
		Deployment *v1.Deployment
		expected   bool
	}{
		"should return true when label exists": {
			Deployment: fixDeployment("withBindingLabel"),
			expected:   true,
		},
		"should return false when label does not exists": {
			Deployment: fixDeployment("withNotExistingLabels"),
			expected:   false,
		},
	} {
		t.Run(tn, func(t *testing.T) {
			// given
			sch, err := v1alpha1.SchemeBuilder.Build()
			require.NoError(t, err)
			err = scheme.AddToScheme(sch)
			require.NoError(t, err)

			fixedBinding := fixBinding()

			kindStorage := storage.NewKindStorage()
			err = kindStorage.Register(*fixTargetKind())
			require.NoError(t, err)

			dc := dynamicFake.NewSimpleDynamicClient(sch, tc.Deployment, fixSecret())
			handler := NewHandler(dc, kindStorage)

			// when
			result, err := handler.LabelExist(fixedBinding)
			require.NoError(t, err)

			// then
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestHandler_RemoveOldAddNewLabel(t *testing.T) {
	t.Run("should regenerate only Binding label when exists", func(t *testing.T) {
		// given
		sch, err := v1alpha1.SchemeBuilder.Build()
		require.NoError(t, err)
		err = scheme.AddToScheme(sch)
		require.NoError(t, err)

		fixedBinding := fixBinding()
		fixedDeploy := fixDeployment("withBindingLabel")

		kindStorage := storage.NewKindStorage()
		err = kindStorage.Register(*fixTargetKind())
		require.NoError(t, err)

		dc := dynamicFake.NewSimpleDynamicClient(sch, fixedDeploy, fixSecret())
		handler := NewHandler(dc, kindStorage)

		// when
		err = handler.RemoveOldAddNewLabel(fixedBinding)
		require.NoError(t, err)

		// then
		resource, err := dc.Resource(schema.GroupVersionResource{
			Group:    "apps",
			Version:  "v1",
			Resource: "deployments",
		}).Namespace(namespace).Get(context.Background(), fixedBinding.Spec.Target.Name, metav1.GetOptions{})
		require.NoError(t, err)

		unstructured := resource.UnstructuredContent()
		var deployment v1.Deployment
		err = runtime.DefaultUnstructuredConverter.
			FromUnstructured(unstructured, &deployment)
		require.NoError(t, err)

		assert.Contains(t, deployment.Spec.Template.Labels, handler.labelKey(fixedBinding))
		assert.NotContains(t, deployment.Spec.Template.Labels, bindingLabelValue)
		assert.NotEqual(t, fixedDeploy.Spec.Template.Labels[handler.labelKey(fixedBinding)], deployment.Spec.Template.Labels[handler.labelKey(fixedBinding)])
		assert.Contains(t, deployment.Spec.Template.Labels, existingLabelKey)
		assert.Equal(t, deployment.Spec.Template.Labels[existingLabelKey], existingLabelValue)
	})

	t.Run("should not modify any labels when Binding label doesn't exists", func(t *testing.T) {
		// given
		sch, err := v1alpha1.SchemeBuilder.Build()
		require.NoError(t, err)
		err = scheme.AddToScheme(sch)
		require.NoError(t, err)

		fixedBinding := fixBinding()

		kindStorage := storage.NewKindStorage()
		err = kindStorage.Register(*fixTargetKind())
		require.NoError(t, err)

		dc := dynamicFake.NewSimpleDynamicClient(sch, fixDeployment("withNotEmptyLabels"), fixSecret())
		handler := NewHandler(dc, kindStorage)

		// when
		err = handler.RemoveOldAddNewLabel(fixedBinding)
		require.NoError(t, err)

		// then
		resource, err := dc.Resource(schema.GroupVersionResource{
			Group:    "apps",
			Version:  "v1",
			Resource: "deployments",
		}).Namespace(namespace).Get(context.Background(), fixedBinding.Spec.Target.Name, metav1.GetOptions{})
		require.NoError(t, err)

		unstructured := resource.UnstructuredContent()
		var deployment v1.Deployment
		err = runtime.DefaultUnstructuredConverter.
			FromUnstructured(unstructured, &deployment)
		require.NoError(t, err)

		assert.NotContains(t, deployment.Spec.Template.Labels, handler.labelKey(fixedBinding))
		assert.Contains(t, deployment.Spec.Template.Labels, existingLabelKey)
		assert.Equal(t, deployment.Spec.Template.Labels[existingLabelKey], existingLabelValue)
	})
}

func TestHandler_RemoveLabel(t *testing.T) {
	t.Run("should remove only Binding labels", func(t *testing.T) {
		// given
		sch, err := v1alpha1.SchemeBuilder.Build()
		require.NoError(t, err)
		err = scheme.AddToScheme(sch)
		require.NoError(t, err)

		fixedBinding := fixBinding()
		fixedDeploy := fixDeployment("withBindingLabel")

		kindStorage := storage.NewKindStorage()
		err = kindStorage.Register(*fixTargetKind())
		require.NoError(t, err)

		dc := dynamicFake.NewSimpleDynamicClient(sch, fixedDeploy, fixSecret())
		handler := NewHandler(dc, kindStorage)

		// when
		err = handler.RemoveLabel(fixedBinding)
		require.NoError(t, err)

		// then
		resource, err := dc.Resource(schema.GroupVersionResource{
			Group:    "apps",
			Version:  "v1",
			Resource: "deployments",
		}).Namespace(namespace).Get(context.Background(), fixedBinding.Spec.Target.Name, metav1.GetOptions{})
		require.NoError(t, err)

		unstructured := resource.UnstructuredContent()
		var deployment v1.Deployment
		err = runtime.DefaultUnstructuredConverter.
			FromUnstructured(unstructured, &deployment)
		require.NoError(t, err)

		assert.NotContains(t, deployment.Spec.Template.Labels, handler.labelKey(fixedBinding))
		assert.Contains(t, deployment.Spec.Template.Labels, existingLabelKey)
		assert.Equal(t, deployment.Spec.Template.Labels[existingLabelKey], existingLabelValue)
	})
}

func fixBinding() *v1alpha1.Binding {
	return &v1alpha1.Binding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bindingName,
			Namespace: namespace,
		},
		Spec: v1alpha1.BindingSpec{
			Source: v1alpha1.Source{
				Kind: v1alpha1.SourceKindSecret,
				Name: secretName,
			},
			Target: v1alpha1.Target{
				Kind: targetKind,
				Name: deployName,
			},
		},
	}
}

func fixDeployment(option string) *v1.Deployment {
	switch option {
	case "withEmptyLabels":
		return &v1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deployName,
				Namespace: namespace,
			},
			Spec: v1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{},
					},
				},
			},
		}
	case "withNotEmptyLabels":
		return &v1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deployName,
				Namespace: namespace,
			},
			Spec: v1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							existingLabelKey: existingLabelValue,
						},
					},
				},
			},
		}
	case "withBindingLabel":
		return &v1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deployName,
				Namespace: namespace,
			},
			Spec: v1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							fmt.Sprintf("%s/%s", v1alpha1.BindingLabelKey, bindingName): bindingLabelValue,
							existingLabelKey: existingLabelValue,
						},
					},
				},
			},
		}
	default:
		return &v1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deployName,
				Namespace: namespace,
			},
		}
	}

}

func fixTargetKind() *v1alpha1.TargetKind {
	return &v1alpha1.TargetKind{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deployName,
			Namespace: namespace,
		},
		Spec: v1alpha1.TargetKindSpec{
			Resource: v1alpha1.Resource{
				Kind:    targetKind,
				Version: "v1",
				Group:   "apps",
			},
			LabelsPath: labelsPath,
		},
	}
}

func fixSecret() *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			"key": []byte("superSecretPassword"),
		},
	}
}
