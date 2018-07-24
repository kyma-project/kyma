package controller

import (
	"encoding/json"

	"github.com/pkg/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	tracingAnnotationKey = "servicebindingusages.servicecatalog.kyma.cx/tracing-information"
)

//go:generate mockery -name=usageBindingAnnotationTracer -output=automock -outpkg=automock -case=underscore

type usageBindingAnnotationTracer interface {
	GetInjectedLabels(objMeta metaV1.ObjectMeta, usageName string) (map[string]string, error)
	DeleteAnnotationAboutBindingUsage(objMeta *metaV1.ObjectMeta, usageName string) error
	SetAnnotationAboutBindingUsage(objMeta *metaV1.ObjectMeta, usageName string, labels map[string]string) error
}

// usageAnnotationTracer adds information in Kubernetes resources that they have been modified
// by given ServiceBindingUsage
type usageAnnotationTracer struct{}

// GetInjectedLabels returns all labels that have been added to given resources
func (c *usageAnnotationTracer) GetInjectedLabels(objMeta metaV1.ObjectMeta, usageName string) (map[string]string, error) {
	if objMeta.Annotations == nil {
		return map[string]string{}, nil
	}

	data, found, err := c.readAnnotationData(&objMeta)
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

func (c *usageAnnotationTracer) DeleteAnnotationAboutBindingUsage(objMeta *metaV1.ObjectMeta, usageName string) error {
	data, found, err := c.readAnnotationData(objMeta)
	if err != nil {
		return errors.Wrap(err, "while reading annotation tracing data")
	}
	if !found {
		return nil
	}
	delete(data, usageName)
	err = c.saveAnnotationData(objMeta, data)
	if err != nil {
		return errors.Wrap(err, "while saving annotation tracing data")
	}
	return nil
}

// SetAnnotationAboutBindingUsage sets annotations about injected labels keys
func (c *usageAnnotationTracer) SetAnnotationAboutBindingUsage(objMeta *metaV1.ObjectMeta, usageName string, labels map[string]string) error {

	data, _, err := c.readAnnotationData(objMeta)
	if err != nil {
		return errors.Wrap(err, "while reading annotation tracing data")
	}

	info := data[usageName]
	info.InjectedLabelKeys = labels
	data[usageName] = info

	err = c.saveAnnotationData(objMeta, data)
	if err != nil {
		return errors.Wrap(err, "while saving annotation tracing data")
	}
	return nil
}

// sbuTracingInfo represents stored (in annotation) data about applied service binding usage
type sbuTracingInfo struct {
	InjectedLabelKeys map[string]string `json:"injectedLabels"`
}

func (c *usageAnnotationTracer) readAnnotationData(objMeta *metaV1.ObjectMeta) (map[string]sbuTracingInfo, bool, error) {
	value, found := objMeta.Annotations[tracingAnnotationKey]
	if !found {
		return map[string]sbuTracingInfo{}, false, nil
	}

	var data map[string]sbuTracingInfo
	err := json.Unmarshal([]byte(value), &data)
	if err != nil {
		return map[string]sbuTracingInfo{}, false, errors.Wrapf(err, "while unmarshalling annotation tracing data")
	}

	return data, true, nil
}

func (c *usageAnnotationTracer) saveAnnotationData(objMeta *metaV1.ObjectMeta, info map[string]sbuTracingInfo) error {
	bytes, err := json.Marshal(info)
	if err != nil {
		return errors.Wrapf(err, "while marshalling annotation tracing data %+v for object (%s/%s)", info, objMeta.Namespace, objMeta.Name)
	}

	objMeta.Annotations = EnsureMapIsInitiated(objMeta.Annotations)
	objMeta.Annotations[tracingAnnotationKey] = string(bytes)
	if len(info) == 0 {
		delete(objMeta.Annotations, tracingAnnotationKey)
	}
	return nil
}
