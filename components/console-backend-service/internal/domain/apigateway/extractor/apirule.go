package extractor

import (
	"github.com/kyma-incubator/api-gateway/api/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/apigateway/pretty"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type ApiRuleUnstructuredExtractor struct{}

func (ext ApiRuleUnstructuredExtractor) Do(obj interface{}) (*v1alpha1.APIRule, error) {
	u, err := ext.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}

	return ext.FromUnstructured(u)
}

func (ext ApiRuleUnstructuredExtractor) ToUnstructured(obj interface{}) (*unstructured.Unstructured, error) {
	if obj == nil {
		return nil, nil
	}

	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting resource %s %s to unstructured", pretty.APIRule, obj)
	}
	if len(u) == 0 {
		return nil, nil
	}

	return &unstructured.Unstructured{Object: u}, nil
}

func (ext ApiRuleUnstructuredExtractor) FromUnstructured(obj *unstructured.Unstructured) (*v1alpha1.APIRule, error) {
	if obj == nil {
		return nil, nil
	}
	var apiRule v1alpha1.APIRule
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &apiRule)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting unstructured to resource %s %s", pretty.APIRule, obj.Object)
	}

	return &apiRule, nil
}
