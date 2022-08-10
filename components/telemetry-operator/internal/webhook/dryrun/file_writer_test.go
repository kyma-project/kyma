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
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func loadCmFromFile(filepath string) (*corev1.ConfigMap, error) {
	cmYAML, err := os.ReadFile("testdata/given/fluent-bit-cm.yaml")
	if err != nil {
		return nil, err
	}

	decode := scheme.Codecs.UniversalDeserializer().Decode
	cmRuntimeObj, _, err := decode(cmYAML, nil, nil)
	if err != nil {
		return nil, err
	}

	return cmRuntimeObj.(*corev1.ConfigMap), nil
}

func TestPreparePipelineDryRun(t *testing.T) {
	fluentBitCm, err := loadCmFromFile("testdata/given/fluent-bit-cm.yaml")
	require.NoError(t, err)

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
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(fluentBitCm, parser).Build()

	sut := fileWriter{
		client: client,
		config: &Config{
			FluentBitConfigMapName: types.NamespacedName{Name: fluentBitCm.Name},
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

	requireEqualFiles(t, "testdata/expected/fluent-bit.conf", "testdata/actual/fluent-bit.conf")
	requireEqualFiles(t, "testdata/expected/custom_parsers.conf", "testdata/actual/custom_parsers.conf")
}

func requireEqualFiles(t *testing.T, expectedFilePath, actualFilePath string) {
	expected, err := os.ReadFile(expectedFilePath)
	require.NoError(t, err)

	actual, err := os.ReadFile(actualFilePath)
	require.NoError(t, err)

	require.Equal(t, string(expected), string(actual))
}
