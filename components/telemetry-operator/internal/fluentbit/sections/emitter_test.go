package sections

import (
	"github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"

	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit"
	"github.com/stretchr/testify/require"
)

func TestGenerateEmitterIncludeNamespaces(t *testing.T) {
	pipelineConfig := fluentbit.PipelineConfig{
		InputTag:          "kube",
		MemoryBufferLimit: "10M",
		StorageType:       "filesystem",
		FsBufferLimit:     "1G",
	}

	logPipeline := &v1alpha1.LogPipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-logpipeline",
		},
		Spec: v1alpha1.LogPipelineSpec{
			Input: v1alpha1.Input{Application: v1alpha1.ApplicationInput{
				Namespaces:        nil,
				ExcludeNamespaces: nil,
				Containers:        nil,
				ExcludeContainers: nil,
			}},
		},
	}

	expected := `
[FILTER]
    Name                  rewrite_tag
    Match                 kube.*
    Emitter_Name          my-logpipeline
    Emitter_Storage.type  filesystem
    Emitter_Mem_Buf_Limit 10M`

	actual := CreateEmitter(pipelineConfig, logPipeline)
	require.Equal(t, expected, actual, "Fluent Bit Emitter config is invalid")
}

func TestGenerateEmitterExcludeNamespaces(t *testing.T) {
	pipelineConfig := fluentbit.PipelineConfig{
		InputTag:          "kube",
		MemoryBufferLimit: "10M",
		StorageType:       "filesystem",
		FsBufferLimit:     "1G",
	}

	logPipeline := &v1alpha1.LogPipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-logpipeline",
		},
		Spec: v1alpha1.LogPipelineSpec{
			Input: v1alpha1.Input{Application: v1alpha1.ApplicationInput{
				Namespaces:        nil,
				ExcludeNamespaces: nil,
				Containers:        nil,
				ExcludeContainers: nil,
			}},
		},
	}

	expected := `
[FILTER]
    Name                  rewrite_tag
    Match                 kube.*
    Emitter_Name          my-logpipeline
    Emitter_Storage.type  filesystem
    Emitter_Mem_Buf_Limit 10M`

	actual := CreateEmitter(pipelineConfig, logPipeline)
	require.Equal(t, expected, actual, "Fluent Bit Emitter config is invalid")
}

func TestGenerateEmitterIncludeContainers(t *testing.T) {
	pipelineConfig := fluentbit.PipelineConfig{
		InputTag:          "kube",
		MemoryBufferLimit: "10M",
		StorageType:       "filesystem",
		FsBufferLimit:     "1G",
	}

	logPipeline := &v1alpha1.LogPipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-logpipeline",
		},
		Spec: v1alpha1.LogPipelineSpec{
			Input: v1alpha1.Input{Application: v1alpha1.ApplicationInput{
				Namespaces:        nil,
				ExcludeNamespaces: nil,
				Containers:        nil,
				ExcludeContainers: nil,
			}},
		},
	}

	expected := `
[FILTER]
    Name                  rewrite_tag
    Match                 kube.*
    Emitter_Name          my-logpipeline
    Emitter_Storage.type  filesystem
    Emitter_Mem_Buf_Limit 10M`

	actual := CreateEmitter(pipelineConfig, logPipeline)
	require.Equal(t, expected, actual, "Fluent Bit Emitter config is invalid")
}

func TestGenerateEmitterExcludeContainers(t *testing.T) {
	pipelineConfig := fluentbit.PipelineConfig{
		InputTag:          "kube",
		MemoryBufferLimit: "10M",
		StorageType:       "filesystem",
		FsBufferLimit:     "1G",
	}

	logPipeline := &v1alpha1.LogPipeline{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-logpipeline",
		},
		Spec: v1alpha1.LogPipelineSpec{
			Input: v1alpha1.Input{Application: v1alpha1.ApplicationInput{
				Namespaces:        nil,
				ExcludeNamespaces: nil,
				Containers:        nil,
				ExcludeContainers: nil,
			}},
		},
	}

	expected := `
[FILTER]
    Name                  rewrite_tag
    Match                 kube.*
    Emitter_Name          my-logpipeline
    Emitter_Storage.type  filesystem
    Emitter_Mem_Buf_Limit 10M`

	actual := CreateEmitter(pipelineConfig, logPipeline)
	require.Equal(t, expected, actual, "Fluent Bit Emitter config is invalid")
}
