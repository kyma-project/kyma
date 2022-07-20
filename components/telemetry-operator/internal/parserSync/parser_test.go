package parserSync

import (
	"context"
	"testing"

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
	result, err := sut.SyncParsersConfigMap(context.Background(), &lp)

	require.Error(t, err)
	require.Equal(t, result, false)
}
