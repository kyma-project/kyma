package validationproxy

import (
	"crypto/x509/pkix"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/application-connectivity-validator/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-connectivity-validator/internal/httptools"
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	CertificateInfoHeader = "X-Forwarded-Client-Cert"
)

type ProxyHandler interface {
	ProxyAppConnectorRequests(w http.ResponseWriter, r *http.Request)
}

//go:generate mockery -name=ApplicationGetter
type ApplicationGetter interface {
	Get(name string, options v1.GetOptions) (*v1alpha1.Application, error)
}

//go:generate mockery -name=Cache
type Cache interface {
	Get(k string) (interface{}, bool)
	Set(k string, x interface{}, d time.Duration)
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

	applicationGetter ApplicationGetter
	cache             Cache
}

func NewProxyHandler(group, tenant, eventServicePathPrefixV1, eventServicePathPrefixV2, eventServiceHost, appRegistryPathPrefix, appRegistryHost string, applicationGetter ApplicationGetter, cache Cache) *proxyHandler {
	return &proxyHandler{
		group:                    group,
		tenant:                   tenant,
		eventServicePathPrefixV1: eventServicePathPrefixV1,
		eventServicePathPrefixV2: eventServicePathPrefixV2,
		eventServiceHost:         eventServiceHost,
		appRegistryPathPrefix:    appRegistryPathPrefix,
		appRegistryHost:          appRegistryHost,

		eventsProxy:      createReverseProxy(eventServiceHost),
		appRegistryProxy: createReverseProxy(appRegistryHost),

		applicationGetter: applicationGetter,
		cache:             cache,
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

	applicationClientIDs, err := ph.getCompassMetadataClientIDs(applicationName)
	if err != nil {
		httptools.RespondWithError(w, apperrors.Internal("Failed to get Application ClientIds: %s", err))
		return
	}

	subjects := extractSubjects(certInfoData)

	if !hasValidSubject(subjects, applicationClientIDs, applicationName, ph.group, ph.tenant) {
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

func (ph *proxyHandler) getCompassMetadataClientIDs(applicationName string) ([]string, apperrors.AppError) {
	applicationClientIDs, found := ph.getClientIDsFromCache(applicationName)
	if !found {
		var err apperrors.AppError
		applicationClientIDs, err = ph.getClientIDsFromResource(applicationName)
		if err != nil {
			return []string{}, err
		}

		ph.cache.Set(applicationName, applicationClientIDs, cache.DefaultExpiration)
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

func (ph *proxyHandler) getClientIDsFromResource(applicationName string) ([]string, apperrors.AppError) {
	application, err := ph.applicationGetter.Get(applicationName, v1.GetOptions{})
	if err != nil {
		return []string{}, apperrors.Internal("failed to get %s application: %s", applicationName, err)
	}
	if application.Spec.CompassMetadata == nil {
		return []string{}, nil
	}

	return application.Spec.CompassMetadata.Authentication.ClientIds, nil
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
