package resource

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

type Convertible interface {
	FromUnstructured(unstructured *unstructured.Unstructured) error
}

type ConvertibleList interface {
	Append() Convertible
}