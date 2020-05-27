package resource

import (
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func ToUnstructured(v interface{}) (*unstructured.Unstructured, error) {
	if v == nil {
		return nil, nil
	}

	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(v)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting resource %T to unstructured", v)
	}
	if len(u) == 0 {
		return nil, nil
	}

	return &unstructured.Unstructured{Object: u}, nil
}

func FromUnstructured(obj *unstructured.Unstructured, v interface{}) error {
	if obj == nil {
		return nil
	}

	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, v)
	if err != nil {
		return errors.Wrapf(err, "while converting unstructured to resource %T %s", v, obj.Object)
	}

	return nil
}
