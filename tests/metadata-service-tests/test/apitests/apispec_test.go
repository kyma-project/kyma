/*
 * [y] hybris Platform
 *
 * Copyright (c) 2000-2018 hybris AG
 * All rights reserved.
 *
 * This software is the confidential and proprietary information of hybris
 * ("Confidential Information"). You shall not disclose such Confidential
 * Information and shall use it only in accordance with the terms of the
 * license agreement you entered into with hybris.
 */
package apitests

import (
	"github.com/stretchr/testify/require"
	"testing"
	"github.com/kyma-project/kyma/tests/metadata-service-tests/test/testkit"
	"net/http"
	"gopkg.in/yaml.v2"
	"github.com/go-openapi/spec"
	"io/ioutil"
)

func TestApiSpec(t *testing.T) {

	config, err := testkit.ReadConfig()
	require.NoError(t, err)

	t.Run("Application Connector Metadata", func(t *testing.T) {

		t.Run("should return api spec", func(t *testing.T) {
			// given
			url := config.MetadataServiceUrl + "/api.yaml"

			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			// when
			response, err := http.DefaultClient.Do(request)
			require.NoError(t, err)

			// then
			require.Equal(t, response.StatusCode, http.StatusOK)

			rawApiSpec, err := ioutil.ReadAll(response.Body)
			require.NoError(t, err)

			var apiSpec spec.Swagger
			err = yaml.Unmarshal(rawApiSpec, &apiSpec)
			require.NoError(t, err)
		})
	})

}

