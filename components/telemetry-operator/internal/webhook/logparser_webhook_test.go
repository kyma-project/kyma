package webhook

import (
	"context"
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fs"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/validation"
	validationmocks "github.com/kyma-project/kyma/components/telemetry-operator/internal/validation/mocks"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/metrics"

	"testing"
)

const (
	FluentBitConfigMap = "telemetry-fluent-bit"
	fluentBitNs        = "default"
)

var pipelineConfig = fluentbit.PipelineConfig{
	InputTag:          "kube",
	MemoryBufferLimit: "10M",
	StorageType:       "filesystem",
	FsBufferLimit:     "1G",
}

var restartsTotal = prometheus.NewCounter(prometheus.CounterOpts{
	Name: "telemetry_fluentbit_restarts_total",
	Help: "Number of triggered Fluent Bit restarts",
})

func TestNewLogPipeline(t *testing.T) {
	metrics.Registry.MustRegister(restartsTotal)
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
	configValidatorMock = &validationmocks.ConfigValidator{}
	lpv := NewLogParserValidator(mockClient,
		FluentBitConfigMap,
		fluentBitNs,
		validation.NewParserValidator(),
		pipelineConfig,
		validation.NewConfigValidator("/tmp", "fluent-bit/lib"),
		fs.NewWrapper(),
		restartsTotal)
	err = lpv.validateLogParser(ctx, lp)

	require.NoError(t, err)
}
