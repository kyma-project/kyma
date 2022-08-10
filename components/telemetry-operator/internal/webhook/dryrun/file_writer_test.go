package dryrun

import (
	"context"
	"os"
	"testing"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestPreparePipelineDryRun(t *testing.T) {
	config, err := os.ReadFile("testdata/given/fluent-bit.conf")
	require.NoError(t, err)

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "fluent-bit",
		},
		Data: map[string]string{
			"fluent-bit.conf": string(config),
		},
	}

	parser := &telemetryv1alpha1.LogParser{
		ObjectMeta: metav1.ObjectMeta{
			Name: "regex-parser",
		},
		Spec: telemetryv1alpha1.LogParserSpec{
			Parser: `
Format regex
Regex  ^(?<user>[^ ]*) (?<pass>[^ ]*)$
Time_Key time
Time_Format %d/%b/%Y:%H:%M:%S %z
Types user:string pass:string
`,
		},
	}

	scheme := runtime.NewScheme()
	corev1.AddToScheme(scheme)
	telemetryv1alpha1.AddToScheme(scheme)
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(cm, parser).Build()

	sut := fileWriter{
		client: client,
		config: &Config{
			FluentBitConfigMapName: types.NamespacedName{Name: cm.Name},
			FluentBitBinPath:       "/usr/local/bin/fluent-bit", //TODO: local testing, remove it
			PipelineConfig: fluentbit.PipelineConfig{
				FsBufferLimit:     "1G",
				MemoryBufferLimit: "10M",
				StorageType:       "filesystem",
			},
		},
	}

	pipeline := &telemetryv1alpha1.LogPipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name: "local",
		},
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Output: telemetryv1alpha1.Output{
				HTTP: telemetryv1alpha1.HTTPOutput{
					Host: telemetryv1alpha1.ValueType{Value: "127.0.0.1"},
				},
			},
		},
	}

	_, err = sut.preparePipelineDryRun(context.Background(), "testdata/actual", pipeline)
	require.NoError(t, err)
}
