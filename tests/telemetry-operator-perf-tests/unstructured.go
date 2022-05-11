package main

import (
	"errors"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func toUnstructured(logPipelineYAML []byte) (*unstructured.Unstructured, error) {
	var logPipeline map[string]interface{}
	if err := yaml.Unmarshal(logPipelineYAML, &logPipeline); err != nil {
		return nil, err
	}
	return &unstructured.Unstructured{Object: logPipeline}, nil
}

func name(obj *unstructured.Unstructured) (string, error) {
	name, found, err := unstructured.NestedString(obj.Object, "metadata", "name")
	if err != nil {
		return "", err
	}
	if !found {
		return "", errors.New("name not found")
	}
	return name, nil
}

func hasRunningCondition(logPipeline *unstructured.Unstructured) (bool, error) {
	status, found, err := unstructured.NestedMap(logPipeline.Object, "status")
	if err != nil {
		return false, err
	}
	if !found {
		return false, nil
	}

	conditions, found, err := unstructured.NestedSlice(status, "conditions")
	if err != nil {
		return false, err
	}
	if !found {
		return false, nil
	}

	for _, cond := range conditions {
		condType := cond.(map[string]interface{})["type"]
		if condType == "Running" {
			return true, nil
		}
	}

	return false, nil
}
