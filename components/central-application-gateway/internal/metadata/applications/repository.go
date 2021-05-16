// Package applications contains components for accessing/modifying Application CRD
package applications

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"
	log "github.com/sirupsen/logrus"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"regexp"
	"strings"
	"unicode"
)

const (
	specAPIType    = "API"
	specEventsType = "Events"
)

// Manager contains operations for managing Application CRD
//go:generate mockery --name=Manager
type Manager interface {
	Get(ctx context.Context, name string, options v1.GetOptions) (*v1alpha1.Application, error)
}

type repository struct {
	appManager Manager
}

// Credentials stores information about credentials needed to call an API
type Credentials struct {
	Type                 string
	SecretName           string
	URL                  string
	CSRFTokenEndpointURL string
}

// ServiceAPI stores information needed to call an API
type ServiceAPI struct {
	TargetURL                   string
	Credentials                 *Credentials
	RequestParametersSecretName string
}

// Service represents a service stored in Application
type Service struct {
	// Mapped to id in Application CRD
	ID string
	// Mapped to name in Application CRD
	Name string
	// Mapped to displayName in Application CRD
	DisplayName string
	// Mapped to longDescription in Application CRD
	LongDescription string
	// Mapped to providerDisplayName in Application CRD
	ProviderDisplayName string
	// Mapped to tags in Application CRD
	Tags []string
	// Mapped to type property under entries element (type: API)
	API *ServiceAPI
}

//go:generate mockery --name=ServiceRepository
// ServiceRepository contains operations for managing services stored in Application CRD
type ServiceRepository interface {
	GetByServiceName(appName, serviceName string) (Service, apperrors.AppError)
	GetByEntryName(appName, serviceName, entryName string) (Service, apperrors.AppError)
}

// NewServiceRepository creates a new ApplicationServiceRepository
func NewServiceRepository(appManager Manager) ServiceRepository {
	return &repository{appManager: appManager}
}

// Get reads Service from Application by service name (bundle SKR mode) and apiName (entry
func (r *repository) GetByServiceName(appName, serviceName string) (Service, apperrors.AppError) {
	return r.get(appName, matchService)
}

func (r *repository) GetByEntryName(appName, serviceName, entryName string) (Service, apperrors.AppError) {

	matchServiceAndEntry := func(service v1alpha1.Service, entry v1alpha1.Entry) bool {
		return matchService(service, entry) && entryName == normalizeName(entry.Name)
	}
	return r.get(appName, matchServiceAndEntry)
}

func matchService(service v1alpha1.Service, entry v1alpha1.Entry) bool {
	return (service.Name == normalizeName(service.DisplayName) ||
		service.Name == normalizeServiceNameWithId(service.DisplayName, service.ID)) &&
		entry.Type == specAPIType
}

func (r *repository) get(appName string, predicate func(service v1alpha1.Service, entry v1alpha1.Entry) bool) (Service, apperrors.AppError) {
	app, err := r.getApplication(appName)
	if err != nil {
		return Service{}, err
	}

	for _, service := range app.Spec.Services {
		for _, entry := range service.Entries {
			if predicate(service, entry) {
				return convert(service, entry)
			}
		}
	}

	return Service{}, apperrors.NotFound("service not found")
}

func (r *repository) getApplication(appName string) (*v1alpha1.Application, apperrors.AppError) {
	app, err := r.appManager.Get(context.Background(), appName, v1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			message := fmt.Sprintf("Application: %s not found.", appName)
			log.Warn(message)
			return nil, apperrors.Internal(message)
		}

		message := fmt.Sprintf("failed to get Application '%s' : %s", appName, err.Error())
		log.Error(message)
		return nil, apperrors.Internal(message)
	}

	return app, nil
}

func convert(service v1alpha1.Service, entry v1alpha1.Entry) (Service, apperrors.AppError) {
	api := &ServiceAPI{
		TargetURL:                   entry.TargetUrl,
		Credentials:                 convertCredentialsFromK8sType(entry.Credentials),
		RequestParametersSecretName: entry.RequestParametersSecretName,
	}

	return Service{
		ID:                  service.ID,
		Name:                service.Name,
		DisplayName:         service.DisplayName,
		LongDescription:     service.LongDescription,
		ProviderDisplayName: service.ProviderDisplayName,
		Tags:                service.Tags,
		API:                 api,
	}, nil
}

func convertCredentialsFromK8sType(credentials v1alpha1.Credentials) *Credentials {
	emptyCredentials := v1alpha1.Credentials{}
	if credentials == emptyCredentials {
		return nil
	}

	csrfTokenEndpointURL := ""
	if credentials.CSRFInfo != nil {
		csrfTokenEndpointURL = credentials.CSRFInfo.TokenEndpointURL
	}

	return &Credentials{
		Type:                 credentials.Type,
		SecretName:           credentials.SecretName,
		URL:                  credentials.AuthenticationUrl,
		CSRFTokenEndpointURL: csrfTokenEndpointURL,
	}
}

var nonAlphaNumeric = regexp.MustCompile("[^A-Za-z0-9]+")

// normalizeServiceNameWithId creates the OSB Service Name for given Application Service.
// The OSB Service Name is used in the Service Catalog as the clusterServiceClassExternalName, so it need to be normalized.
//
// Normalization rules:
// - MUST only contain lowercase characters, numbers and hyphens (no spaces).
// - MUST be unique across all service objects returned in this response. MUST be a non-empty string.
func normalizeServiceNameWithId(displayName, id string) string {

	normalizedName := normalizeName(displayName)

	// add suffix
	// generate 5 characters suffix from the id
	sha := sha1.New()
	sha.Write([]byte(id))
	suffix := hex.EncodeToString(sha.Sum(nil))[:5]
	normalizedName = fmt.Sprintf("%s-%s", normalizedName, suffix)
	// remove dash prefix if exists
	//  - can happen, if the name was empty before adding suffix empty or had dash prefix
	normalizedName = strings.TrimPrefix(normalizedName, "-")

	return normalizedName
}

func normalizeName(displayName string) string {

	// remove all characters, which is not alpha numeric
	normalizedName := nonAlphaNumeric.ReplaceAllString(displayName, "-")
	// to lower
	normalizedName = strings.Map(unicode.ToLower, normalizedName)
	// trim dashes if exists
	normalizedName = strings.TrimSuffix(displayName, "-")
	if len(normalizedName) > 57 {
		normalizedName = normalizedName[:57]
	}

	return strings.TrimPrefix(normalizedName, "-")
}
