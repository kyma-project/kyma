package controllers

import (
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit/config/builder"
	"k8s.io/apimachinery/pkg/types"
)

var (
	TestFluentBitK8sResources = fluentbit.KubernetesResources{
		DaemonSet:         types.NamespacedName{Name: "test-telemetry-fluent-bit", Namespace: "default"},
		SectionsConfigMap: types.NamespacedName{Name: "test-telemetry-fluent-bit-sections", Namespace: "default"},
		FilesConfigMap:    types.NamespacedName{Name: "test-telemetry-fluent-bit-files", Namespace: "default"},
		EnvSecret:         types.NamespacedName{Name: "test-telemetry-fluent-bit-env", Namespace: "default"},
		ParsersConfigMap:  types.NamespacedName{Name: "test-telemetry-fluent-bit-parsers", Namespace: "default"},
	}
	TestPipelineConfig = builder.PipelineConfig{
		InputTag:          "kube",
		MemoryBufferLimit: "10M",
		StorageType:       "filesystem",
		FsBufferLimit:     "1G",
	}
)
