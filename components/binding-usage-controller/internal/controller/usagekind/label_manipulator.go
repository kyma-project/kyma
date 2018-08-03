package usagekind

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type labelManipulator struct {
	fields []string
}

func newLabelManipulator(path string) *labelManipulator {
	return &labelManipulator{
		fields: strings.Split(path, "."),
	}
}

func (lm *labelManipulator) EnsureLabelsAreApplied(res *unstructured.Unstructured, labelsToApply map[string]string) error {
	val, err := lm.findOrCreateLabelsField(res)
	if err != nil {
		return err
	}

	if labelMap, ok := val.(map[string]interface{}); ok {
		for k, v := range labelsToApply {
			labelMap[k] = v
		}
	} else {
		return fmt.Errorf("expected type of field is map[string]string, but was %T", val)
	}

	return nil
}

func (lm *labelManipulator) EnsureLabelsAreDeleted(res *unstructured.Unstructured, labelsToDelete map[string]string) error {
	val, err := lm.findOrCreateLabelsField(res)
	if err != nil {
		return err
	}

	if labelMap, ok := val.(map[string]interface{}); ok {
		for k := range labelsToDelete {
			delete(labelMap, k)
		}
	} else {
		return fmt.Errorf("expected type of field is map[string]string, but was %T", val)
	}
	return nil
}

func (lm *labelManipulator) findOrCreateLabelsField(res *unstructured.Unstructured) (interface{}, error) {
	var val interface{} = res.Object

	for i, field := range lm.fields {
		if m, ok := val.(map[string]interface{}); i < len(lm.fields) && ok {
			val, ok = m[field]
			if !ok {
				m[field] = map[string]interface{}{}
				val = m[field]
			}
		} else {
			return nil, fmt.Errorf("accessor error: %v is of the type %T, expected map[string]interface{}", val, val)
		}
	}

	return val, nil
}
