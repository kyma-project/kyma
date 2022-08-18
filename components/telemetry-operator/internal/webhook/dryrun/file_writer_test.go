package dryrun

import (
	"context"
	"os"
	"testing"

	"github.com/kyma-project/kyma/components/telemetry-operator/internal/configbuilder"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
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
	parser := mustLoadManifest[*telemetryv1alpha1.LogParser](scheme, "testdata/given/logparser-1.yaml")

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(fluentBitCm, parser).Build()

	sut := fileWriterImpl{
		client: client,
		config: &Config{
			FluentBitConfigMapName: types.NamespacedName{Name: fluentBitCm.Name},
			PipelineConfig: configbuilder.PipelineConfig{
				FsBufferLimit:     "1G",
				MemoryBufferLimit: "10M",
				StorageType:       "filesystem",
			},
		},
	}

	pipeline := mustLoadManifest[*telemetryv1alpha1.LogPipeline](scheme, "testdata/given/logpipeline-1.yaml")
	_, err := sut.preparePipelineDryRun(context.Background(), "testdata/actual/pipelines", pipeline)
	require.NoError(t, err)

	requireEqualFiles(t, "testdata/expected/pipelines/fluent-bit.conf", "testdata/actual/pipelines/fluent-bit.conf")
	requireEqualFiles(t, "testdata/expected/pipelines/custom_parsers.conf", "testdata/actual/pipelines/custom_parsers.conf")
	requireEqualFiles(t, "testdata/expected/pipelines/dynamic-parsers/parsers.conf", "testdata/actual/pipelines/dynamic-parsers/parsers.conf")
	requireEqualFiles(t, "testdata/expected/pipelines/dynamic/logpipeline-1.conf", "testdata/actual/pipelines/dynamic/logpipeline-1.conf")
	requireEqualFiles(t, "testdata/expected/pipelines/files/dummy.txt", "testdata/actual/pipelines/files/dummy.txt")
}

func TestPrepareParsersDryRunAddParser(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = telemetryv1alpha1.AddToScheme(scheme)
	existing := mustLoadManifest[*telemetryv1alpha1.LogParser](scheme, "testdata/given/logparser-1.yaml")
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(existing).Build()
	sut := fileWriterImpl{client: client}

	added := mustLoadManifest[*telemetryv1alpha1.LogParser](scheme, "testdata/given/logparser-2.yaml")
	_, err := sut.prepareParserDryRun(context.Background(), "testdata/actual/parsers", added)
	require.NoError(t, err)

	requireEqualFiles(t, "testdata/expected/parsers/dynamic-parsers/parsers.conf", "testdata/actual/parsers/dynamic-parsers/parsers.conf")
}

func TestPrepareParsersDryRunUpdateParser(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = telemetryv1alpha1.AddToScheme(scheme)
	existing1 := mustLoadManifest[*telemetryv1alpha1.LogParser](scheme, "testdata/given/logparser-1.yaml")
	existing2 := mustLoadManifest[*telemetryv1alpha1.LogParser](scheme, "testdata/given/logparser-2.yaml")
	existing2.Spec.Parser = `
    Format      regex
`
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(existing1, existing2).Build()
	sut := fileWriterImpl{client: client}

	updated := mustLoadManifest[*telemetryv1alpha1.LogParser](scheme, "testdata/given/logparser-2.yaml")
	_, err := sut.prepareParserDryRun(context.Background(), "testdata/actual/parsers", updated)
	require.NoError(t, err)

	requireEqualFiles(t, "testdata/expected/parsers/dynamic-parsers/parsers.conf", "testdata/actual/parsers/dynamic-parsers/parsers.conf")
}

func requireEqualFiles(t *testing.T, expectedFilePath, actualFilePath string) {
	expected, err := os.ReadFile(expectedFilePath)
	require.NoError(t, err)

	actual, err := os.ReadFile(actualFilePath)
	require.NoError(t, err)

	require.Equal(t, string(expected), string(actual))
}
