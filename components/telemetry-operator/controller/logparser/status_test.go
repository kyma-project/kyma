package logparser

import (
	"context"
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/controller/logparser/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestUpdateStatus(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = telemetryv1alpha1.AddToScheme(scheme)

	t.Run("should add pending condition if fluent bit is not ready", func(t *testing.T) {
		parserName := "parser"
		parser := &telemetryv1alpha1.LogParser{
			ObjectMeta: metav1.ObjectMeta{
				Name: parserName,
			},
		}
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(parser).Build()

		proberStub := &mocks.DaemonSetProber{}
		proberStub.On("IsReady", mock.Anything, mock.Anything).Return(false, nil)

		sut := Reconciler{
			Client: fakeClient,
			config: Config{
				DaemonSet:        types.NamespacedName{Name: "fluent-bit"},
				ParsersConfigMap: types.NamespacedName{Name: "parsers"},
			},
			prober: proberStub,
		}

		err := sut.updateStatus(context.Background(), parser.Name)
		require.NoError(t, err)

		var updatedParser telemetryv1alpha1.LogParser
		_ = fakeClient.Get(context.Background(), types.NamespacedName{Name: parserName}, &updatedParser)
		require.Len(t, updatedParser.Status.Conditions, 1)
		require.Equal(t, updatedParser.Status.Conditions[0].Type, telemetryv1alpha1.LogParserPending)
		require.Equal(t, updatedParser.Status.Conditions[0].Reason, telemetryv1alpha1.FluentBitDSNotReadyReason)
	})

	t.Run("should add running condition if fluent bit becomes ready", func(t *testing.T) {
		parserName := "parser"
		parser := &telemetryv1alpha1.LogParser{
			ObjectMeta: metav1.ObjectMeta{
				Name: parserName,
			},
			Status: telemetryv1alpha1.LogParserStatus{
				Conditions: []telemetryv1alpha1.LogParserCondition{
					{Reason: telemetryv1alpha1.FluentBitDSNotReadyReason, Type: telemetryv1alpha1.LogParserPending},
				},
			},
		}
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(parser).Build()

		proberStub := &mocks.DaemonSetProber{}
		proberStub.On("IsReady", mock.Anything, mock.Anything).Return(true, nil)

		sut := Reconciler{
			Client: fakeClient,
			config: Config{
				DaemonSet:        types.NamespacedName{Name: "fluent-bit"},
				ParsersConfigMap: types.NamespacedName{Name: "parsers"},
			},
			prober: proberStub,
		}

		err := sut.updateStatus(context.Background(), parser.Name)
		require.NoError(t, err)

		var updatedParser telemetryv1alpha1.LogParser
		_ = fakeClient.Get(context.Background(), types.NamespacedName{Name: parserName}, &updatedParser)
		require.Len(t, updatedParser.Status.Conditions, 2)
		require.Equal(t, updatedParser.Status.Conditions[0].Type, telemetryv1alpha1.LogParserPending)
		require.Equal(t, updatedParser.Status.Conditions[0].Reason, telemetryv1alpha1.FluentBitDSNotReadyReason)
		require.Equal(t, updatedParser.Status.Conditions[1].Type, telemetryv1alpha1.LogParserRunning)
		require.Equal(t, updatedParser.Status.Conditions[1].Reason, telemetryv1alpha1.FluentBitDSReadyReason)
	})

	t.Run("should reset conditions and add pending if fluent bit becomes not ready again", func(t *testing.T) {
		parserName := "parser"
		parser := &telemetryv1alpha1.LogParser{
			ObjectMeta: metav1.ObjectMeta{
				Name: parserName,
			},
			Status: telemetryv1alpha1.LogParserStatus{
				Conditions: []telemetryv1alpha1.LogParserCondition{
					{Reason: telemetryv1alpha1.FluentBitDSNotReadyReason, Type: telemetryv1alpha1.LogParserPending},
					{Reason: telemetryv1alpha1.FluentBitDSReadyReason, Type: telemetryv1alpha1.LogParserRunning},
				},
			},
		}
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(parser).Build()

		proberStub := &mocks.DaemonSetProber{}
		proberStub.On("IsReady", mock.Anything, mock.Anything).Return(false, nil)

		sut := Reconciler{
			Client: fakeClient,
			config: Config{
				DaemonSet:        types.NamespacedName{Name: "fluent-bit"},
				ParsersConfigMap: types.NamespacedName{Name: "parsers"},
			},
			prober: proberStub,
		}

		err := sut.updateStatus(context.Background(), parser.Name)
		require.NoError(t, err)

		var updatedParser telemetryv1alpha1.LogParser
		_ = fakeClient.Get(context.Background(), types.NamespacedName{Name: parserName}, &updatedParser)
		require.Len(t, updatedParser.Status.Conditions, 1)
		require.Equal(t, updatedParser.Status.Conditions[0].Type, telemetryv1alpha1.LogParserPending)
		require.Equal(t, updatedParser.Status.Conditions[0].Reason, telemetryv1alpha1.FluentBitDSNotReadyReason)
	})
}
