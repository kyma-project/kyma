package parserSync

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/sync/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var daemonSetConfig = FluentBitDaemonSetConfig{
	FluentBitDaemonSetName:    types.NamespacedName{Name: "telemetry-fluent-bit", Namespace: "cm-ns"},
	FluentBitParsersConfigMap: types.NamespacedName{Name: "telemetry-fluent-bit", Namespace: "cm-ns"},
}

func TestSyncParsersConfigMapErrorClientErrorReturnsError(t *testing.T) {
	mockClient := &mocks.Client{}
	badReqErr := errors.NewBadRequest("")
	mockClient.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(badReqErr)
	sut := NewLogParserSyncer(mockClient, daemonSetConfig)

	lp := telemetryv1alpha1.LogParser{
		ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "fooNs"},
		Spec: telemetryv1alpha1.LogParserSpec{Parser: `
Format regex`},
	}
	res, err := sut.SyncParsersConfigMap(context.Background(), &lp)

	var syncRes result
	require.Error(t, err)
	require.Equal(t, res, syncRes)
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
	sut := NewLogParserSyncer(mockClient, daemonSetConfig)

	changed, err := sut.SyncParsersConfigMap(context.Background(), lp)
	var expectedResult result
	expectedResult.ConfigMapUpdated = true
	expectedResult.CrUpdated = true

	require.NoError(t, err)
	require.Equal(t, expectedResult, changed)

	var cm corev1.ConfigMap
	err = sut.Get(ctx, daemonSetConfig.FluentBitParsersConfigMap, &cm)
	require.NoError(t, err)
	expectedCMData := "[PARSER]\n    Format regex\n    Name foo\n\n"
	require.Contains(t, cm.Data[parsersConfigMapKey], expectedCMData)
}
