package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	requestParameters = &RequestParameters{
		Headers: &map[string][]string{
			"TestHeader": {
				"header value",
			},
		},
		QueryParameters: &map[string][]string{
			"testQueryParam": {
				"query parameter value",
			},
		},
	}

	requestParametersJsonMap = map[string][]byte{
		"headers":         []byte(`{"TestHeader":["header value"]}`),
		"queryParameters": []byte(`{"testQueryParam":["query parameter value"]}`),
	}
)

func TestMapToRequestParameters(t *testing.T) {
	t.Run("convert map to request parameters", func(t *testing.T) {
		// when
		convertedRequestParameters, err := MapToRequestParameters(requestParametersJsonMap)

		// then
		require.NoError(t, err)
		assert.Equal(t, requestParameters, convertedRequestParameters)
	})

	t.Run("return nil if request parameters are empty", func(t *testing.T) {
		// given
		jsonMap := map[string][]byte{"some key": []byte(`{"key":["value"]}`)}

		// when
		convertedRequestParameters, err := MapToRequestParameters(jsonMap)

		// then
		require.NoError(t, err)
		assert.Nil(t, convertedRequestParameters)
	})
}

func TestRequestParametersToMap(t *testing.T) {
	t.Run("convert request parameters to map", func(t *testing.T) {
		// when
		convertedJsonMap, err := RequestParametersToMap(requestParameters)

		// then
		require.NoError(t, err)
		assert.Equal(t, requestParametersJsonMap, convertedJsonMap)
	})
}
