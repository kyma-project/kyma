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

type mutationEndpoint struct {
	name    string
	mutator Mutator
}

// Mutator is the interface implemented by objects that can mutate other objects.
type Mutator interface {
	Mutate(ctx context.Context, reader io.Reader, parameters string) ([]byte, bool, error)
}

var _ service.HTTPEndpoint = &mutationEndpoint{}

var (
	httpServeAndMutationHistogram = promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "cms_services_http_request_and_mutation_duration_seconds",
		Help: "Request handling and mutation duration distribution",
	})
	mutationStatusCodeCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cms_services_handle_mutation_status_code",
		Help: "Status code returned by mutation handler",
	}, []string{"status_code"})
)

func incrementMutationStatusCodeCounter(status int) {
	mutationStatusCodeCounter.WithLabelValues(strconv.Itoa(status)).Inc()
}

// NewMutation is the constructor that creates a new Mutation Endpoint.
func NewMutation(name string, mutator Mutator) service.HTTPEndpoint {
	return &mutationEndpoint{
		name:    name,
		mutator: mutator,
	}
}

// Name returns the name of the endpoint.
func (e *mutationEndpoint) Name() string {
	return e.name
}

// Handle processes an HTTP request and calls the Mutator.
func (e *mutationEndpoint) Handle(writer http.ResponseWriter, request *http.Request) {
	start := time.Now()

	defer request.Body.Close()

	if request.Method != http.MethodPost {
		http.Error(writer, "Invalid request method", http.StatusMethodNotAllowed)
		incrementMutationStatusCodeCounter(http.StatusMethodNotAllowed)
		return
	}

	if err := request.ParseMultipartForm(32 << 20); err != nil {
		log.Error(errors.Wrap(err, "while parsing a multipart request"))
		http.Error(writer, err.Error(), http.StatusBadRequest)
		incrementMutationStatusCodeCounter(http.StatusBadRequest)
		return
	}
	defer request.MultipartForm.RemoveAll()

	content, _, err := request.FormFile("content")
	if err != nil {
		log.Error(errors.Wrap(err, "while accessing the content"))
		http.Error(writer, err.Error(), http.StatusBadRequest)
		incrementMutationStatusCodeCounter(http.StatusBadRequest)
		return
	}
	defer content.Close()

	parameters := request.FormValue("parameters")

	result, modified, err := e.mutator.Mutate(request.Context(), content, parameters)
	if err != nil {
		log.Error(errors.Wrap(err, "while mutating the request"))
		http.Error(writer, err.Error(), http.StatusUnprocessableEntity)
		incrementMutationStatusCodeCounter(http.StatusUnprocessableEntity)
		return
	}

	if !modified {
		writer.WriteHeader(http.StatusNotModified)
		incrementMutationStatusCodeCounter(http.StatusNotModified)
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write(result)
	incrementMutationStatusCodeCounter(http.StatusOK)
	
	httpServeAndMutationHistogram.Observe(time.Since(start).Seconds())
}
