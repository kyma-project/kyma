package specification

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kyma-project/kyma/components/application-gateway/pkg/authorization"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/assetstore"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/assetstore/docstopic"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/download"

	"github.com/go-openapi/spec"
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/model"
)

const (
	oDataSpecFormat      = "%s/$metadata"
	oDataSpecType        = "odata"
	targetSwaggerVersion = "2.0"
)

type Service interface {
	GetSpec(id string) ([]byte, []byte, []byte, apperrors.AppError)
	RemoveSpec(id string) apperrors.AppError
	PutSpec(serviceDef *model.ServiceDefinition, gatewayUrl string) apperrors.AppError
}

type specService struct {
	assetStoreService assetstore.Service
	downloadClient    download.Client
}

func NewSpecService(assetStoreService assetstore.Service, specRequestTimeout int, insecureSpecDownload bool) Service {
	return &specService{
		assetStoreService: assetStoreService,
		downloadClient: download.NewClient(&http.Client{
			Timeout: time.Duration(specRequestTimeout) * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureSpecDownload},
			},
		}, authorization.NewStrategyFactory(authorization.FactoryConfiguration{OAuthClientTimeout: specRequestTimeout})),
	}
}

func (svc *specService) GetSpec(id string) ([]byte, []byte, []byte, apperrors.AppError) {
	return svc.assetStoreService.Get(id)
}

func (svc *specService) RemoveSpec(id string) apperrors.AppError {
	return svc.assetStoreService.Remove(id)
}

func (svc *specService) PutSpec(serviceDef *model.ServiceDefinition, gatewayUrl string) apperrors.AppError {
	var apiSpec []byte
	var err apperrors.AppError

	if serviceDef.Api != nil {
		apiSpec, err = svc.processAPISpecification(serviceDef.Api, gatewayUrl)
		if err != nil {
			return err
		}
	}

	apiType := toApiSpecType(serviceDef.Api)

	return svc.insertSpecs(serviceDef.ID, apiType, serviceDef.Documentation, apiSpec, serviceDef.Events)
}

func (svc *specService) insertSpecs(id string, apiType docstopic.ApiType, docs []byte, apiSpec []byte, events *model.Events) apperrors.AppError {
	var eventsSpec []byte

	if events != nil {
		eventsSpec = events.Spec
	}

	err := svc.assetStoreService.Put(id, apiType, docs, apiSpec, eventsSpec)
	if err != nil {
		return apperrors.Internal("Inserting specs failed, %s", err.Error())
	}

	return nil
}

func (svc *specService) processAPISpecification(api *model.API, gatewayUrl string) ([]byte, apperrors.AppError) {
	apiSpec := api.Spec

	var err apperrors.AppError

	if shouldFetchSpec(api) {
		apiSpec, err = svc.fetchSpec(api)
		if err != nil {
			return nil, err
		}
	}

	if shouldModifySpec(apiSpec, api.ApiType) {
		apiSpec, err = modifyAPISpec(apiSpec, gatewayUrl)
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

func toApiSpecType(api *model.API) docstopic.ApiType {
	if api == nil {
		return docstopic.NoneApiType
	}

	if strings.ToLower(api.ApiType) == oDataSpecType {
		return docstopic.ODataApiType
	}

	return docstopic.OpenApiType
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
	specUrl, apperr := determineSpecUrl(api)
	if apperr != nil {
		return nil, apperr
	}

	specificationCredentials := toSpecAuthorizationCredentials(api)

	return svc.downloadClient.Fetch(specUrl, specificationCredentials, api.SpecificationRequestParameters)
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

func modifyAPISpec(rawApiSpec []byte, gatewayUrl string) ([]byte, apperrors.AppError) {
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

	newSpec, err := updateBaseUrl(apiSpec, gatewayUrl)
	if err != nil {
		return rawApiSpec, apperrors.Internal("Updating base url failed, %s", err.Error())
	}

	modifiedSpec, err := json.Marshal(newSpec)
	if err != nil {
		return rawApiSpec, apperrors.Internal("Marshalling updated API spec failed, %s", err.Error())
	}

	return modifiedSpec, nil
}

func updateBaseUrl(apiSpec spec.Swagger, gatewayUrl string) (spec.Swagger, apperrors.AppError) {
	fullUrl, err := url.Parse(gatewayUrl)
	if err != nil {
		return spec.Swagger{}, apperrors.Internal("Failed to parse gateway URL, %s", err.Error())
	}

	apiSpec.Host = fullUrl.Hostname()
	apiSpec.BasePath = ""
	apiSpec.Schemes = []string{"http"}

	return apiSpec, nil
}
