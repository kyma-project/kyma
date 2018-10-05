package externalapi

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/metadata-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/metadata-service/internal/httpconsts"
	"github.com/kyma-project/kyma/components/metadata-service/internal/httperrors"
	"github.com/kyma-project/kyma/components/metadata-service/internal/httptools"
	"github.com/kyma-project/kyma/components/metadata-service/internal/metadata"
)

type metadataHandler struct {
	validator                ServiceDetailsValidator
	ServiceDefinitionService metadata.ServiceDefinitionService
	detailedErrorResponse    bool
}

func NewMetadataHandler(validator ServiceDetailsValidator, serviceDefinitionService metadata.ServiceDefinitionService, detailedErrorResponse bool) MetadataHandler {
	return &metadataHandler{
		validator:                validator,
		ServiceDefinitionService: serviceDefinitionService,
		detailedErrorResponse:    detailedErrorResponse,
	}
}

func (mh *metadataHandler) CreateService(w http.ResponseWriter, r *http.Request) {
	contextLogger := httptools.ContextLogger(r)
	httptools.DumpRequestToLog(r, contextLogger)

	serviceDefinition, apperr := mh.prepareServiceDefinition(r.Body)
	if apperr != nil {
		contextLogger.Errorf("metadata handler: preparing new service failed, %s", apperr.Error())
		mh.handleErrors(w, apperr)
		return
	}

	serviceId, apperr := mh.ServiceDefinitionService.Create(mux.Vars(r)["remoteEnvironment"], &serviceDefinition)
	if apperr != nil {
		contextLogger.Errorf("metadata handler: creating new service failed, %s", apperr.Error())
		mh.handleErrors(w, apperr)
		return
	}

	responseBody := CreateServiceResponse{ID: serviceId}
	respondWithBody(w, http.StatusOK, responseBody)

	contextLogger.Infof("Service with ID %s created successfully.", serviceId)
}

func (mh *metadataHandler) GetService(w http.ResponseWriter, r *http.Request) {
	contextLogger := httptools.ContextLoggerWithId(r)
	httptools.DumpRequestToLog(r, contextLogger)

	service, apperr := mh.ServiceDefinitionService.GetByID(mux.Vars(r)["remoteEnvironment"], mux.Vars(r)["serviceId"])
	if apperr != nil {
		contextLogger.Errorf("metadata handler: getting service by ID failed, %s", apperr.Error())
		mh.handleErrors(w, apperr)
		return
	}

	responseBody, apperr := serviceDefinitionToServiceDetails(service)
	if apperr != nil {
		contextLogger.Errorf("metadata handler: getting service failed, %s", apperr.Error())
		mh.handleErrors(w, apperr)
		return
	}

	respondWithBody(w, http.StatusOK, responseBody)
	contextLogger.Info("Service read successfully.")
}

func (mh *metadataHandler) GetServices(w http.ResponseWriter, r *http.Request) {
	contextLogger := httptools.ContextLogger(r)
	httptools.DumpRequestToLog(r, contextLogger)

	services, apperr := mh.ServiceDefinitionService.GetAll(mux.Vars(r)["remoteEnvironment"])
	if apperr != nil {
		contextLogger.Errorf("metadata handler: getting all services failed, %s", apperr.Error())
		mh.handleErrors(w, apperr)
		return
	}

	responseBody := make([]Service, 0)
	for _, element := range services {
		responseBody = append(responseBody, serviceDefinitionToService(element))
	}

	respondWithBody(w, http.StatusOK, responseBody)
	contextLogger.Info("Services read successfully.")
}

func (mh *metadataHandler) UpdateService(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contextLogger := httptools.ContextLoggerWithId(r)
	httptools.DumpRequestToLog(r, contextLogger)

	serviceDefinition, apperr := mh.prepareServiceDefinition(r.Body)
	if apperr != nil {
		contextLogger.Errorf("metadata handler: updating service failed, %s", apperr.Error())
		mh.handleErrors(w, apperr)
		return
	}

	svc, apperr := mh.ServiceDefinitionService.Update(vars["remoteEnvironment"], vars["serviceId"], &serviceDefinition)
	if apperr != nil {
		contextLogger.Errorf("metadata handler: updating service failed, %s", apperr.Error())
		mh.handleErrors(w, apperr)
		return
	}

	responseBody, apperr := serviceDefinitionToServiceDetails(svc)
	if apperr != nil {
		contextLogger.Errorf("metadata handler: updating service failed, %s", apperr.Error())
		mh.handleErrors(w, apperr)
		return
	}

	respondWithBody(w, http.StatusOK, responseBody)
	contextLogger.Info("Service updated successfully.")
}

func (mh *metadataHandler) DeleteService(w http.ResponseWriter, r *http.Request) {
	contextLogger := httptools.ContextLoggerWithId(r)
	httptools.DumpRequestToLog(r, contextLogger)

	vars := mux.Vars(r)

	apperr := mh.ServiceDefinitionService.Delete(vars["remoteEnvironment"], vars["serviceId"])
	if apperr != nil {
		contextLogger.Errorf("metadata handler: deleting service failed, %s", apperr.Error())
		mh.handleErrors(w, apperr)
		return
	}

	respond(w, http.StatusNoContent)
	contextLogger.Infof("Service deleted successfully.")
}

func (mh *metadataHandler) prepareServiceDefinition(body io.ReadCloser) (metadata.ServiceDefinition, apperrors.AppError) {
	b, err := ioutil.ReadAll(body)
	if err != nil {
		return metadata.ServiceDefinition{}, apperrors.WrongInput("request body: failed to read, %s", err.Error())
	}
	defer body.Close()

	var serviceDetails ServiceDetails
	err = json.Unmarshal(b, &serviceDetails)
	if err != nil {
		return metadata.ServiceDefinition{}, apperrors.WrongInput("request body: failed to unmarshal, %s", err.Error())
	}

	appErr := mh.validator.Validate(serviceDetails)
	if appErr != nil {
		return metadata.ServiceDefinition{}, apperrors.WrongInput("request body: failed to validate, %s", appErr.Error())
	}

	return serviceDetailsToServiceDefinition(serviceDetails)
}

func (mh *metadataHandler) handleErrors(w http.ResponseWriter, apperr apperrors.AppError) {
	statusCode, responseBody := httperrors.AppErrorToResponse(apperr, mh.detailedErrorResponse)

	respond(w, statusCode)
	json.NewEncoder(w).Encode(responseBody)
}

func respond(w http.ResponseWriter, statusCode int) {
	w.Header().Set(httpconsts.HeaderContentType, httpconsts.ContentTypeApplicationJson)
	w.WriteHeader(statusCode)
}

func respondWithBody(w http.ResponseWriter, statusCode int, responseBody interface{}) {
	respond(w, statusCode)
	json.NewEncoder(w).Encode(responseBody)
}

func compact(src []byte) []byte {
	buffer := new(bytes.Buffer)
	err := json.Compact(buffer, src)
	if err != nil {
		return src
	}
	return buffer.Bytes()
}
