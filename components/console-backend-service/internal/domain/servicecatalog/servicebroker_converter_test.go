package servicecatalog

import (
	"testing"
	"time"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestServiceBrokerConverter_ToGQL(t *testing.T) {
	t.Run("All properties are given", func(t *testing.T) {
		converter := serviceBrokerConverter{}
		var zeroTimeStamp time.Time
		labels := map[string]string{
			"label1": "labelValue1",
			"label2": "labelValue2",
		}

		item := fixServiceBroker()

		expected := gqlschema.ServiceBroker{
			Name:              "exampleName",
			CreationTimestamp: zeroTimeStamp,
			Labels:            labels,
			URL:               "ExampleURL",
			Status: &gqlschema.ServiceBrokerStatus{
				Ready:   true,
				Reason:  "ExampleReason",
				Message: "ExampleMessage",
			},
		}

		result, err := converter.ToGQL(item)
		require.NoError(t, err)
		assert.Equal(t, &expected, result)
	})

	t.Run("Empty", func(t *testing.T) {
		converter := &serviceBrokerConverter{}
		_, err := converter.ToGQL(&v1beta1.ServiceBroker{})
		require.NoError(t, err)
	})

	t.Run("Empty auth info", func(t *testing.T) {
		converter := &serviceBrokerConverter{}
		_, err := converter.ToGQL(&v1beta1.ServiceBroker{
			Spec: v1beta1.ServiceBrokerSpec{
				AuthInfo: &v1beta1.ServiceBrokerAuthInfo{},
			},
		})
		require.NoError(t, err)
	})

	t.Run("Empty basic and bearer", func(t *testing.T) {
		converter := &serviceBrokerConverter{}
		_, err := converter.ToGQL(&v1beta1.ServiceBroker{
			Spec: v1beta1.ServiceBrokerSpec{
				AuthInfo: &v1beta1.ServiceBrokerAuthInfo{
					Basic:  &v1beta1.BasicAuthConfig{},
					Bearer: &v1beta1.BearerTokenAuthConfig{},
				},
			},
		})
		require.NoError(t, err)
	})

	t.Run("Nil", func(t *testing.T) {
		converter := &serviceBrokerConverter{}
		item, err := converter.ToGQL(nil)

		require.NoError(t, err)
		assert.Nil(t, item)
	})
}

func TestServiceBrokerConverter_ToGQLs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		brokers := []*v1beta1.ServiceBroker{
			fixServiceBroker(),
			fixServiceBroker(),
		}

		converter := serviceBrokerConverter{}
		result, err := converter.ToGQLs(brokers)

		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "exampleName", result[0].Name)
	})

	t.Run("Empty", func(t *testing.T) {
		var brokers []*v1beta1.ServiceBroker

		converter := serviceBrokerConverter{}
		result, err := converter.ToGQLs(brokers)

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("With nil", func(t *testing.T) {
		brokers := []*v1beta1.ServiceBroker{
			nil,
			fixServiceBroker(),
			nil,
		}

		converter := serviceBrokerConverter{}
		result, err := converter.ToGQLs(brokers)

		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "exampleName", result[0].Name)
	})
}

func fixServiceBroker() *v1beta1.ServiceBroker {
	var mockTimeStamp metav1.Time
	labels := map[string]string{
		"label1": "labelValue1",
		"label2": "labelValue2",
	}

	return &v1beta1.ServiceBroker{

		ObjectMeta: metav1.ObjectMeta{
			Name:              "exampleName",
			CreationTimestamp: mockTimeStamp,
			Labels:            labels,
		},
		Spec: v1beta1.ServiceBrokerSpec{
			CommonServiceBrokerSpec: v1beta1.CommonServiceBrokerSpec{
				URL: "ExampleURL",
			},
		},
		Status: v1beta1.ServiceBrokerStatus{
			CommonServiceBrokerStatus: v1beta1.CommonServiceBrokerStatus{
				Conditions: []v1beta1.ServiceBrokerCondition{
					{
						Type:               v1beta1.ServiceBrokerConditionType("Ready"),
						Status:             v1beta1.ConditionStatus("True"),
						LastTransitionTime: mockTimeStamp,
						Reason:             "ExampleReason",
						Message:            "ExampleMessage",
					},
				},
			},
		},
	}
}
