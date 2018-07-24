package controller

import (
	"encoding/json"

	"fmt"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type genericUsageBindingAnnotationTracer interface {
	GetInjectedLabels(res *unstructured.Unstructured, usageName string) (map[string]string, error)
	DeleteAnnotationAboutBindingUsage(res *unstructured.Unstructured, usageName string) error
	SetAnnotationAboutBindingUsage(res *unstructured.Unstructured, usageName string, labels map[string]string) error
}

// genericUsageAnnotationTracer adds information in any kind of Kubernetes resources that they have been modified
// by given ServiceBindingUsage
type genericUsageAnnotationTracer struct{}

// GetInjectedLabels returns all labels that have been added to given resources
func (c *genericUsageAnnotationTracer) GetInjectedLabels(res *unstructured.Unstructured, usageName string) (map[string]string, error) {
	data, found, err := c.readAnnotationData(res)
	if err != nil {
		return map[string]string{}, errors.Wrapf(err, "while reading binding usage annotation tracing data")
	}
	if !found {
		return map[string]string{}, nil
	}

	info, found := data[usageName]
	if !found {
		return map[string]string{}, nil
	}

	return info.InjectedLabelKeys, nil
}

func (c *genericUsageAnnotationTracer) DeleteAnnotationAboutBindingUsage(res *unstructured.Unstructured, usageName string) error {
	data, found, err := c.readAnnotationData(res)
	if err != nil {
		return errors.Wrap(err, "while reading annotation tracing data")
	}
	if !found {
		return nil
	}
	delete(data, usageName)
	err = c.saveAnnotationData(res, data)
	if err != nil {
		return errors.Wrap(err, "while saving annotation tracing data")
	}
	return nil
}

// SetAnnotationAboutBindingUsage sets annotations about injected labels keys
func (c *genericUsageAnnotationTracer) SetAnnotationAboutBindingUsage(res *unstructured.Unstructured, usageName string, labels map[string]string) error {
	data, _, err := c.readAnnotationData(res)
	if err != nil {
		return errors.Wrap(err, "while reading annotation tracing data")
	}

	info := data[usageName]
	info.InjectedLabelKeys = labels
	data[usageName] = info

	err = c.saveAnnotationData(res, data)
	if err != nil {
		return errors.Wrap(err, "while saving annotation tracing data")
	}
	return nil
}

func (c *genericUsageAnnotationTracer) readAnnotationData(res *unstructured.Unstructured) (map[string]sbuTracingInfo, bool, error) {
	annotations, err := c.extractAnnotations(res)
	if err != nil {
		return map[string]sbuTracingInfo{}, false, fmt.Errorf("while extracting annotations")
	}

	value, found := annotations[tracingAnnotationKey]
	if !found {
		return map[string]sbuTracingInfo{}, false, nil
	}
	valueAsString, ok := value.(string)
	if !ok {
		return map[string]sbuTracingInfo{}, false, fmt.Errorf("incorrect annotations, expect string but was %T", valueAsString)
	}

	var data map[string]sbuTracingInfo
	err = json.Unmarshal([]byte(valueAsString), &data)
	if err != nil {
		return map[string]sbuTracingInfo{}, false, errors.Wrapf(err, "while unmarshalling annotation tracing data")
	}

	return data, true, nil
}

func (c *genericUsageAnnotationTracer) saveAnnotationData(res *unstructured.Unstructured, info map[string]sbuTracingInfo) error {
	bytes, err := json.Marshal(info)
	if err != nil {
		return errors.Wrapf(err, "while marshalling annotation tracing data %+v for object (%s/%s)", info, res.GetNamespace(), res.GetName())
	}

	annotations, err := c.extractAnnotations(res)
	if err != nil {
		return fmt.Errorf("while extracting annotations")
	}

	metadataAsMap, ok := res.Object["metadata"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("incorrect metadata, expect map[string]interface{} but was %T", res.Object["metadata"])
	}

	metadataAsMap["annotations"] = annotations
	annotations[tracingAnnotationKey] = string(bytes)

	if len(info) == 0 {
		delete(annotations, tracingAnnotationKey)
	}
	return nil
}

func (c *genericUsageAnnotationTracer) extractAnnotations(res *unstructured.Unstructured) (map[string]interface{}, error) {
	metadata, exists := res.Object["metadata"]
	if !exists {
		return map[string]interface{}{}, nil
	}
	metadataAsMap, ok := metadata.(map[string]interface{})
	if !ok {
		return map[string]interface{}{}, fmt.Errorf("incorrect metadata, expect map[string]interface{} but was %T", metadata)
	}

	annotations, exists := metadataAsMap["annotations"]
	if !exists {
		return map[string]interface{}{}, nil
	}

	annotationsAsMap, ok := annotations.(map[string]interface{})
	if !ok {
		return map[string]interface{}{}, fmt.Errorf("incorrect annotations, expect map[string]interface{} but was %T", annotations)
	}

	return annotationsAsMap, nil
}
