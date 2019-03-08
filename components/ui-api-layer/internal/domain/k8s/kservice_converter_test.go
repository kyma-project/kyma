package k8s

import (
	"testing"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	_assert "github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestServiceConverter_ToGQL(t *testing.T) {

	assert := _assert.New(t)

	t.Run("Nil", func(t *testing.T) {
		converter := &kserviceConverter{}
		result := converter.ToGQL(nil)
		assert.Nil(result)
	})

	t.Run("Empty", func(t *testing.T) {
		converter := &kserviceConverter{}
		expected := &gqlschema.KService{}
		result := converter.ToGQL(&v1.Service{})
		assert.Equal(expected, result)
	})

	t.Run("Success", func(t *testing.T) {
		converter := &kserviceConverter{}
		in := v1.Service{
			Spec: v1.ServiceSpec{
				Ports: []v1.ServicePort{
					{
						Name:       "test",
						Protocol:   v1.ProtocolTCP,
						Port:       1,
						TargetPort: intstr.FromInt(2),
						NodePort:   3,
					},
				},
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:              "exampleName",
				CreationTimestamp: metav1.Time{},
				Labels: map[string]string{
					"exampleKey":  "exampleValue",
					"exampleKey2": "exampleValue2",
				},
			},
			Status: v1.ServiceStatus{
				LoadBalancer: v1.LoadBalancerStatus{
					Ingress: []v1.LoadBalancerIngress{
						{
							IP:       "123.123.123.123",
							Hostname: "test",
						},
					},
				},
			},
		}
		expected := gqlschema.KService{
			Name: "exampleName",
			Labels: map[string]string{
				"exampleKey":  "exampleValue",
				"exampleKey2": "exampleValue2",
			},
			Ports: []gqlschema.ServicePort{
				{
					Name:            "test",
					ServiceProtocol: gqlschema.ServiceProtocolTcp,
					Port:            1,
					NodePort:        3,
					TargetPort:      2,
				},
			},
			Status: gqlschema.ServiceStatus{
				LoadBalancer: gqlschema.LoadBalancerStatus{
					Ingress: []gqlschema.LoadBalancerIngress{
						{
							IP:       "123.123.123.123",
							HostName: "test",
						},
					},
				},
			},
		}
		result := converter.ToGQL(&in)
		assert.Equal(&expected, result)
	})
}

func TestKServiceConverter_ToGQLs(t *testing.T) {

	assert := _assert.New(t)

	t.Run("Success", func(t *testing.T) {
		converter := kserviceConverter{}
		expectedName := "exampleName"
		in := []*v1.Service{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: expectedName,
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "exampleName2",
				},
			},
		}
		result := converter.ToGQLs(in)
		assert.Len(result, 2)
		assert.Equal(expectedName, result[0].Name)
	})

	t.Run("Empty", func(t *testing.T) {
		converter := kserviceConverter{}
		var in []*v1.Service
		result := converter.ToGQLs(in)
		assert.Empty(result)
	})

	t.Run("With nil", func(t *testing.T) {
		converter := kserviceConverter{}
		expectedName := "exampleName"
		in := []*v1.Service{
			nil,
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: expectedName,
				},
			},
			nil,
		}
		result := converter.ToGQLs(in)
		assert.Len(result, 1)
		assert.Equal(expectedName, result[0].Name)
	})
}

func TestKServiceConverter_toGQLSchemaServicePort(t *testing.T) {

	assert := _assert.New(t)

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
			ServiceProtocol: gqlschema.ServiceProtocolUdp,
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

func TestKServiceConverter_toGQLSchemaServiceProtocol(t *testing.T) {

	assert := _assert.New(t)

	tests := []struct {
		protocol v1.Protocol
		expected gqlschema.ServiceProtocol
	}{
		{
			protocol: v1.ProtocolTCP,
			expected: gqlschema.ServiceProtocolTcp,
		},
		{
			protocol: v1.ProtocolUDP,
			expected: gqlschema.ServiceProtocolUdp,
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
