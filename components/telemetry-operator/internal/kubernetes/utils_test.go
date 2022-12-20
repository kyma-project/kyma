package kubernetes

import (
	"context"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"github.com/kyma-project/kyma/components/telemetry-operator/internal/kubernetes/mocks"
)

func TestGetOrCreateConfigMapError(t *testing.T) {
	mockClient := &mocks.Client{}
	badReqErr := errors.NewBadRequest("")
	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(badReqErr)

	configMapName := types.NamespacedName{Name: "some-cm", Namespace: "cm-ns"}
	_, err := GetOrCreateConfigMap(context.Background(), mockClient, configMapName)

	require.Error(t, err)
	require.Equal(t, badReqErr, err)
}

func TestGetOrCreateConfigMapGetSuccess(t *testing.T) {
	mockClient := &mocks.Client{}
	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	configMapName := types.NamespacedName{Name: "some-cm", Namespace: "cm-ns"}
	cm, err := GetOrCreateConfigMap(context.Background(), mockClient, configMapName)

	require.NoError(t, err)
	require.Equal(t, "some-cm", cm.Name)
	require.Equal(t, "cm-ns", cm.Namespace)
}

func TestGetOrCreateConfigMapCreateSuccess(t *testing.T) {
	mockClient := &mocks.Client{}
	notFoundErr := errors.NewNotFound(schema.GroupResource{}, "")
	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(notFoundErr)
	mockClient.On("Create", mock.Anything, mock.Anything).Return(nil)

	configMapName := types.NamespacedName{Name: "some-cm", Namespace: "cm-ns"}
	cm, err := GetOrCreateConfigMap(context.Background(), mockClient, configMapName)

	require.NoError(t, err)
	require.Equal(t, "some-cm", cm.Name)
	require.Equal(t, "cm-ns", cm.Namespace)
}

func TestGetOrCreateSecretError(t *testing.T) {
	mockClient := &mocks.Client{}
	badReqErr := errors.NewBadRequest("")
	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(badReqErr)

	secretName := types.NamespacedName{Name: "some-secret", Namespace: "secret-ns"}
	_, err := GetOrCreateSecret(context.Background(), mockClient, secretName)

	require.Error(t, err)
	require.Equal(t, badReqErr, err)
}

func TestGetOrCreateSecretSuccess(t *testing.T) {
	mockClient := &mocks.Client{}
	notFoundErr := errors.NewNotFound(schema.GroupResource{}, "")
	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(notFoundErr)
	mockClient.On("Create", mock.Anything, mock.Anything).Return(nil)

	secretName := types.NamespacedName{Name: "some-secret", Namespace: "secret-ns"}
	secret, err := GetOrCreateSecret(context.Background(), mockClient, secretName)

	require.NoError(t, err)
	require.Equal(t, "some-secret", secret.Name)
	require.Equal(t, "secret-ns", secret.Namespace)
}

func TestMergeKubectlAnnotations(t *testing.T) {
	existing := v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "some-deployment",
			Annotations: map[string]string{
				"kubectl.kubernetes.io/1": "1",
				"kubectl.kubernetes.io/2": "2",
				"kubectl.kubernetes.io/3": "3",
			},
		},
		Spec:   v1.DeploymentSpec{},
		Status: v1.DeploymentStatus{},
	}

	desired := v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "some-deployment",
			Annotations: map[string]string{
				"kubectl.kubernetes.io/2": "b",
				"kubectl.kubernetes.io/3": "3",
				"kubectl.kubernetes.io/4": "4",
			},
		},
		Spec:   v1.DeploymentSpec{},
		Status: v1.DeploymentStatus{},
	}

	mergeKubectlAnnotations(&desired.ObjectMeta, existing.ObjectMeta)

	require.Equal(t, len(desired.Annotations), 4)
	require.Contains(t, desired.Annotations, "kubectl.kubernetes.io/1")
	require.Contains(t, desired.Annotations, "kubectl.kubernetes.io/2")
	require.Contains(t, desired.Annotations, "kubectl.kubernetes.io/3")
	require.Contains(t, desired.Annotations, "kubectl.kubernetes.io/4")
	require.Equal(t, desired.Annotations["kubectl.kubernetes.io/1"], "1")
	require.Equal(t, desired.Annotations["kubectl.kubernetes.io/2"], "b")
	require.Equal(t, desired.Annotations["kubectl.kubernetes.io/3"], "3")
	require.Equal(t, desired.Annotations["kubectl.kubernetes.io/4"], "4")
}

func TestMergeChecksumAnnotations(t *testing.T) {
	existing := v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "some-deployment",
			Annotations: map[string]string{
				"checksum/1": "1",
				"checksum/2": "2",
				"checksum/3": "3",
			},
		},
		Spec:   v1.DeploymentSpec{},
		Status: v1.DeploymentStatus{},
	}

	desired := v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "some-deployment",
			Annotations: map[string]string{
				"checksum/2": "b",
				"checksum/3": "3",
				"checksum/4": "4",
			},
		},
		Spec:   v1.DeploymentSpec{},
		Status: v1.DeploymentStatus{},
	}

	mergeChecksumAnnotations(&desired.ObjectMeta, existing.ObjectMeta)

	require.Equal(t, len(desired.Annotations), 4)
	require.Contains(t, desired.Annotations, "checksum/1")
	require.Contains(t, desired.Annotations, "checksum/2")
	require.Contains(t, desired.Annotations, "checksum/3")
	require.Contains(t, desired.Annotations, "checksum/4")
	require.Equal(t, desired.Annotations["checksum/1"], "1")
	require.Equal(t, desired.Annotations["checksum/2"], "b")
	require.Equal(t, desired.Annotations["checksum/3"], "3")
	require.Equal(t, desired.Annotations["checksum/4"], "4")
}
