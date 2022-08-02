package fluentbit

import (
	"testing"

	"github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/require"
)

func TestGenerateEmitterIncludeNamespaces(t *testing.T) {
	pipelineConfig := PipelineConfig{
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
				Namespaces: []string{"namespace1", "namespace2"},
			}},
		},
	}

	expected := `[FILTER]
    Name                  rewrite_tag
    Match                 kube.*
    Emitter_Name          logpipeline1
    Emitter_Storage.type  filesystem
    Emitter_Mem_Buf_Limit 10M
    Rule                  $kubernetes['namespace_name'] "^(namespace1|namespace2)$" logpipeline1.$TAG true

`
	actual := CreateRewriteTagFilter(pipelineConfig, logPipeline)
	require.Equal(t, expected, actual)
}

func TestGenerateEmitterExcludeNamespaces(t *testing.T) {
	pipelineConfig := PipelineConfig{
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
				ExcludeNamespaces: []string{"namespace1", "namespace2"},
			}},
		},
	}

	expected := `[FILTER]
    Name                  rewrite_tag
    Match                 kube.*
    Emitter_Name          logpipeline1
    Emitter_Storage.type  filesystem
    Emitter_Mem_Buf_Limit 10M
    Rule                  $kubernetes['namespace_name'] "^(?!namespace1$|namespace2$).*" logpipeline1.$TAG true

`
	actual := CreateRewriteTagFilter(pipelineConfig, logPipeline)
	require.Equal(t, expected, actual)
}

func TestGenerateEmitterIncludeContainers(t *testing.T) {
	pipelineConfig := PipelineConfig{
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
				Containers: []string{"container1", "container2"},
			}},
		},
	}

	expected := `[FILTER]
    Name                  rewrite_tag
    Match                 kube.*
    Emitter_Name          logpipeline1
    Emitter_Storage.type  filesystem
    Emitter_Mem_Buf_Limit 10M
    Rule                  $kubernetes['container_name'] "^(container1|container2)$" logpipeline1.$TAG true

`
	actual := CreateRewriteTagFilter(pipelineConfig, logPipeline)
	require.Equal(t, expected, actual)
}

func TestGenerateEmitterExcludeContainers(t *testing.T) {
	pipelineConfig := PipelineConfig{
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
				ExcludeContainers: []string{"container1", "container2"},
			}},
		},
	}

	expected := `[FILTER]
    Name                  rewrite_tag
    Match                 kube.*
    Emitter_Name          logpipeline1
    Emitter_Storage.type  filesystem
    Emitter_Mem_Buf_Limit 10M
    Rule                  $kubernetes['container_name'] "^(?!container1$|container2$).*" logpipeline1.$TAG true

`
	actual := CreateRewriteTagFilter(pipelineConfig, logPipeline)
	require.Equal(t, expected, actual)
}

func TestGenerateEmitterExcludeNamespacesAndExcludeContainers(t *testing.T) {
	pipelineConfig := PipelineConfig{
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
				ExcludeNamespaces: []string{"namespace1", "namespace2"},
				ExcludeContainers: []string{"container1"},
			}},
		},
	}

	expected := `[FILTER]
    Name                  rewrite_tag
    Match                 kube.*
    Emitter_Name          logpipeline1
    Emitter_Storage.type  filesystem
    Emitter_Mem_Buf_Limit 10M
    Rule                  $kubernetes['namespace_name'] "^(?!namespace1$|namespace2$).*" logpipeline1.$TAG true
    Rule                  $kubernetes['container_name'] "^(?!container1$).*" logpipeline1.$TAG true

`
	actual := CreateRewriteTagFilter(pipelineConfig, logPipeline)
	require.Equal(t, expected, actual)
}
