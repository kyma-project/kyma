package validationproxy

import (
	"crypto/x509/pkix"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/common/logging/logger"
	"github.com/kyma-project/kyma/components/central-application-connectivity-validator/internal/apperrors"
	"github.com/kyma-project/kyma/components/central-application-connectivity-validator/internal/httptools"
)

const (
	CertificateInfoHeader = "X-Forwarded-Client-Cert"

	handlerName = "validation_proxy_handler"
)

type ProxyHandler interface {
	ProxyAppConnectorRequests(w http.ResponseWriter, r *http.Request)
}

type Cache interface {
	Get(k string) (interface{}, bool)
	Set(k string, x interface{}, d time.Duration)
}

type proxyHandler struct {
	appNamePlaceholder       string
	eventingPathPrefixV1     string
	eventingPathPrefixV2     string
	eventingPathPrefixEvents string
	eventingPublisherHost    string

	legacyEventsProxy *httputil.ReverseProxy
	cloudEventsProxy  *httputil.ReverseProxy

	log          *logger.Logger
	subjectRegex *regexp.Regexp

	cache Cache
}

func NewProxyHandler(
	appNamePlaceholder string,
	eventingPathPrefixV1 string,
	eventingPathPrefixV2 string,
	eventingPathPrefixEvents string,
	eventingPublisherHost string,
	eventingDestinationPath string,
	cache Cache,
	log *logger.Logger) *proxyHandler {
	return &proxyHandler{
		appNamePlaceholder:       appNamePlaceholder,
		eventingPathPrefixV1:     eventingPathPrefixV1,
		eventingPathPrefixV2:     eventingPathPrefixV2,
		eventingPathPrefixEvents: eventingPathPrefixEvents,
		eventingPublisherHost:    eventingPublisherHost,

		legacyEventsProxy: createReverseProxy(log, eventingPublisherHost, withEmptyRequestHost, withEmptyXFwdClientCert, withHTTPScheme),
		cloudEventsProxy:  createReverseProxy(log, eventingPublisherHost, withRewriteBaseURL(eventingDestinationPath), withEmptyRequestHost, withEmptyXFwdClientCert, withHTTPScheme),

		cache:        cache,
		log:          log,
		subjectRegex: regexp.MustCompile(`Subject="(.*?)"`),
	}
}

func (ph *proxyHandler) ProxyAppConnectorRequests(w http.ResponseWriter, r *http.Request) {
	certInfoData := r.Header.Get(CertificateInfoHeader)
	if certInfoData == "" {
		httptools.RespondWithError(ph.log.WithTracing(r.Context()).With("handler", handlerName), w, apperrors.Internal("%s header not found", CertificateInfoHeader))
		return
	}

	applicationName := mux.Vars(r)["application"]
	if applicationName == "" {
		httptools.RespondWithError(ph.log.WithTracing(r.Context()).With("handler", handlerName), w, apperrors.BadRequest("application name not specified"))
		return
	}

	ph.log.WithTracing(r.Context()).With("handler", handlerName).With("application", applicationName).With("proxyPath", r.URL.Path).Infof("Proxying request for application...")

	applicationClientIDs, err := ph.getCompassMetadataClientIDs(applicationName)
	if err != nil {
		httptools.RespondWithError(ph.log.WithTracing(r.Context()).With("handler", handlerName).With("applicationName", applicationName), w, apperrors.NotFound("while getting application ClientIds: %s", err))
		return
	}

	ph.log.WithTracing(r.Context()).With("handler", handlerName).With("application", applicationName).With("proxyPath", r.URL.Path).With("CertInfoData", certInfoData).Infof("Debug cert info")

	subjects := ph.extractSubjects(certInfoData)

	for _, s := range subjects {
		ph.log.WithTracing(r.Context()).With("handler", handlerName).With("application", applicationName).With("proxyPath", r.URL.Path).With("Subject", s).Infof("Debug cert header...")
	}

	if !hasValidSubject(subjects, applicationClientIDs, applicationName) {
		httptools.RespondWithError(ph.log.WithTracing(r.Context()).With("handler", handlerName).With("applicationName", applicationName), w, apperrors.Forbidden("no valid subject found"))
		return
	}

	reverseProxy, err := ph.mapRequestToProxy(r.URL.Path, applicationName)
	if err != nil {
		httptools.RespondWithError(ph.log.WithTracing(r.Context()).With("handler", handlerName).With("applicationName", applicationName), w, err)
		return
	}

	reverseProxy.ServeHTTP(w, r)
}

func (ph *proxyHandler) getCompassMetadataClientIDs(applicationName string) ([]string, apperrors.AppError) {
	applicationClientIDs, found := ph.getClientIDsFromCache(applicationName)
	if !found {
		err := apperrors.NotFound("application with name %s is not found in the cache. Please retry", applicationName)
		return nil, err
	}
	return applicationClientIDs, nil
}

