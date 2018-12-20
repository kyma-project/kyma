package externalapi

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/httpconsts"
	"github.com/kyma-project/kyma/components/application-registry/internal/httperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/httptools"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/model"
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
		contextLogger.Errorf("Creating new service failed, %s", apperr.Error())
		mh.handleErrors(w, apperr)
		return
	}

	serviceId, apperr := mh.ServiceDefinitionService.Create(mux.Vars(r)["application"], &serviceDefinition)
	if apperr != nil {
		contextLogger.Errorf("Creating new service failed, %s", apperr.Error())
		mh.handleErrors(w, apperr)
		return
	}

	responseBody := CreateServiceResponse{ID: serviceId}
	apperr = mh.respondWithBody(w, http.StatusOK, responseBody)
	if apperr != nil {
		contextLogger.Errorf("Creating service failed, %s", apperr.Error())
		mh.handleErrors(w, apperr)
		return
	}
	contextLogger.Infof("Service with ID %s created successfully.", serviceId)
}

func (mh *metadataHandler) GetService(w http.ResponseWriter, r *http.Request) {
	contextLogger := httptools.ContextLoggerWithId(r)
	httptools.DumpRequestToLog(r, contextLogger)

	service, apperr := mh.ServiceDefinitionService.GetByID(mux.Vars(r)["application"], mux.Vars(r)["serviceId"])
	if apperr != nil {
		contextLogger.Errorf("Getting service by ID failed, %s", apperr.Error())
		mh.handleErrors(w, apperr)
		return
	}

	responseBody, apperr := serviceDefinitionToServiceDetails(service)
	if apperr != nil {
		contextLogger.Errorf("Getting service failed, %s", apperr.Error())
		mh.handleErrors(w, apperr)
		return
	}

	apperr = mh.respondWithBody(w, http.StatusOK, responseBody)
	if apperr != nil {
		contextLogger.Errorf("Getting service failed, %s", apperr.Error())
		mh.handleErrors(w, apperr)
		return
	}
	contextLogger.Info("Service read successfully.")
}

func (mh *metadataHandler) GetServices(w http.ResponseWriter, r *http.Request) {
	contextLogger := httptools.ContextLogger(r)
	httptools.DumpRequestToLog(r, contextLogger)

	services, apperr := mh.ServiceDefinitionService.GetAll(mux.Vars(r)["application"])
	if apperr != nil {
		contextLogger.Errorf("Getting all services failed, %s", apperr.Error())
		mh.handleErrors(w, apperr)
		return
	}

	responseBody := make([]Service, 0)
	for _, element := range services {
		responseBody = append(responseBody, serviceDefinitionToService(element))
	}

	apperr = mh.respondWithBody(w, http.StatusOK, responseBody)
	if apperr != nil {
		contextLogger.Errorf("Getting all services failed, %s", apperr.Error())
		mh.handleErrors(w, apperr)
		return
	}
	contextLogger.Info("Services read successfully.")
}

func (mh *metadataHandler) UpdateService(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contextLogger := httptools.ContextLoggerWithId(r)
	httptools.DumpRequestToLog(r, contextLogger)

	serviceDefinition, apperr := mh.prepareServiceDefinition(r.Body)
	if apperr != nil {
		contextLogger.Errorf("Updating service failed, %s", apperr.Error())
		mh.handleErrors(w, apperr)
		return
	}
	serviceDefinition.ID = vars["serviceId"]

	svc, apperr := mh.ServiceDefinitionService.Update(vars["application"], &serviceDefinition)
	if apperr != nil {
		contextLogger.Errorf("Updating service failed, %s", apperr.Error())
		mh.handleErrors(w, apperr)
		return
	}

	responseBody, apperr := serviceDefinitionToServiceDetails(svc)
	if apperr != nil {
		contextLogger.Errorf("Updating service failed, %s", apperr.Error())
		mh.handleErrors(w, apperr)
		return
	}

	apperr = mh.respondWithBody(w, http.StatusOK, responseBody)
	if apperr != nil {
		contextLogger.Errorf("Updating service failed, %s", apperr.Error())
		mh.handleErrors(w, apperr)
		return
	}
	contextLogger.Info("Service updated successfully.")
}

func (mh *metadataHandler) DeleteService(w http.ResponseWriter, r *http.Request) {
	contextLogger := httptools.ContextLoggerWithId(r)
	httptools.DumpRequestToLog(r, contextLogger)

	vars := mux.Vars(r)

	apperr := mh.ServiceDefinitionService.Delete(vars["application"], vars["serviceId"])
	if apperr != nil {
		contextLogger.Errorf("Deleting service failed, %s", apperr.Error())
		mh.handleErrors(w, apperr)
		return
	}

	respond(w, http.StatusNoContent)
	contextLogger.Infof("Service deleted successfully.")
}

func (mh *metadataHandler) prepareServiceDefinition(body io.ReadCloser) (model.ServiceDefinition, apperrors.AppError) {
	b, err := ioutil.ReadAll(body)
	if err != nil {
		return model.ServiceDefinition{}, apperrors.WrongInput("Failed to read request body, %s", err.Error())
	}
	defer body.Close()

	var serviceDetails ServiceDetails
	err = json.Unmarshal(b, &serviceDetails)
	if err != nil {
		return model.ServiceDefinition{}, apperrors.WrongInput("Failed to unmarshal request body, %s", err.Error())
	}

	appErr := mh.validator.Validate(serviceDetails)
	if appErr != nil {
		return model.ServiceDefinition{}, apperrors.WrongInput("Failed to validate request body, %s", appErr.Error())
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

func (mh *metadataHandler) respondWithBody(w http.ResponseWriter, statusCode int, responseBody interface{}) apperrors.AppError {
	var b bytes.Buffer

	err := json.NewEncoder(&b).Encode(responseBody)
	if err != nil {
		return apperrors.Internal("Failed to marshall body, %s", err.Error())
	}

	respond(w, statusCode)
	w.Write(b.Bytes())
	return nil
}

func compact(src []byte) []byte {
	buffer := new(bytes.Buffer)
	err := json.Compact(buffer, src)
	if err != nil {
		return src
	}
	return buffer.Bytes()
}
