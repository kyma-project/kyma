package logparser

import (
	"context"
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"testing"
)

func TestUpdateStatus(t *testing.T) {
	t.Run("should add pending condition if fluent bit is not ready", func(t *testing.T) {
		sut := Reconciler{
			Client: nil,
			config: Config{},
		}

		sut.updateStatus(context.Background(), &telemetryv1alpha1.LogParser{})
	})
}