func (ph *proxyHandler) getClientIDsFromCache(applicationName string) ([]string, bool) {
	clientIDs, found := ph.cache.Get(applicationName)
	if !found {
		return []string{}, found
	}
	return clientIDs.([]string), found
}

func (ph *proxyHandler) mapRequestToProxy(path string, applicationName string) (*httputil.ReverseProxy, apperrors.AppError) {
	switch {
	// legacy-events reaching /{application}/v1/events are routed to /{application}/v1/events endpoint of event-publisher-proxy
	case strings.HasPrefix(path, ph.getApplicationPrefix(ph.eventingPathPrefixV1, applicationName)):
		return ph.legacyEventsProxy, nil

	// cloud-events reaching /{application}/v2/events or /{application}/events are routed to /publish endpoint of event-publisher-proxy
	case strings.HasPrefix(path, ph.getApplicationPrefix(ph.eventingPathPrefixV2, applicationName)):
		return ph.cloudEventsProxy, nil

	// cloud-events reaching /{application}/events are routed to /publish endpoint of event-publisher-proxy
	case strings.HasPrefix(path, ph.getApplicationPrefix(ph.eventingPathPrefixEvents, applicationName)):
		return ph.cloudEventsProxy, nil
	}

	return nil, apperrors.NotFound("could not determine destination host, requested resource not found")
}

func (ph *proxyHandler) getApplicationPrefix(path string, applicationName string) string {
	if ph.appNamePlaceholder != "" {
		return strings.ReplaceAll(path, ph.appNamePlaceholder, applicationName)
	}
	return path
}

func hasValidSubject(subjects, applicationClientIDs []string, appName string) bool {
	subjectValidator := newSubjectValidator(applicationClientIDs, appName)

	for _, s := range subjects {
		parsedSubject := parseSubject(s)

		if subjectValidator(parsedSubject) {
			return true
		}
	}

	return false
}

func newSubjectValidator(applicationClientIDs []string, appName string) func(subject pkix.Name) bool {
	validateCommonNameWithAppName := func(subject pkix.Name) bool {
		return appName == subject.CommonName
	}
	validateCommonNameWithClientIDs := func(subject pkix.Name) bool {
		for _, id := range applicationClientIDs {
			if subject.CommonName == id {
				return true
			}
		}
		return false
	}
	if len(applicationClientIDs) == 0 {
		return validateCommonNameWithAppName
	} else {
		return validateCommonNameWithClientIDs
	}
}

func (ph *proxyHandler) extractSubjects(certInfoData string) []string {
	var subjects []string

	subjectMatches := ph.subjectRegex.FindAllStringSubmatch(certInfoData, -1)

	for _, subjectMatch := range subjectMatches {
		subject := get(subjectMatch, 1)

		if subject != "" {
			subjects = append(subjects, subject)
		}
	}

	return subjects
}

func get(array []string, index int) string {
	if len(array) > index {
		return array[index]
	}

	return ""
}

func parseSubject(rawSubject string) pkix.Name {
	subjectInfo := extractSubject(rawSubject)

	return pkix.Name{
		CommonName:         subjectInfo["CN"],
		Country:            []string{subjectInfo["C"]},
		Organization:       []string{subjectInfo["O"]},
		OrganizationalUnit: []string{subjectInfo["OU"]},
		Locality:           []string{subjectInfo["L"]},
		Province:           []string{subjectInfo["ST"]},
	}
}

func extractSubject(subject string) map[string]string {
	result := map[string]string{}

	segments := strings.Split(subject, ",")

	for _, segment := range segments {
		parts := strings.Split(segment, "=")
		result[parts[0]] = parts[1]
	}

	return result
}

func createReverseProxy(log *logger.Logger, destinationHost string, reqOpts ...requestOption) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Director: func(request *http.Request) {
			request.URL.Host = destinationHost
			for _, opt := range reqOpts {
				opt(request)
			}

			log.WithTracing(request.Context()).With("handler", handlerName).With("targetURL", request.URL).Infof("Proxying request to target URL...")
		},
		ModifyResponse: func(response *http.Response) error {
			log.WithContext().With("handler", handlerName).Infof("Host responded with status %s", response.Status)
			return nil
		},
	}
}

type requestOption func(req *http.Request)

// withRewriteBaseURL rewrites the Request's Path.
func withRewriteBaseURL(path string) requestOption {
	return func(req *http.Request) {
		req.URL.Path = path
	}
}

// withEmptyRequestHost clears the Request's Host field to ensure
// the 'Host' HTTP header is set to the host name defined in the Request's URL.
func withEmptyRequestHost(req *http.Request) {
	req.Host = ""
}

// withHTTPScheme sets the URL scheme to HTTP
func withHTTPScheme(req *http.Request) {
	req.URL.Scheme = "http"
}

// withEmptyXFwdClientCert clears the value of X-Forwarded-Client-Cert header
func withEmptyXFwdClientCert(req *http.Request) {
	req.Header.Del("X-Forwarded-Client-Cert")
}
