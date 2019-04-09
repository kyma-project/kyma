package testing

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

func NewUnstructured(apiVersion, kind string, metadata, spec, status map[string]interface{}) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": apiVersion,
			"kind":       kind,
			"metadata":   metadata,
			"spec":       spec,
			"status":     status,
		},
	}
}
