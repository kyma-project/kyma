package specification

import (
	"encoding/json"
	"fmt"
	"github.com/go-openapi/spec"
	"github.com/kyma-project/kyma/components/metadata-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/model"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/specification/minio"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	oDataSpecFormat      = "%s/$metadata"
	oDataSpecType        = "odata"
	targetSwaggerVersion = "2.0"

	specRequestTimeout = time.Duration(5 * time.Second)
)

type Service interface {
	GetSpec(id string) ([]byte, []byte, []byte, apperrors.AppError)
	RemoveSpec(id string) apperrors.AppError
	PutSpec(serviceDef *model.ServiceDefinition, gatewayUrl string) apperrors.AppError
}

type specService struct {
	minioService minio.Service
}

func NewSpecService(minioService minio.Service) Service {
	return &specService{
		minioService: minioService,
	}
}

func (svc *specService) GetSpec(id string) ([]byte, []byte, []byte, apperrors.AppError) {
	return svc.minioService.Get(id)
}

func (svc *specService) RemoveSpec(id string) apperrors.AppError {
	return svc.minioService.Remove(id)
}

func (svc *specService) PutSpec(serviceDef *model.ServiceDefinition, gatewayUrl string) apperrors.AppError {
	var apiSpec []byte
	var err apperrors.AppError

	if serviceDef.Api != nil {
		apiSpec, err = processAPISpecification(serviceDef.Api, gatewayUrl)
		if err != nil {
			return err
		}
	}

	return svc.insertSpecs(serviceDef.ID, serviceDef.Documentation, apiSpec, serviceDef.Events)
}

func (svc *specService) insertSpecs(id string, docs []byte, apiSpec []byte, events *model.Events) apperrors.AppError {
	var eventsSpec []byte

	if events != nil {
		eventsSpec = events.Spec
	}

	err := svc.minioService.Put(id, docs, apiSpec, eventsSpec)
	if err != nil {
		return apperrors.Internal("Inserting specs failed, %s", err.Error())
	}

	return nil
}

func processAPISpecification(api *model.API, gatewayUrl string) ([]byte, apperrors.AppError) {
	apiSpec := api.Spec

	var err apperrors.AppError

	if shouldFetchSpec(api) {
		apiSpec, err = fetchSpec(api)
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
	return array == nil || len(array) == 0
}

func fetchSpec(api *model.API) ([]byte, apperrors.AppError) {
	specUrl, apperr := determineSpecUrl(api)
	if apperr != nil {
		return nil, apperr
	}

	response, apperr := requestAPISpec(specUrl)
	if apperr != nil {
		return nil, apperr
	}

	apiSpec, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, apperrors.Internal("Reading API spec response body failed, %s", err.Error())
	}

	return apiSpec, nil
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

func requestAPISpec(specUrl string) (*http.Response, apperrors.AppError) {
	req, err := http.NewRequest(http.MethodGet, specUrl, nil)
	if err != nil {
		return nil, apperrors.Internal("Creating request for fetching API spec from %s failed, %s", specUrl, err.Error())
	}

	httpClient := http.Client{
		Timeout: specRequestTimeout,
	}

	response, err := httpClient.Do(req)
	if err != nil {
		return nil, apperrors.UpstreamServerCallFailed("Fetching API spec from %s failed, %s", specUrl, err.Error())
	}

	if response.StatusCode != http.StatusOK {
		return nil, apperrors.UpstreamServerCallFailed("Fetching API spec from %s failed with status %s", specUrl, response.Status)
	}

	return response, nil
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
