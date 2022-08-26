package fluentbit

import "k8s.io/apimachinery/pkg/types"

type KubernetesResources struct {
	DaemonSet         types.NamespacedName
	SectionsConfigMap types.NamespacedName
	FilesConfigMap    types.NamespacedName
	ParsersConfigMap  types.NamespacedName
	EnvSecret         types.NamespacedName
}
