package specification

import (
	"encoding/json"
	"fmt"
	"github.com/go-openapi/spec"
	"github.com/kyma-project/kyma/components/metadata-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/serviceapi"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata/specification/minio"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	oDataSpecFormat      = "%s/$metadata"
	oDataSpecType        = "odata"
	targetSwaggerVersion = "2.0"
)

type SpecService interface {
	GetSpec(id string) ([]byte, []byte, []byte, apperrors.AppError)
	RemoveSpec(id string) apperrors.AppError
	SaveServiceSpecs(specData SpecData) apperrors.AppError
}

type specService struct {
	minioService minio.Service
}

func NewSpecService(minioService minio.Service) SpecService {
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

func (svc *specService) SaveServiceSpecs(specData SpecData) apperrors.AppError {
	var apiSpec []byte
	var err apperrors.AppError

	if specData.API != nil {
		apiSpec, err = processAPISpecification(specData.API, specData.GatewayUrl)
		if err != nil {
			return err
		}
	}

	return svc.insertSpecs(specData.Id, specData.Docs, apiSpec, specData.Events)
}

func (svc *specService) insertSpecs(id string, docs []byte, apiSpec []byte, events *Events) apperrors.AppError {
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

func processAPISpecification(api *serviceapi.API, gatewayUrl string) ([]byte, apperrors.AppError) {
	apiSpec := api.Spec

	var err apperrors.AppError

	if isNilOrEmpty(apiSpec) {
		apiSpec, err = fetchSpec(api)
		if err != nil {
			return nil, err
		}
	}

	apiSpec, err = modifyAPISpec(apiSpec, gatewayUrl)
	if err != nil {
		return nil, apperrors.Internal("Modifying API spec failed, %s", err.Error())
	}

	str := string(apiSpec)
	fmt.Println(str)

	return apiSpec, nil
}

func fetchSpec(api *serviceapi.API) ([]byte, apperrors.AppError) {
	specUrl, err := url.ParseRequestURI(api.SpecUrl)
	if err != nil {
		specUrl, err = url.ParseRequestURI(fmt.Sprintf(oDataSpecFormat, api.TargetUrl))
		if err != nil {
			return nil, apperrors.Internal("Parsing OData spec url failed, %s", err.Error())
		}
		api.Type = oDataSpecType
	}

	response, apperr := requestAPISpec(specUrl.String(), api.Credentials)
	if apperr != nil {
		return nil, apperr
	}

	spec, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, apperrors.Internal("Reading API spec response body failed, %s", err.Error())
	}

	return spec, nil
}

func isNilOrEmpty(array []byte) bool {
	return array == nil || len(array) == 0
}

func requestAPISpec(specUrl string, credentials *serviceapi.Credentials) (*http.Response, apperrors.AppError) {
	req, err := http.NewRequest(http.MethodGet, specUrl, nil)
	if err != nil {
		return nil, apperrors.Internal("Creating request for fetching API spec from %s failed, %s", specUrl, err.Error())
	}

	if credentials != nil {
		// TODO: setup authentication
	}

	response, err := http.DefaultClient.Do(req)
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
