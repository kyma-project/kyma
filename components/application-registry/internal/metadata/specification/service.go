package specification

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/go-openapi/spec"
	"github.com/kyma-project/kyma/components/application-gateway/pkg/authorization"
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/model"
)

const (
	oDataSpecFormat      = "%s/$metadata"
	oDataSpecType        = "odata"
	targetSwaggerVersion = "2.0"
)

type specService struct{}

func (svc *specService) GetSpec(id string) ([]byte, []byte, []byte, apperrors.AppError) {
	return []byte(""), []byte(""), []byte(""), nil
}

func (svc *specService) RemoveSpec(id string) apperrors.AppError {
	return nil
}

func (svc *specService) PutSpec(serviceDef *model.ServiceDefinition, centralGatewayUrl string) apperrors.AppError {
	return nil
}

func (svc *specService) processAPISpecification(api *model.API, centralGatewayUrl string) ([]byte, apperrors.AppError) {
	apiSpec := api.Spec

	var err apperrors.AppError

	if shouldFetchSpec(api) {
		apiSpec, err = svc.fetchSpec(api)
		if err != nil {
			return nil, err
		}
	}

	if shouldModifySpec(apiSpec, api.ApiType) {
		apiSpec, err = modifyAPISpec(apiSpec, centralGatewayUrl)
		if err != nil {
			return nil, apperrors.Internal("Modifying API spec failed, %s", err.Error())
		}
	}

	return apiSpec, nil
}

func shouldFetchSpec(api *model.API) bool {
	return isNilOrEmpty(api.Spec) && (api.SpecificationUrl != "" || strings.ToLower(api.ApiType) == oDataSpecType)
}

func shouldModifySpec(apiSpec []byte, apiType string) bool {
	return !isNilOrEmpty(apiSpec) && strings.ToLower(apiType) != oDataSpecType
}

func isNilOrEmpty(array []byte) bool {
	return array == nil || len(array) == 0 || string(array) == "null"
}

func toSpecAuthorizationCredentials(api *model.API) *authorization.Credentials {
	if api.SpecificationCredentials != nil {
		basicCredentials := api.SpecificationCredentials.Basic

		if api.SpecificationCredentials.Basic != nil {
			return &authorization.Credentials{
				BasicAuth: &authorization.BasicAuth{
					Username: basicCredentials.Username,
					Password: basicCredentials.Password,
				},
			}
		}

		if api.SpecificationCredentials.Oauth != nil {
			oauth := api.SpecificationCredentials.Oauth

			return &authorization.Credentials{
				OAuth: &authorization.OAuth{
					ClientID:     oauth.ClientID,
					ClientSecret: oauth.ClientSecret,
					URL:          oauth.URL,
				},
			}
		}
	}

	return nil
}

func (svc *specService) fetchSpec(api *model.API) ([]byte, apperrors.AppError) {
	return []byte(""), nil
}

func determineSpecUrl(api *model.API) (string, apperrors.AppError) {
	var specUrl *url.URL
	var err error

	if api.SpecificationUrl != "" {
		specUrl, err = url.Parse(api.SpecificationUrl)
		if err != nil {
			return "", apperrors.Internal("Parsing specification url failed, %s", err.Error())
		}
	} else {
		targetUrl := strings.TrimSuffix(api.TargetUrl, "/")
		specUrl, err = url.Parse(fmt.Sprintf(oDataSpecFormat, targetUrl))
		if err != nil {
			return "", apperrors.Internal("Parsing OData specification url failed, %s", err.Error())
		}
	}

	return specUrl.String(), nil
}

func modifyAPISpec(rawApiSpec []byte, centralGatewayUrl string) ([]byte, apperrors.AppError) {
	if rawApiSpec == nil {
		return rawApiSpec, nil
	}

	var apiSpec spec.Swagger
	err := json.Unmarshal(rawApiSpec, &apiSpec)
	if err != nil {
		// API spec might have different type than JSON
		return rawApiSpec, nil
	}

	if apiSpec.Swagger != targetSwaggerVersion {
		return rawApiSpec, nil
	}

	newSpec, err := updateBaseUrl(apiSpec, centralGatewayUrl)
	if err != nil {
		return rawApiSpec, apperrors.Internal("Updating base url failed, %s", err.Error())
	}

	modifiedSpec, err := json.Marshal(newSpec)
	if err != nil {
		return rawApiSpec, apperrors.Internal("Marshalling updated API spec failed, %s", err.Error())
	}

	return modifiedSpec, nil
}

func updateBaseUrl(apiSpec spec.Swagger, centralGatewayUrl string) (spec.Swagger, apperrors.AppError) {
	fullUrl, err := url.Parse(centralGatewayUrl)
	if err != nil {
		return spec.Swagger{}, apperrors.Internal("Failed to parse central gateway URL, %s", err.Error())
	}

	apiSpec.Host = fullUrl.Host + fullUrl.Path
	apiSpec.BasePath = ""
	apiSpec.Schemes = []string{"http"}

	return apiSpec, nil
}
