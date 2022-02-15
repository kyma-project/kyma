package storage

import "k8s.io/apimachinery/pkg/runtime/schema"

type ResourceData struct {
	Schema      schema.GroupVersionResource
	LabelsPath  string
	LabelFields []string
}
