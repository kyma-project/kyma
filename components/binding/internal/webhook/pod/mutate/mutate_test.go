package mutate

import (
	"context"
	"fmt"
	"github.com/kyma-project/kyma/components/binding/internal/webhook"
	"testing"

	"github.com/kyma-project/kyma/components/binding/pkg/apis/v1alpha1"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gomodules.xyz/jsonpatch/v2"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	namespace        = "test"
	bindingNameOne   = "binding-test-one"
	bindingNameTwo   = "binding-test-two"
	bindingNameThree = "binding-test-three"
	secretNameOne    = "secret-test-one"
	secretNameTwo    = "secret-test-two"
	configMapName    = "config-map-test"
	secretEnvOneKey  = "PASSWORD"
	secretEnvTwoKey  = "TOKEN"
	configMapEnvKey  = "CONFIG"
)

func TestMutationHandler_Handle(t *testing.T) {
	// given
	sch, err := v1alpha1.SchemeBuilder.Build()
	require.NoError(t, err)
	err = scheme.AddToScheme(sch)
	require.NoError(t, err)

	request := admission.Request{
		AdmissionRequest: v1beta1.AdmissionRequest{
			UID:       "1234-abcd",
			Operation: v1beta1.Create,
			Name:      "test-pod",
			Namespace: namespace,
			Kind: metav1.GroupVersionKind{
				Kind:    "Pod",
				Version: "v1",
				Group:   "",
			},
			Object: runtime.RawExtension{Raw: rawPod()},
		},
	}

	fakeClient := fake.NewFakeClientWithScheme(sch, fixBindingOne(), fixBindingTwo(), fixBindingThree(), fixSecretOne(), fixSecretTwo(), fixConfigMap())
	decoder, err := admission.NewDecoder(scheme.Scheme)
	require.NoError(t, err)

	handler := NewMutationHandler(webhook.NewClient(fakeClient), logrus.New())
	err = handler.InjectDecoder(decoder)
	require.NoError(t, err)

	// when
	response := handler.Handle(context.TODO(), request)

	// then
	assert.True(t, response.Allowed)

	// filtering out status cause k8s api-server will discard this too
	patches := filterOutStatusPatch(response.Patches)
	assert.Len(t, patches, 2)

	for _, patch := range patches {
		assert.Equal(t, "add", patch.Operation)
		assert.Contains(t, []string{"/spec/containers/0/env", "/spec/containers/1/env"}, patch.Path)
		assert.Len(t, patch.Value, 3)
		assert.ElementsMatch(t, patch.Value, []interface{}{
			map[string]interface{}{
				"name": secretEnvOneKey,
				"valueFrom": map[string]interface{}{
					"secretKeyRef": map[string]interface{}{
						"key":  secretEnvOneKey,
						"name": secretNameOne,
					},
				},
			},
			map[string]interface{}{
				"name": secretEnvTwoKey,
				"valueFrom": map[string]interface{}{
					"secretKeyRef": map[string]interface{}{
						"key":  secretEnvTwoKey,
						"name": secretNameTwo,
					},
				},
			},
			map[string]interface{}{
				"name": configMapEnvKey,
				"valueFrom": map[string]interface{}{
					"configMapKeyRef": map[string]interface{}{
						"key":  configMapEnvKey,
						"name": configMapName,
					},
				},
			},
		})
	}
}

func fixBindingOne() *v1alpha1.Binding {
	return &v1alpha1.Binding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bindingNameOne,
			Namespace: namespace,
		},
		Spec: v1alpha1.BindingSpec{
			Source: v1alpha1.Source{
				Kind: v1alpha1.SourceKindSecret,
				Name: secretNameOne,
			},
		},
		Status: v1alpha1.BindingStatus{
			Phase:   v1alpha1.BindingReady,
			Message: "la loza lorem ipsum dolores sit onface",
			Source:  fmt.Sprintf("%s/%s", v1alpha1.SourceKindSecret, secretNameOne),
		},
	}
}

func fixBindingTwo() *v1alpha1.Binding {
	return &v1alpha1.Binding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bindingNameTwo,
			Namespace: namespace,
		},
		Spec: v1alpha1.BindingSpec{
			Source: v1alpha1.Source{
				Kind: v1alpha1.SourceKindSecret,
				Name: secretNameTwo,
			},
		},
		Status: v1alpha1.BindingStatus{
			Phase:   v1alpha1.BindingReady,
			Message: "lorem ipsum dolor sit amet",
			Source:  fmt.Sprintf("%s/%s", v1alpha1.SourceKindSecret, secretNameTwo),
		},
	}
}

func fixBindingThree() *v1alpha1.Binding {
	return &v1alpha1.Binding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bindingNameThree,
			Namespace: namespace,
		},
		Spec: v1alpha1.BindingSpec{
			Source: v1alpha1.Source{
				Kind: v1alpha1.SourceKindConfigMap,
				Name: configMapName,
			},
		},
		Status: v1alpha1.BindingStatus{
			Phase:   v1alpha1.BindingReady,
			Message: "consectetur adipiscing elit, sed do eiusmod tempor incididunt",
			Source:  fmt.Sprintf("%s/%s", v1alpha1.SourceKindConfigMap, configMapName),
		},
	}
}

func fixSecretOne() *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretNameOne,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			secretEnvOneKey: []byte("superSecretPassword"),
		},
	}
}

func fixSecretTwo() *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretNameTwo,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			secretEnvTwoKey: []byte("superSecretToken"),
		},
	}
}

func fixConfigMap() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: namespace,
		},
		Data: map[string]string{
			configMapEnvKey: "configForPod",
		},
	}
}

func rawPod() []byte {
	return []byte(fmt.Sprintf(`{
		"apiVersion": "v1",
		"kind": "Pod",
		"metadata": {
		  "creationTimestamp": null,
		  "name": "test-pod",
		  "labels": {
			"%s/%s": "1234",
			"%s/%s": "4567",
			"%s/%s": "9875"
		  }
		},
		"spec": {
		  "containers": [
			{
			  "name": "test",
			  "image": "test",
			  "resources": {}
            },
			{
			  "name": "test2",
			  "image": "test2",
			  "resources": {}
			}
		  ]
		}
	}`, v1alpha1.BindingLabelKey, bindingNameOne, v1alpha1.BindingLabelKey, bindingNameTwo, v1alpha1.BindingLabelKey, bindingNameThree))
}

func filterOutStatusPatch(operations []jsonpatch.JsonPatchOperation) []jsonpatch.JsonPatchOperation {
	var filtered []jsonpatch.JsonPatchOperation
	for _, op := range operations {
		if op.Path != "/status" {
			filtered = append(filtered, op)
		}
	}

	return filtered
}
