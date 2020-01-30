package filter

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
)

func ByLabels(items []interface{}, labels []string) ([]interface{}, error) {
	serializedLabels := serializeLabels(labels)

	var filteredItems []interface{}
	for _, item := range items {
		meta, err := meta.Accessor(item)
		if err != nil {
			return nil, fmt.Errorf("while gathering meta from resource %v", item)
		}

		labels := meta.GetLabels()
		if !containsLabels(labels, serializedLabels) {
			filteredItems = append(filteredItems, item)
		}
	}
	return filteredItems, nil
}

func containsLabels(itemLabels map[string]string, labels map[string]string) bool {
	for itemLabel, itemLabelValue := range itemLabels {
		for label, labelValue := range labels {
			if itemLabel == label {
				if labelValue == "" || (labelValue != "" && labelValue == itemLabelValue) {
					return true
				}
			}
		}
	}
	return false
}

func serializeLabels(labels []string) map[string]string {
	serializedLabels := make(map[string]string)
	for _, label := range labels {
		values := strings.Split(label, "=")
		key := values[0]
		value := ""
		if len(values) == 2 {
			value = values[1]
		}
		serializedLabels[key] = value
	}
	return serializedLabels
}
