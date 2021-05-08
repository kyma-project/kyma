package appsecrets

import (
	"encoding/json"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/model"
)

const (
	requestParametersHeadersKey         = "headers"
	requestParametersQueryParametersKey = "queryParameters"
)

func MapToRequestParameters(data map[string][]byte) (*model.RequestParameters, apperrors.AppError) {
	requestParameters := &model.RequestParameters{}

	headersData := data[requestParametersHeadersKey]
	if headersData != nil {
		var headers = &map[string][]string{}
		err := json.Unmarshal(headersData, headers)
		if err != nil {
			return nil, apperrors.Internal("Failed to unmarshal headers, %v", err)
		}

		requestParameters.Headers = headers
	}

	queryParamsData := data[requestParametersQueryParametersKey]
	if queryParamsData != nil {
		var queryParameters = &map[string][]string{}
		err := json.Unmarshal(queryParamsData, queryParameters)
		if err != nil {
			return nil, apperrors.Internal("Failed to unmarshal query parameters, %v", err)
		}

		requestParameters.QueryParameters = queryParameters
	}

	if requestParameters.Headers == nil && requestParameters.QueryParameters == nil {
		return nil, nil
	}

	return requestParameters, nil
}

func RequestParametersToMap(requestParameters *model.RequestParameters) (map[string][]byte, apperrors.AppError) {
	data := make(map[string][]byte)
	if requestParameters == nil {
		return map[string][]byte{}, nil
	}
	if requestParameters.Headers != nil {
		headers, err := json.Marshal(requestParameters.Headers)
		if err != nil {
			return map[string][]byte{}, apperrors.Internal("Failed to marshall headers from request parameters: %v", err)
		}
		data[requestParametersHeadersKey] = headers
	}
	if requestParameters.QueryParameters != nil {
		queryParameters, err := json.Marshal(requestParameters.QueryParameters)
		if err != nil {
			return map[string][]byte{}, apperrors.Internal("Failed to marshall query parameters from request parameters: %v", err)
		}
		data[requestParametersQueryParametersKey] = queryParameters
	}
	return data, nil
}
