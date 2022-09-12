package builder

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"

	"github.com/stretchr/testify/require"
)

func TestCreateRewriteTagFilterIncludeNamespaces(t *testing.T) {
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
				Namespaces: v1alpha1.InputNamespaces{
					Include: []string{"namespace1", "namespace2"},
				},
			}},
		},
	}

	expected := `[FILTER]
    name                  rewrite_tag
    match                 kube.*
    emitter_mem_buf_limit 10M
    emitter_name          logpipeline1-blabla
    emitter_storage.type  filesystem
    rule                  $kubernetes['namespace_name'] "^(namespace1|namespace2)$" logpipeline1.$TAG true

`
	actual := createRewriteTagFilterSection(logPipeline, pipelineConfig)
	require.Equal(t, expected, actual)
}

func TestCreateRewriteTagFilterExcludeNamespaces(t *testing.T) {
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
				Namespaces: v1alpha1.InputNamespaces{
					Exclude: []string{"namespace1", "namespace2"}},
			}},
		},
	}

	expected := `[FILTER]
    name                  rewrite_tag
    match                 kube.*
    emitter_mem_buf_limit 10M
    emitter_name          logpipeline1-blabla
    emitter_storage.type  filesystem
    rule                  $kubernetes['namespace_name'] "^(?!namespace1$|namespace2$).*" logpipeline1.$TAG true

`
	actual := createRewriteTagFilterSection(logPipeline, pipelineConfig)
	require.Equal(t, expected, actual)
}

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
					Include: []string{"container1", "container2"}},
			}},
		},
	}

	expected := `[FILTER]
    name                  rewrite_tag
    match                 kube.*
    emitter_mem_buf_limit 10M
    emitter_name          logpipeline1-blabla
    emitter_storage.type  filesystem
    rule                  $kubernetes['container_name'] "^(container1|container2)$" logpipeline1.$TAG true
    rule                  $kubernetes['namespace_name'] "^(?!kyma-system$|kyma-integration$|kube-system$|istio-system$).*" logpipeline1.$TAG true

`
	actual := createRewriteTagFilterSection(logPipeline, pipelineConfig)
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
					Exclude: []string{"container1", "container2"}},
			}},
		},
	}

	expected := `[FILTER]
    name                  rewrite_tag
    match                 kube.*
    emitter_mem_buf_limit 10M
    emitter_name          logpipeline1-blabla
    emitter_storage.type  filesystem
    rule                  $kubernetes['container_name'] "^(?!container1$|container2$).*" logpipeline1.$TAG true
    rule                  $kubernetes['namespace_name'] "^(?!kyma-system$|kyma-integration$|kube-system$|istio-system$).*" logpipeline1.$TAG true

`
	actual := createRewriteTagFilterSection(logPipeline, pipelineConfig)
	require.Equal(t, expected, actual)
}

func TestCreateRewriteTagFilterExcludeNamespacesAndExcludeContainers(t *testing.T) {
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
				Namespaces: v1alpha1.InputNamespaces{
					Exclude: []string{"namespace1", "namespace2"}},
				Containers: v1alpha1.InputContainers{
					Exclude: []string{"container1"}},
			}},
		},
	}

	expected := `[FILTER]
    name                  rewrite_tag
    match                 kube.*
    emitter_mem_buf_limit 10M
    emitter_name          logpipeline1-blabla
    emitter_storage.type  filesystem
    rule                  $kubernetes['container_name'] "^(?!container1$).*" logpipeline1.$TAG true
    rule                  $kubernetes['namespace_name'] "^(?!namespace1$|namespace2$).*" logpipeline1.$TAG true

`
	actual := createRewriteTagFilterSection(logPipeline, pipelineConfig)
	require.Equal(t, expected, actual)
}
