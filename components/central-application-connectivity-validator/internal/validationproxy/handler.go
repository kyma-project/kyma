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
	group                    string
	tenant                   string
	eventServicePathPrefixV1 string
	eventServicePathPrefixV2 string
	eventServiceHost         string
	eventMeshPathPrefix      string
	eventMeshHost            string
	appRegistryPathPrefix    string
	appRegistryHost          string

	eventsProxy      *httputil.ReverseProxy
	eventMeshProxy   *httputil.ReverseProxy
	appRegistryProxy *httputil.ReverseProxy

	log *logger.Logger

	cache Cache
}

func NewProxyHandler(
	appNamePlaceholder string,
	group string,
	tenant string,
	eventServicePathPrefixV1 string,
	eventServicePathPrefixV2 string,
	eventMeshPathPrefix string,
	eventMeshHost string,
	eventMeshDestinationPath string,
	appRegistryPathPrefix string,
	appRegistryHost string,
	cache Cache,
	log *logger.Logger) *proxyHandler {
	return &proxyHandler{
		appNamePlaceholder:       appNamePlaceholder,
		group:                    group,
		tenant:                   tenant,
		eventServicePathPrefixV1: eventServicePathPrefixV1,
		eventServicePathPrefixV2: eventServicePathPrefixV2,
		eventMeshPathPrefix:      eventMeshPathPrefix,
		eventMeshHost:            eventMeshHost,
		appRegistryPathPrefix:    appRegistryPathPrefix,
		appRegistryHost:          appRegistryHost,

		eventsProxy:      createReverseProxy(log, eventMeshHost, withEmptyRequestHost, withEmptyXFwdClientCert, withHTTPScheme),
		eventMeshProxy:   createReverseProxy(log, eventMeshHost, withRewriteBaseURL(eventMeshDestinationPath), withEmptyRequestHost, withEmptyXFwdClientCert, withHTTPScheme),
		appRegistryProxy: createReverseProxy(log, appRegistryHost, withEmptyRequestHost, withHTTPScheme),

		cache: cache,
		log:   log,
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
		httptools.RespondWithError(ph.log.WithTracing(r.Context()).With("handler", handlerName).With("applicationName", applicationName), w, apperrors.Internal("while getting application ClientIds: %s", err))
		return
	}

	subjects := extractSubjects(certInfoData)

	if !hasValidSubject(subjects, applicationClientIDs, applicationName, ph.group, ph.tenant) {
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
		err := apperrors.Internal("application with name %s is not found in the cache. Please retry", applicationName)
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
	// In a BEB enabled cluster, validator should forward the event coming to /{application}/v2/events and /{application}/events to /publish endpoint of EventPublisherProxy (https://github.com/kyma-project/kyma/tree/main/components/event-publisher-proxy#send-events)
	switch {
	// Events reaching with prefix /{application}/v1/events will be routed to /{application}/v1/events endpoint of event-publisher-proxy
	case strings.HasPrefix(path, ph.getApplicationPrefix(ph.eventServicePathPrefixV1, applicationName)):
		return ph.eventsProxy, nil

	// Events reaching /{application}/v2/events will be routed to /publish endpoint of event-publisher-proxy
	case strings.HasPrefix(path, ph.getApplicationPrefix(ph.eventServicePathPrefixV2, applicationName)):
		return ph.eventMeshProxy, nil

	// Events reaching /{application}/events will be routed to /publish endpoint of event-publisher-proxy
	case strings.HasPrefix(path, ph.getApplicationPrefix(ph.eventMeshPathPrefix, applicationName)):
		return ph.eventMeshProxy, nil

	case strings.HasPrefix(path, ph.getApplicationPrefix(ph.appRegistryPathPrefix, applicationName)):
		return ph.appRegistryProxy, nil
	}

	return nil, apperrors.NotFound("could not determine destination host, requested resource not found")
}

func (ph *proxyHandler) getApplicationPrefix(path string, applicationName string) string {
	if ph.appNamePlaceholder != "" {
		return strings.ReplaceAll(path, ph.appNamePlaceholder, applicationName)
	}
	return path
}

func hasValidSubject(subjects, applicationClientIDs []string, appName, group, tenant string) bool {
	subjectValidator := newSubjectValidator(applicationClientIDs, appName, group, tenant)

	for _, s := range subjects {
		parsedSubject := parseSubject(s)

		if subjectValidator(parsedSubject) {
			return true
		}
	}

	return false
}

func newSubjectValidator(applicationClientIDs []string, appName, group, tenant string) func(subject pkix.Name) bool {
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
	validateSubjectField := func(subjectField []string, expectedValue string) bool {
		return len(subjectField) == 1 && subjectField[0] == expectedValue
	}

	switch {
	case len(applicationClientIDs) == 0 && !areStringsFilled(group, tenant):
		return validateCommonNameWithAppName

	case len(applicationClientIDs) == 0 && areStringsFilled(group, tenant):
		return func(subject pkix.Name) bool {
			return validateCommonNameWithAppName(subject) && validateSubjectField(subject.Organization, tenant) && validateSubjectField(subject.OrganizationalUnit, group)
		}

	case len(applicationClientIDs) != 0 && !areStringsFilled(group, tenant):
		return validateCommonNameWithClientIDs

	case len(applicationClientIDs) != 0 && areStringsFilled(group, tenant):
		return func(subject pkix.Name) bool {
			return validateCommonNameWithClientIDs(subject) && validateSubjectField(subject.Organization, tenant) && validateSubjectField(subject.OrganizationalUnit, group)
		}

	default:
		return func(subject pkix.Name) bool {
			return false
		}
	}
}

func areStringsFilled(strs ...string) bool {
	for _, str := range strs {
		if str == "" {
			return false
		}
	}
	return true
}

func extractSubjects(certInfoData string) []string {
	var subjects []string

	subjectRegex := regexp.MustCompile(`Subject="(.*?)"`)
	subjectMatches := subjectRegex.FindAllStringSubmatch(certInfoData, -1)

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
