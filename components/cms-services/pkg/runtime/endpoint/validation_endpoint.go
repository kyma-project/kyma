package endpoint

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/kyma-project/kyma/components/cms-services/pkg/runtime/service"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
)

type validationEndpoint struct {
	name      string
	validator Validator
}

// Validator is the interface implemented by objects that can validate requests.
type Validator interface {
	Validate(ctx context.Context, reader io.Reader, parameters string) error
}

var _ service.HTTPEndpoint = &validationEndpoint{}

var (
	httpServeAnValidationHistogram = promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "cms_services_http_request_and_validation_duration_seconds",
		Help: "Request handling and validation duration distribution",
	})
	validatorStatusCodeCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cms_services_handle_validation_status_code",
		Help: "Status code returned by validation handler",
	}, []string{"status_code"})
)

func incrementValidationStatusCounter(status int) {
	validatorStatusCodeCounter.WithLabelValues(strconv.Itoa(status)).Inc()
}

// NewValidation is the constructor that creates a new Validation Endpoint.
func NewValidation(name string, validator Validator) service.HTTPEndpoint {
	return &validationEndpoint{
		name:      name,
		validator: validator,
	}
}

// Name returns the name of the endpoint.
func (e *validationEndpoint) Name() string {
	return e.name
}

// Handle processes an HTTP request and calls the Validator.
func (e *validationEndpoint) Handle(writer http.ResponseWriter, request *http.Request) {
	start := time.Now()

	defer request.Body.Close()

	if request.Method != http.MethodPost {
		http.Error(writer, "Invalid request method", http.StatusMethodNotAllowed)
		incrementValidationStatusCounter(http.StatusMethodNotAllowed)
		return
	}

	if err := request.ParseMultipartForm(32 << 20); err != nil {
		log.Error(errors.Wrap(err, "while parsing a multipart request"))
		http.Error(writer, err.Error(), http.StatusBadRequest)
		incrementValidationStatusCounter(http.StatusBadRequest)
		return
	}
	defer request.MultipartForm.RemoveAll()

	content, _, err := request.FormFile("content")
	if err != nil {
		log.Error(errors.Wrap(err, "while accessing the content"))
		http.Error(writer, err.Error(), http.StatusBadRequest)
		incrementValidationStatusCounter(http.StatusBadRequest)
		return
	}
	defer content.Close()

	parameters := request.FormValue("parameters")

	if err := e.validator.Validate(request.Context(), content, parameters); err != nil {
		log.Error(errors.Wrap(err, "while validating the request"))
		http.Error(writer, err.Error(), http.StatusUnprocessableEntity)
		incrementValidationStatusCounter(http.StatusUnprocessableEntity)
		return
	}

	writer.WriteHeader(http.StatusOK)
	incrementValidationStatusCounter(http.StatusOK)
	
	httpServeAnValidationHistogram.Observe(time.Since(start).Seconds())
}
