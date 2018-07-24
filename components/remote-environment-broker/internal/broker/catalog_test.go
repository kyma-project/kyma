package broker_test

import (
	"fmt"
	"testing"

	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/remote-environment-broker/internal"
	"github.com/kyma-project/kyma/components/remote-environment-broker/internal/broker"
)

func TestConvertService(t *testing.T) {
	for tn, tc := range map[string]struct {
		givenSource  internal.Source
		givenService func() *internal.Service

		expectedService func() *osb.Service
	}{
		"simpleAPIBasedService": {
			givenSource: *fixSource(),
			givenService: func() *internal.Service {
				svc := fixAPIBasedService()
				svc.DisplayName = "*Service Name\ną-'#$\tÜ"
				return svc
			},
			expectedService: func() *osb.Service {
				svc := fixOsbService()
				svc.Name = "service-name-c7fe3"
				svc.Metadata["displayName"] = "*Service Name\ną-'#$\tÜ"
				return svc
			},
		},
		"emptyDisplayName": {
			givenSource: *fixSource(),
			givenService: func() *internal.Service {
				svc := fixAPIBasedService()
				svc.DisplayName = ""
				return svc
			},
			expectedService: func() *osb.Service {
				svc := fixOsbService()
				svc.Name = "c7fe3"
				svc.Metadata["displayName"] = ""
				return svc
			},
		},
		"longDisplayName": {
			givenSource: *fixSource(),
			givenService: func() *internal.Service {
				svc := fixAPIBasedService()
				svc.DisplayName = "12345678901234567890123456789012345678901234567890123456789012345678901234567890"
				return svc
			},
			expectedService: func() *osb.Service {
				svc := fixOsbService()
				svc.Name = "123456789012345678901234567890123456789012345678901234567-c7fe3"
				svc.Metadata["displayName"] = "12345678901234567890123456789012345678901234567890123456789012345678901234567890"
				return svc
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			// given
			converter := broker.NewConverter()

			// when
			result, err := converter.Convert(&tc.givenSource, tc.givenService())
			require.NoError(t, err)

			// then
			assert.Equal(t, *tc.expectedService(), result)
			assert.True(t, len(tc.expectedService().Name) < 64)
		})
	}

}

func TestFailConvertServiceWhenAccessLabelNotProvided(t *testing.T) {
	// given
	converter := broker.NewConverter()

	// when
	_, err := converter.Convert(fixSource(), &internal.Service{
		APIEntry: &internal.APIEntry{},
	})

	// then
	assert.EqualError(t, err, "cannot create binding labels: accessLabel field is required to build bindingLabels")

}

func TestIsBindableFalseForEventsBasedService(t *testing.T) {
	// given
	converter := broker.NewConverter()

	// when
	a, err := converter.Convert(fixSource(), fixEventsBasedService())

	// then
	assert.NoError(t, err)
	assert.Equal(t, a.Bindable, false)
}
func TestIsBindableTrueForAPIBasedService(t *testing.T) {
	// given
	converter := broker.NewConverter()

	// when
	a, err := converter.Convert(fixSource(), fixAPIBasedService())

	// then
	assert.NoError(t, err)
	assert.Equal(t, a.Bindable, true)
}

func fixAPIBasedService() *internal.Service {
	return &internal.Service{
		ID:                  internal.RemoteServiceID("0023-abcd-2098"),
		LongDescription:     "long description",
		DisplayName:         "service name",
		ProviderDisplayName: "HakunaMatata",
		Tags:                []string{"tag1", "tag2"},
		APIEntry: &internal.APIEntry{
			AccessLabel: "access-label-1",
			GatewayURL:  "www.gate.com",
		},
	}
}

func fixEventsBasedService() *internal.Service {
	return &internal.Service{}
}

func fixSource() *internal.Source {
	return &internal.Source{
		Environment: "prod",
		Type:        "commerce",
		Namespace:   "com.hakuna.matata",
	}
}

func fixOsbService() *osb.Service {
	return &osb.Service{
		ID:          "0023-abcd-2098",
		Description: "long description",
		Bindable:    true,
		Name:        "service-name",
		Plans: []osb.Plan{{
			Name:        "default",
			Description: "Default plan",
			ID:          fmt.Sprintf("%s-plan", "0023-abcd-2098"),
			Metadata: map[string]interface{}{
				"displayName": "Default",
			},
		}},
		Tags: []string{"tag1", "tag2"},
		Metadata: map[string]interface{}{
			"providerDisplayName":        "HakunaMatata",
			"displayName":                "service-name",
			"longDescription":            "long description",
			"remoteEnvironmentServiceId": "0023-abcd-2098",
			"source": map[string]interface{}{
				"environment": "prod",
				"type":        "commerce",
				"namespace":   "com.hakuna.matata",
			},
			"bindingLabels": map[string]string{
				"access-label-1": "true",
			},
		},
	}
}
