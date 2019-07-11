package validationproxy

import (
	"crypto/x509/pkix"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/components/application-connectivity-validator/internal/httptools"

	"github.com/kyma-project/kyma/components/application-connectivity-validator/internal/apperrors"

	"github.com/gorilla/mux"
)

const (
	CertificateInfoHeader = "X-Forwarded-Client-Cert"
)

type ProxyHandler interface {
	ProxyAppConnectorRequests(w http.ResponseWriter, r *http.Request)
}

type proxyHandler struct {
	group                    string
	tenant                   string
	eventServicePathPrefixV1 string
	eventServicePathPrefixV2 string
	eventServiceHost         string
	appRegistryPathPrefix    string
	appRegistryHost          string

	eventsProxy      *httputil.ReverseProxy
	appRegistryProxy *httputil.ReverseProxy
}

func NewProxyHandler(group, tenant, eventServicePathPrefixV1, eventServicePathPrefixV2, eventServiceHost, appRegistryPathPrefix, appRegistryHost string) *proxyHandler {
	return &proxyHandler{
		group:  group,
		tenant: tenant,
		eventServicePathPrefixV1: eventServicePathPrefixV1,
		eventServicePathPrefixV2: eventServicePathPrefixV2,
		eventServiceHost:         eventServiceHost,
		appRegistryPathPrefix:    appRegistryPathPrefix,
		appRegistryHost:          appRegistryHost,

		eventsProxy:      createReverseProxy(eventServiceHost),
		appRegistryProxy: createReverseProxy(appRegistryHost),
	}
}

func (ph *proxyHandler) ProxyAppConnectorRequests(w http.ResponseWriter, r *http.Request) {
	certInfoData := r.Header.Get(CertificateInfoHeader)
	if certInfoData == "" {
		httptools.RespondWithError(w, apperrors.Internal("%s header not found", CertificateInfoHeader))
		return
	}

	applicationName := mux.Vars(r)["application"]
	if applicationName == "" {
		httptools.RespondWithError(w, apperrors.BadRequest("Application name not specified"))
		return
	}

	log.Infof("Proxying request for %s application. Path: %s", applicationName, r.URL.Path)

	subjects := extractSubjects(certInfoData)
	if !hasValidSubject(subjects, applicationName, ph.group, ph.tenant) {
		httptools.RespondWithError(w, apperrors.Forbidden("No valid subject found"))
		return
	}

	reverseProxy, err := ph.mapRequestToProxy(r.URL.Path)
	if err != nil {
		httptools.RespondWithError(w, err)
		return
	}

	reverseProxy.ServeHTTP(w, r)
}

func (ph *proxyHandler) mapRequestToProxy(path string) (*httputil.ReverseProxy, apperrors.AppError) {

	if strings.HasPrefix(path, ph.eventServicePathPrefixV1) {
		return ph.eventsProxy, nil
	}

	if strings.HasPrefix(path, ph.eventServicePathPrefixV2) {
		return ph.eventsProxy, nil
	}

	if strings.HasPrefix(path, ph.appRegistryPathPrefix) {
		return ph.appRegistryProxy, nil
	}

	return nil, apperrors.NotFound("Could not determine destination host. Requested resource not found")
}

func hasValidSubject(subjects []string, appName, group, tenant string) bool {
	subjectValidator := newSubjectValidator(appName, group, tenant)

	for _, s := range subjects {
		parsedSubject := parseSubject(s)

		if subjectValidator(parsedSubject) {
			return true
		}
	}

	return false
}

func newSubjectValidator(appName, group, tenant string) func(subject pkix.Name) bool {
	validateCommonName := func(subject pkix.Name) bool {
		return appName == subject.CommonName
	}

	if group == "" || tenant == "" {
		return validateCommonName
	}

	return func(subject pkix.Name) bool {
		return validateCommonName(subject) && validateSubjectField(subject.Organization, tenant) && validateSubjectField(subject.OrganizationalUnit, group)
	}
}

func validateSubjectField(subjectField []string, expectedValue string) bool {
	return len(subjectField) == 1 && subjectField[0] == expectedValue
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

func createReverseProxy(destinationHost string) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Director: func(request *http.Request) {
			request.URL.Scheme = "http"
			request.URL.Host = destinationHost
			log.Infof("Proxying request to target URL: %s", request.URL)
		},
		ModifyResponse: func(response *http.Response) error {
			log.Infof("Host responded with status: %s", response.Status)
			return nil
		},
	}
}
