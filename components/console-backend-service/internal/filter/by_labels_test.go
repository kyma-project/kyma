package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFilterBy(t *testing.T) {
	namespace := "namespace"
	labels := []string{
		"serving.knative.dev/revision",
		"foo=barbar",
	}

	t.Run("Success", func(t *testing.T) {
		service1 := fixService("service1", namespace, map[string]string{
			"foo": "bar",
		})
		service2 := fixService("service2", namespace, map[string]string{
			"serving.knative.dev/revision":                "foo",
			"serving.knative.dev/configurationGeneration": "bar",
		})
		service3 := fixService("service3", namespace, map[string]string{
			"serving.knative.dev/revision": "foo",
		})
		service4 := fixService("service4", namespace, map[string]string{
			"foo": "barbar",
		})

		rm, err := ByLabels([]interface{}{
			service1, service2, service3, service4,
		}, labels)

		var services []*v1.Service
		for _, item := range rm {
			service, ok := item.(*v1.Service)
			assert.Equal(t, true, ok)

			services = append(services, service)
		}

		require.NoError(t, err)
		assert.Equal(t, 1, len(rm))
		assert.Equal(t, []*v1.Service{service1}, services)
	})

	t.Run("Invalid input", func(t *testing.T) {
		rm, err := ByLabels([]interface{}{1, "string", true}, labels)

		require.Error(t, err)
		assert.Empty(t, rm)
	})

	t.Run("Empty input", func(t *testing.T) {
		rm, err := ByLabels([]interface{}{}, labels)

		require.NoError(t, err)
		assert.Empty(t, rm)
	})
}

func fixService(name, namespace string, labels map[string]string) *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
	}
}
