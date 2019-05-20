package proxy

import (
	"crypto/x509/pkix"
	"fmt"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/components/application-connector-proxy/internal/httptools"

	"github.com/kyma-project/kyma/components/application-connector-proxy/internal/apperrors"

	"github.com/gorilla/mux"
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	CertificateInfoHeader = "X-Forwarded-Client-Cert"
)

type ProxyHandler interface {
	ProxyAppConnectorRequests(w http.ResponseWriter, r *http.Request)
}

// ApplicationManager contains operations for managing Application CRD
type ApplicationManager interface {
	Update(application *v1alpha1.Application) (*v1alpha1.Application, error)
	Get(name string, options v1.GetOptions) (*v1alpha1.Application, error)
}

type proxyHandler struct {
	applicationClient      ApplicationManager
	eventServicePathPrefix string
	eventServiceHost       string
	appRegistryPathPrefix  string
	appRegistryHost        string
}

func NewProxyHandler(applicationClient ApplicationManager, eventServicePathPrefix, eventServiceHost, appRegistryPathPrefix, appRegistryHost string) *proxyHandler {
	return &proxyHandler{
		applicationClient:      applicationClient,
		eventServicePathPrefix: eventServicePathPrefix,
		eventServiceHost:       eventServiceHost,
		appRegistryPathPrefix:  appRegistryPathPrefix,
		appRegistryHost:        appRegistryHost,
	}
}

func (ph *proxyHandler) ProxyAppConnectorRequests(w http.ResponseWriter, r *http.Request) {
	certInfoData := r.Header.Get(CertificateInfoHeader)
	applicationName := mux.Vars(r)["application"]

	if applicationName == "" {
		httptools.RespondWithError(w, apperrors.BadRequest("Application name not specified"))
		return
	}

	log.Infof("Proxying request for %s application. Path: %s", applicationName, r.URL.Path)

	application, err := ph.getApplication(applicationName)
	if err != nil {
		httptools.RespondWithError(w, err)
		return
	}

	subjects := extractSubjects(certInfoData)
	if !hasValidSubject(subjects, application.Name, application.Spec.Group, application.Spec.Tenant) {
		httptools.RespondWithError(w, apperrors.Forbidden("No valid subject found"))
		return
	}

	path := strings.TrimPrefix(r.URL.Path, fmt.Sprintf("/%s", applicationName))

	host, err := ph.determineHost(path)
	if err != nil {
		httptools.RespondWithError(w, err)
		return
	}

	reverseProxy := &httputil.ReverseProxy{
		Director: func(request *http.Request) {
			request.URL.Scheme = "http"
			request.URL.Host = host
		},
		ModifyResponse: func(response *http.Response) error {
			log.Infof("Host responded with status: %s", response.Status)
			return nil
		},
	}

	reverseProxy.ServeHTTP(w, r)
}

func (ph *proxyHandler) getApplication(appName string) (*v1alpha1.Application, apperrors.AppError) {
	application, err := ph.applicationClient.Get(appName, v1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, apperrors.NotFound("Application %s not found", appName)
		}

		return nil, apperrors.Internal("Failed to get %s Application: %s", appName, err.Error())
	}

	return application, nil
}

func (ph *proxyHandler) determineHost(path string) (string, apperrors.AppError) {

	if strings.HasPrefix(path, ph.eventServicePathPrefix) {
		return ph.eventServiceHost, nil
	}

	if strings.HasPrefix(path, ph.appRegistryPathPrefix) {
		return ph.appRegistryHost, nil
	}

	return "", apperrors.NotFound("Could not determine destination host. Requested resource not found")
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

	if group == "" && tenant == "" {
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
