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
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func mustLoadManifest[T runtime.Object](scheme *runtime.Scheme, filepath string) T {
	manifest, err := os.ReadFile(filepath)
	if err != nil {
		panic(err)
	}

	decode := serializer.NewCodecFactory(scheme).UniversalDeserializer().Decode
	obj, _, err := decode(manifest, nil, nil)
	if err != nil {
		panic(err)
	}

	return obj.(T)
}

func TestPreparePipelineDryRun(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = telemetryv1alpha1.AddToScheme(scheme)

	fluentBitCm := mustLoadManifest[*corev1.ConfigMap](scheme, "testdata/given/fluent-bit-configmap.yaml")
	parser := mustLoadManifest[*telemetryv1alpha1.LogParser](scheme, "testdata/given/regex-logparser.yaml")

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(fluentBitCm, parser).Build()

	sut := fileWriterImpl{
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
			Files: []telemetryv1alpha1.FileMount{
				{
					Name:    "dummy.txt",
					Content: "dummy",
				},
			},
			Output: telemetryv1alpha1.Output{
				HTTP: telemetryv1alpha1.HTTPOutput{
					Host: telemetryv1alpha1.ValueType{Value: "127.0.0.1"},
				},
			},
		},
	}

	_, err := sut.preparePipelineDryRun(context.Background(), "testdata/actual", pipeline)
	require.NoError(t, err)

	requireEqualFiles(t, "testdata/expected/fluent-bit.conf", "testdata/actual/fluent-bit.conf")
	requireEqualFiles(t, "testdata/expected/custom_parsers.conf", "testdata/actual/custom_parsers.conf")
	requireEqualFiles(t, "testdata/expected/dynamic-parsers/parsers.conf", "testdata/actual/dynamic-parsers/parsers.conf")
	requireEqualFiles(t, "testdata/expected/dynamic/local.conf", "testdata/actual/dynamic/local.conf")
	requireEqualFiles(t, "testdata/expected/files/dummy.txt", "testdata/actual/files/dummy.txt")
}

func requireEqualFiles(t *testing.T, expectedFilePath, actualFilePath string) {
	expected, err := os.ReadFile(expectedFilePath)
	require.NoError(t, err)

	actual, err := os.ReadFile(actualFilePath)
	require.NoError(t, err)

	require.Equal(t, string(expected), string(actual))
}
