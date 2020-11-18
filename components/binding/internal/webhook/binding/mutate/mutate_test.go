package mutate

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kyma-project/kyma/components/binding/internal/storage"
	"github.com/kyma-project/kyma/components/binding/internal/webhook"
	v1 "k8s.io/api/apps/v1"
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
	dynamicFake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	namespace   = "test"
	bindingName = "binding-test"
	secretName  = "secret-test"
	deployName  = "deploy"
	targetKind  = "Deployment"
)

func TestMutationHandler_Handle(t *testing.T) {
	// given
	sch, err := v1alpha1.SchemeBuilder.Build()
	require.NoError(t, err)
	err = scheme.AddToScheme(sch)
	require.NoError(t, err)

	binding := fixBinding()
	rawBinding, err := json.Marshal(binding)
	require.NoError(t, err)

	request := admission.Request{
		AdmissionRequest: v1beta1.AdmissionRequest{
			UID:       "1234-abcd",
			Operation: v1beta1.Create,
			Name:      "test-binding",
			Namespace: namespace,
			Kind: metav1.GroupVersionKind{
				Kind:    "Binding",
				Version: "v1alpha1",
				Group:   "bindings.kyma-project.io",
			},
			Object: runtime.RawExtension{Raw: rawBinding},
		},
	}

	kindStorage := storage.NewKindStorage()
	err = kindStorage.Register(*fixTargetKind())
	require.NoError(t, err)

	dc := dynamicFake.NewSimpleDynamicClient(sch, fixDeployment())
	fakeClient := fake.NewFakeClientWithScheme(sch, fixBinding(), fixSecret())
	decoder, err := admission.NewDecoder(scheme.Scheme)
	require.NoError(t, err)

	handler := NewMutationHandler(kindStorage, webhook.NewClient(fakeClient), dc, logrus.New())
	err = handler.InjectDecoder(decoder)
	require.NoError(t, err)

	// when
	response := handler.Handle(context.TODO(), request)

	// then
	assert.True(t, response.Allowed)

	// filtering out status cause k8s api-server will discard this too
	patches := filterOutStatusPatch(response.Patches)
	fmt.Println(patches)
	assert.Len(t, patches, 2)

	for _, patch := range patches {
		assert.Equal(t, "add", patch.Operation)
		assert.Contains(t, []string{"/metadata/finalizers", "/metadata/labels"}, patch.Path)
		assert.Len(t, patch.Value, 1)
	}
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

func fixDeployment() *v1.Deployment {
	return &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deployName,
			Namespace: namespace,
		},
	}
}

func fixTargetKind() *v1alpha1.TargetKind {
	return &v1alpha1.TargetKind{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deployName,
			Namespace: namespace,
		},
		Spec: v1alpha1.TargetKindSpec{Resource: v1alpha1.Resource{
			Kind:    targetKind,
			Version: "v1",
			Group:   "apps",
		}},
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

func filterOutStatusPatch(operations []jsonpatch.JsonPatchOperation) []jsonpatch.JsonPatchOperation {
	var filtered []jsonpatch.JsonPatchOperation
	for _, op := range operations {
		if op.Path != "/status" {
			filtered = append(filtered, op)
		}
	}

	return filtered
}
