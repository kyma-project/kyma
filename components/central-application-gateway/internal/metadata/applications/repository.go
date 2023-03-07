// Package applications contains components for accessing/modifying Application CRD
package applications

import (
	"context"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/normalization"
)

const (
	specAPIType = "API"
)

// Manager contains operations for managing Application CRD
//
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
	SkipVerify                  bool
	EncodeURL                   bool
}

type predicateFunc func(service v1alpha1.Service, entry v1alpha1.Entry) bool

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

// ServiceRepository contains operations for managing services stored in Application CRD
//
//go:generate mockery --name=ServiceRepository
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
	return r.get(appName, getMatchFunction(serviceName))
}

func (r *repository) GetByEntryName(appName, serviceName, entryName string) (Service, apperrors.AppError) {

	matchServiceAndEntry := func(service v1alpha1.Service, entry v1alpha1.Entry) bool {
		serviceMatchFunc := getMatchFunction(serviceName)
		return serviceMatchFunc(service, entry) && entryName == normalization.NormalizeName(entry.Name)
	}
	return r.get(appName, matchServiceAndEntry)
}

func getMatchFunction(serviceName string) predicateFunc {
	return func(service v1alpha1.Service, entry v1alpha1.Entry) bool {
		return serviceName == normalization.NormalizeName(service.DisplayName) && entry.Type == specAPIType
	}
}

func (r *repository) get(appName string, predicate func(service v1alpha1.Service, entry v1alpha1.Entry) bool) (Service, apperrors.AppError) {
	app, err := r.getApplication(appName)
	if err != nil {
		return Service{}, err
	}
	services := make([]Service, 0)
	infos := make([]string, 0)
	for _, service := range app.Spec.Services {
		for _, entry := range service.Entries {
			if predicate(service, entry) {
				services = append(services, convert(service, entry, app.Spec.SkipVerify, app.Spec.EncodeURL))
				infos = append(infos, fmt.Sprintf("service.ID: '%s', service.DisplayName: '%s', entry.Name: '%s'", service.ID, service.DisplayName, entry.Name))
			}
		}
	}

	if len(services) == 1 {
		return services[0], nil
	} else if len(services) > 1 {
		return Service{}, apperrors.WrongInput("multiple services found: %s", strings.Join(infos, " | "))
	} else {
		return Service{}, apperrors.NotFound("service not found")
	}
}

func (r *repository) getApplication(appName string) (*v1alpha1.Application, apperrors.AppError) {
	app, err := r.appManager.Get(context.Background(), appName, v1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			message := fmt.Sprintf("Application: %s not found.", appName)
			log.Warn(message)
			return nil, apperrors.NotFound(message)
		}

		message := fmt.Sprintf("failed to get Application '%s' : %s", appName, err.Error())
		log.Error(message)
		return nil, apperrors.Internal(message)
	}

	return app, nil
}

func convert(service v1alpha1.Service, entry v1alpha1.Entry, skipVerify, encodeURL bool) Service {
	api := &ServiceAPI{
		TargetURL:                   entry.TargetUrl,
		Credentials:                 convertCredentialsFromK8sType(entry.Credentials),
		RequestParametersSecretName: entry.RequestParametersSecretName,
		SkipVerify:                  skipVerify,
		EncodeURL:                   encodeURL,
	}

	return Service{
		ID:                  service.ID,
		Name:                service.Name,
		DisplayName:         service.DisplayName,
		LongDescription:     service.LongDescription,
		ProviderDisplayName: service.ProviderDisplayName,
		Tags:                service.Tags,
		API:                 api,
	}
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
