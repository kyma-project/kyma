package externalapi

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/application-connector/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-connector/internal/httpconsts"
	"github.com/kyma-project/kyma/components/application-connector/internal/httperrors"
	"github.com/kyma-project/kyma/components/application-connector/internal/httptools"
	"github.com/kyma-project/kyma/components/application-connector/internal/metadata"
)

type metadataHandler struct {
	validator                ServiceDetailsValidator
	ServiceDefinitionService metadata.ServiceDefinitionService
}

func NewMetadataHandler(validator ServiceDetailsValidator, serviceDefinitionService metadata.ServiceDefinitionService) MetadataHandler {
	return &metadataHandler{
		validator:                validator,
		ServiceDefinitionService: serviceDefinitionService,
	}
}

func (mh *metadataHandler) CreateService(w http.ResponseWriter, r *http.Request) {
	contextLogger := httptools.ContextLogger(r)
	httptools.DumpRequestToLog(r, contextLogger)

	serviceDefinition, apperr := mh.prepareServiceDefinition(r.Body)
	if apperr != nil {
		contextLogger.Errorf("Error creating new service: %s.", apperr.Error())
		handleErrors(w, apperr)
		return
	}

	serviceId, apperr := mh.ServiceDefinitionService.Create(mux.Vars(r)["remoteEnvironment"], &serviceDefinition)
	if apperr != nil {
		contextLogger.Errorf("Error creating new service: %s.", apperr.Error())
		handleErrors(w, apperr)
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
		contextLogger.Errorf("Error getting service: %s.", apperr.Error())
		handleErrors(w, apperr)
		return
	}

	responseBody, apperr := serviceDefinitionToServiceDetails(service)
	if apperr != nil {
		contextLogger.Errorf("Error getting service: %s.", apperr.Error())
		handleErrors(w, apperr)
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
		contextLogger.Errorf("Error getting services: %s.", apperr.Error())
		handleErrors(w, apperr)
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
		contextLogger.Errorf("Error updating service: %s.", apperr.Error())
		handleErrors(w, apperr)
		return
	}

	apperr = mh.ServiceDefinitionService.Update(vars["remoteEnvironment"], vars["serviceId"], &serviceDefinition)
	if apperr != nil {
		contextLogger.Errorf("Error updating service: %s.", apperr.Error())
		handleErrors(w, apperr)
		return
	}

	responseBody, apperr := serviceDefinitionToServiceDetails(serviceDefinition)
	if apperr != nil {
		contextLogger.Errorf("Error updating service: %s.", apperr.Error())
		handleErrors(w, apperr)
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
		contextLogger.Errorf("Error deleting service: %s.", apperr.Error())
		handleErrors(w, apperr)
		return
	}

	respond(w, http.StatusNoContent)
	contextLogger.Infof("Service deleted successfully.")
}

func (mh *metadataHandler) prepareServiceDefinition(body io.ReadCloser) (metadata.ServiceDefinition, apperrors.AppError) {
	b, err := ioutil.ReadAll(body)
	if err != nil {
		return metadata.ServiceDefinition{}, apperrors.WrongInput("failed to read request body, %s", err)
	}
	defer body.Close()

	var serviceDetails ServiceDetails
	err = json.Unmarshal(b, &serviceDetails)
	if err != nil {
		return metadata.ServiceDefinition{}, apperrors.WrongInput("failed to unmarshal request body, %s", err.Error())
	}

	appErr := mh.validator.Validate(serviceDetails)
	if appErr != nil {
		return metadata.ServiceDefinition{}, apperrors.WrongInput("failed to validate request body, %s", appErr.Error())
	}

	return serviceDetailsToServiceDefinition(serviceDetails)
}

func handleErrors(w http.ResponseWriter, apperr apperrors.AppError) {
	statusCode, responseBody := httperrors.AppErrorToResponse(apperr)

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
