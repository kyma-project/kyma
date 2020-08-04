package usagekind_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/service-binding-usage-controller/internal/controller/usagekind"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestNewLabelManipulator_EnsureLabelsAreDeleted(t *testing.T) {
	for tn, tc := range map[string]struct {
		Object map[string]interface{}
	}{
		"empty labels": {
			Object: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{},
				},
			},
		},
		"existing labels": {
			Object: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"key-1": "val",
					},
				},
			},
		},
		"not existing labels": {
			Object: map[string]interface{}{
				"metadata": map[string]interface{}{},
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			// given
			svc := usagekind.NewLabelManipulator("metadata.labels")
			obj := &unstructured.Unstructured{
				Object: tc.Object,
			}

			// when
			svc.EnsureLabelsAreDeleted(obj, fixLabels())

			// then
			assert.NotContains(t, obj.GetLabels(), "key-1")
		})

	}
}

func TestLabelManipulator_EnsureLabelsAreApplied(t *testing.T) {
	for tn, tc := range map[string]struct {
		Object map[string]interface{}
	}{
		"empty labels": {
			Object: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{},
				},
			},
		},
		"not empty labels": {
			Object: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"john": "smith",
					},
				},
			},
		},
		"not existing labels": {
			Object: map[string]interface{}{
				"metadata": map[string]interface{}{},
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			// given
			svc := usagekind.NewLabelManipulator("metadata.labels")
			obj := &unstructured.Unstructured{
				Object: tc.Object,
			}

			// when
			svc.EnsureLabelsAreApplied(obj, fixLabels())

			// then
			assert.Equal(t, "val-1", obj.GetLabels()["key-1"])
		})
	}
}

func fixLabels() map[string]string {
	return map[string]string{"key-1": "val-1"}
}
