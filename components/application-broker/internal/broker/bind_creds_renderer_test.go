package broker

import (
	"testing"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/stretchr/testify/assert"
)

func TestBindingCredentialsRendererAPIGatewayURLKey(t *testing.T) {
	tests := map[string]struct {
		given          internal.Entry
		expectedPrefix string
	}{
		"spaces, digits and not allowed characters should be removed from name": {
			given: internal.Entry{
				APIEntry: &internal.APIEntry{
					ID:   "924e9f10-1e22-4f23-bce8-9b78b02c6f8e",
					Name: " 123  test some    123 @  #$ @3 %d \t \n  name   ",
				},
			},
			expectedPrefix: "TEST_SOME_123_D_NAME_924E9F10_1E22_4F23_BCE8_9B78B02C6F8E",
		},
		"spaces and not allowed characters should be removed from ID": {
			given: internal.Entry{
				APIEntry: &internal.APIEntry{
					ID:   "  924e# %^& 9f10-1e22-4f23-bce8-9b78b02c6f8e",
					Name: "API name value",
				},
			},
			expectedPrefix: "API_NAME_VALUE_924E9F10_1E22_4F23_BCE8_9B78B02C6F8E",
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			renderer := &BindingCredentialsRenderer{}

			// when
			gotGatewayKey := renderer.apiGatewayURLKey(tc.given)
			gotTargetKey := renderer.apiTargetURLKey(tc.given)

			// then
			assert.Equal(t, tc.expectedPrefix+"_GATEWAY_URL", gotGatewayKey)
			assert.Equal(t, tc.expectedPrefix+"_TARGET_URL", gotTargetKey)
		})
	}
}

func TestBindingCredentialsRendererAPIGatewayURL(t *testing.T) {
	// given
	var (
		givenSecretName           = "secret-name"
		givenNamespace            = "ns"
		givenGatewayBaseURLFormat = "http://%s-gateway"
		expected                  = "http://ns-gateway/secret/secret-name/api/API_NAME_924E9F10_1E22_4F23_BCE8_9B78B02C6F8E"
		givenEntry                = internal.Entry{
			APIEntry: &internal.APIEntry{
				ID:   "924e9f10-1e22-4f23-bce8-9b78b02c6f8e",
				Name: " api name",
			},
		}
	)
	renderer := &BindingCredentialsRenderer{
		gatewayBaseURLFormat: givenGatewayBaseURLFormat,
	}

	// when
	got := renderer.apiGatewayURL(givenNamespace, givenSecretName, givenEntry)

	// then
	assert.Equal(t, expected, got)
}

func TestBindingCredentialsRendererAPITargetURL(t *testing.T) {
	// given
	givenEntry := internal.Entry{
		APIEntry: &internal.APIEntry{
			Name:      " api name",
			TargetURL: "http://target.io",
			ID:        "924e9f10-1e22-4f23-bce8-9b78b02c6f8e",
		},
	}

	renderer := &BindingCredentialsRenderer{}

	// when
	got := renderer.apiTargetURL(givenEntry)

	// then
	assert.Equal(t, givenEntry.TargetURL, got)
}
