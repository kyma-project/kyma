package logparser

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/kyma-project/kyma/components/telemetry-operator/controller"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/kubernetes/mocks"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
)

func TestSyncParsersConfigMapErrorClientErrorReturnsError(t *testing.T) {
	mockClient := &mocks.Client{}
	badReqErr := errors.NewBadRequest("")
	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(badReqErr)
	sut := newSyncer(mockClient, controller.TestFluentBitK8sResources)

	lp := telemetryv1alpha1.LogParser{
		ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "fooNs"},
		Spec: telemetryv1alpha1.LogParserSpec{Parser: `
Format regex`},
	}
	result, err := sut.SyncParsersConfigMap(context.Background(), &lp)

	require.Error(t, err)
	require.Equal(t, result, false)
}

func TestSuccessfulParserConfigMap(t *testing.T) {
	var ctx context.Context
	lp := &telemetryv1alpha1.LogParser{
		ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "fooNs"},
		Spec: telemetryv1alpha1.LogParserSpec{Parser: `
Format regex`},
	}
	s := scheme.Scheme
	err := telemetryv1alpha1.AddToScheme(s)
	require.NoError(t, err)
	mockClient := fake.NewClientBuilder().WithScheme(s).WithObjects(lp).Build()
	sut := newSyncer(mockClient, controller.TestFluentBitK8sResources)

	changed, err := sut.SyncParsersConfigMap(context.Background(), lp)

	require.NoError(t, err)
	require.Equal(t, true, changed)

	var cm corev1.ConfigMap
	err = sut.Get(ctx, controller.TestFluentBitK8sResources.ParsersConfigMap, &cm)
	require.NoError(t, err)
	expectedCMData := "[PARSER]\n    Name foo\n    Format regex\n\n"
	require.Contains(t, cm.Data[parsersConfigMapKey], expectedCMData)
}
