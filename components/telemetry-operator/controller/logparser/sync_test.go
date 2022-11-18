package logparser

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/kyma-project/kyma/components/telemetry-operator/internal/kubernetes/mocks"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
)

var (
	testConfig = Config{
		DaemonSet:        types.NamespacedName{Name: "test-telemetry-fluent-bit", Namespace: "default"},
		ParsersConfigMap: types.NamespacedName{Name: "test-telemetry-fluent-bit-parsers", Namespace: "default"},
	}
)

func TestSyncParsersConfigMapErrorClientErrorReturnsError(t *testing.T) {
	mockClient := &mocks.Client{}
	badReqErr := errors.NewBadRequest("")
	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(badReqErr)
	sut := newSyncer(mockClient, testConfig)

	err := sut.SyncParsersConfigMap(context.Background())

	require.Error(t, err)
}

func TestSuccessfulParserConfigMap(t *testing.T) {
	var ctx context.Context

	s := scheme.Scheme
	err := telemetryv1alpha1.AddToScheme(scheme.Scheme)
	require.NoError(t, err)

	lp := &telemetryv1alpha1.LogParser{
		ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "fooNs"},
		Spec: telemetryv1alpha1.LogParserSpec{Parser: `
Format regex`},
	}
	mockClient := fake.NewClientBuilder().WithScheme(s).WithObjects(lp).Build()
	sut := newSyncer(mockClient, testConfig)

	err = sut.SyncParsersConfigMap(context.Background())

	require.NoError(t, err)

	var cm corev1.ConfigMap
	err = sut.Get(ctx, testConfig.ParsersConfigMap, &cm)
	require.NoError(t, err)
	expectedCMData := "[PARSER]\n    Name foo\n    Format regex\n\n"
	require.Contains(t, cm.Data[parsersConfigMapKey], expectedCMData)
}
