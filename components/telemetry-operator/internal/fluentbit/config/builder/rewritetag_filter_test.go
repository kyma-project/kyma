package builder

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"

	"github.com/stretchr/testify/require"
)

func TestCreateRewriteTagFilterIncludeContainers(t *testing.T) {
	pipelineConfig := PipelineDefaults{
		InputTag:          "kube",
		MemoryBufferLimit: "10M",
		StorageType:       "filesystem",
	}

	logPipeline := &v1alpha1.LogPipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name: "logpipeline1",
		},
		Spec: v1alpha1.LogPipelineSpec{
			Input: v1alpha1.Input{Application: v1alpha1.ApplicationInput{
				Containers: v1alpha1.InputContainers{
					Include: []string{"container1", "container2"}}}}}}

	expected := `[FILTER]
    name                  rewrite_tag
    match                 kube.*
    emitter_mem_buf_limit 10M
    emitter_name          logpipeline1
    emitter_storage.type  filesystem
    rule                  $kubernetes['container_name'] "^(container1|container2)$" logpipeline1.$TAG true

`
	actual := createRewriteTagFilter(logPipeline, pipelineConfig)
	require.Equal(t, expected, actual)
}

func TestCreateRewriteTagFilterExcludeContainers(t *testing.T) {
	pipelineConfig := PipelineDefaults{
		InputTag:          "kube",
		MemoryBufferLimit: "10M",
		StorageType:       "filesystem",
	}

	logPipeline := &v1alpha1.LogPipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name: "logpipeline1",
		},
		Spec: v1alpha1.LogPipelineSpec{
			Input: v1alpha1.Input{Application: v1alpha1.ApplicationInput{
				Containers: v1alpha1.InputContainers{
					Exclude: []string{"container1", "container2"}}}}}}

	expected := `[FILTER]
    name                  rewrite_tag
    match                 kube.*
    emitter_mem_buf_limit 10M
    emitter_name          logpipeline1
    emitter_storage.type  filesystem
    rule                  $kubernetes['container_name'] "^(?!container1$|container2$).*" logpipeline1.$TAG true

`
	actual := createRewriteTagFilter(logPipeline, pipelineConfig)
	require.Equal(t, expected, actual)
}

func TestCreateRewriteTagFilterWithCustomOutput(t *testing.T) {
	pipelineConfig := PipelineDefaults{
		InputTag:          "kube",
		MemoryBufferLimit: "10M",
		StorageType:       "filesystem",
	}

	logPipeline := &v1alpha1.LogPipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name: "logpipeline1",
		},
		Spec: v1alpha1.LogPipelineSpec{
			Output: v1alpha1.Output{
				Custom: `
    name stdout`,
			},
		},
	}

	expected := `[FILTER]
    name                  rewrite_tag
    match                 kube.*
    emitter_mem_buf_limit 10M
    emitter_name          logpipeline1-stdout
    emitter_storage.type  filesystem
    rule                  $kubernetes['namespace_name'] "^(?!kyma-system$|kyma-integration$|kube-system$|istio-system$|compass-system$).*" logpipeline1.$TAG true

`
	actual := createRewriteTagFilterSection(logPipeline, pipelineConfig)
	require.Equal(t, expected, actual)
}
