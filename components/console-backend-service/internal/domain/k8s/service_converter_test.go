package k8s

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestServiceConverter_ToGQL(t *testing.T) {
	assert := assert.New(t)

	t.Run("Nil", func(t *testing.T) {
		converter := &serviceConverter{}
		result, err := converter.ToGQL(nil)
		require.NoError(t, err)
		assert.Nil(result)
	})

	t.Run("Empty", func(t *testing.T) {
		converter := &serviceConverter{}
		emptyServiceJSON, err := converter.serviceToGQLJSON(&v1.Service{})
		expected := &gqlschema.Service{
			JSON: emptyServiceJSON,
			Status: &gqlschema.ServiceStatus{
				LoadBalancer: &gqlschema.LoadBalancerStatus{},
			},
			Labels: make(map[string]string),
		}
		result, err := converter.ToGQL(&v1.Service{})
		require.NoError(t, err)
		assert.Equal(expected, result)
	})

	t.Run("Success", func(t *testing.T) {
		converter := &serviceConverter{}
		name := "test_name"
		namespace := "test_namespace"
		in := fixService(name, namespace)
		expectedJSON, err := converter.serviceToGQLJSON(in)
		expected := gqlschema.Service{
			Name: name,
			Labels: map[string]string{
				"exampleKey":  "exampleValue",
				"exampleKey2": "exampleValue2",
			},
			Ports: []*gqlschema.ServicePort{
				{
					Name:            "test",
					ServiceProtocol: gqlschema.ServiceProtocolTCP,
					Port:            1,
					NodePort:        3,
					TargetPort:      2,
				},
			},
			Status: &gqlschema.ServiceStatus{
				LoadBalancer: &gqlschema.LoadBalancerStatus{
					Ingress: []*gqlschema.LoadBalancerIngress{
						{
							IP:       "123.123.123.123",
							HostName: "test",
						},
					},
				},
			},
			JSON: expectedJSON,
		}
		result, err := converter.ToGQL(in)
		require.NoError(t, err)
		assert.Equal(&expected, result)
	})
}

func TestServiceConverter_ToGQLs(t *testing.T) {
	assert := assert.New(t)

	t.Run("Success", func(t *testing.T) {
		converter := serviceConverter{}
		expectedName := "exampleName"
		in := []*v1.Service{
			fixServiceWithName(expectedName, ""),
			fixServiceWithName("exampleName2", ""),
		}
		result, err := converter.ToGQLs(in)
		require.NoError(t, err)
		assert.Len(result, 2)
		assert.Equal(expectedName, result[0].Name)
	})

	t.Run("Empty", func(t *testing.T) {
		converter := serviceConverter{}
		var in []*v1.Service
		result, err := converter.ToGQLs(in)
		require.NoError(t, err)
		assert.Empty(result)
	})

	t.Run("With nil", func(t *testing.T) {
		converter := serviceConverter{}
		expectedName := "exampleName"
		in := []*v1.Service{
			nil,
			fixServiceWithName(expectedName, ""),
			nil,
		}
		result, err := converter.ToGQLs(in)
		require.NoError(t, err)
		assert.Len(result, 1)
		assert.Equal(expectedName, result[0].Name)
	})
}

func TestServiceConverter_toGQLSchemaServicePort(t *testing.T) {
	assert := assert.New(t)

	t.Run("Success", func(t *testing.T) {
		actual := toGQLSchemaServicePort(&v1.ServicePort{
			Name:       "testName",
			Protocol:   v1.ProtocolUDP,
			Port:       1,
			TargetPort: intstr.FromInt(2),
			NodePort:   3,
		})
		assert.Equal(&gqlschema.ServicePort{
			Name:            "testName",
			ServiceProtocol: gqlschema.ServiceProtocolUDP,
			NodePort:        3,
			TargetPort:      2,
			Port:            1,
		}, actual)
	})

	t.Run("Empty", func(t *testing.T) {
		toGQLSchemaServicePort(&v1.ServicePort{})
	})

	t.Run("Nil", func(t *testing.T) {
		result := toGQLSchemaServicePort(nil)
		assert.Nil(result)
	})
}

func TestServiceConverter_toGQLSchemaServiceProtocol(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		protocol v1.Protocol
		expected gqlschema.ServiceProtocol
	}{
		{
			protocol: v1.ProtocolTCP,
			expected: gqlschema.ServiceProtocolTCP,
		},
		{
			protocol: v1.ProtocolUDP,
			expected: gqlschema.ServiceProtocolUDP,
		},
		{
			protocol: v1.Protocol("FTP"),
			expected: gqlschema.ServiceProtocolUnknown,
		},
	}
	for _, test := range tests {
		t.Run(string(test.protocol), func(t *testing.T) {
			actual := toGQLSchemaServiceProtocol(&test.protocol)
			assert.Equal(test.expected, actual)
		})
	}
}

func TestServiceConverter_GQLJSONToService(t *testing.T) {
	assert := assert.New(t)
	emptyService := v1.Service{}
	converter := &serviceConverter{}

	t.Run("Nil", func(t *testing.T) {
		result, err := converter.GQLJSONToService(nil)
		assert.Nil(err)
		assert.Equal(emptyService, result)
	})

	t.Run("Success", func(t *testing.T) {
		originalService := fixService("originalService", "namespace")
		gqlJSON, err := converter.serviceToGQLJSON(originalService)
		assert.Nil(err)
		assert.NotEmpty(gqlJSON)
		convertedService, err := converter.GQLJSONToService(gqlJSON)
		assert.Equal(originalService, &convertedService)
	})
}

func fixServiceWithName(name, namespace string) *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}

func fixService(name, namespace string) *v1.Service {
	result := fixServiceWithName(name, namespace)
	result.Spec = v1.ServiceSpec{
		Ports: []v1.ServicePort{
			{
				Name:       "test",
				Protocol:   v1.ProtocolTCP,
				Port:       1,
				TargetPort: intstr.FromInt(2),
				NodePort:   3,
			},
		},
	}
	result.ObjectMeta.CreationTimestamp = metav1.Time{}
	result.Labels = map[string]string{
		"exampleKey":  "exampleValue",
		"exampleKey2": "exampleValue2",
	}
	result.Status = v1.ServiceStatus{
		LoadBalancer: v1.LoadBalancerStatus{
			Ingress: []v1.LoadBalancerIngress{
				{
					IP:       "123.123.123.123",
					Hostname: "test",
				},
			},
		},
	}
	return result
}
