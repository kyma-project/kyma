// Package applications contains components for accessing/modifying Application CRD
package applications

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apperrors"
	log "github.com/sirupsen/logrus"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	GatewayURL                  string
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
	// Mapped to type property under entries element (type: Events)
	Events bool
}

//go:generate mockery --name=ServiceRepository
// ServiceRepository contains operations for managing services stored in Application CRD
type ServiceRepository interface {
	Get(appName, serviceName, apiName string) (Service, apperrors.AppError)
}

// NewServiceRepository creates a new ApplicationServiceRepository
func NewServiceRepository(appManager Manager) ServiceRepository {
	return &repository{appManager: appManager}
}

// Get reads Service from Application by service name (bundle SKR mode) and apiName (entry
func (r *repository) Get(appName, serviceName, apiName string) (Service, apperrors.AppError) {
	app, err := r.getApplication(appName)
	if err != nil {
		return Service{}, err
	}

	for _, service := range app.Spec.Services {

		// service.Name personalization-webservices-v1-3b0c4
		// service.displayName Personalization Webservices v1 (oryginalny
		// serviceName personalization-webservices-v1 (to przychodzi z API)

		usedServiceName := serviceName
		//if len(apiName) == 0 {
		//	usedServiceName = createServiceName(serviceName, service.ID)
		//}

		if service.Name == usedServiceName {
			return convertFromK8sType(service, apiName)
		}
	}

	message := fmt.Sprintf("Service with name %s not found", serviceName)
	log.Warn(message)

	return Service{}, apperrors.NotFound(message)
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

func convertFromK8sType(service v1alpha1.Service, apiName string) (Service, apperrors.AppError) {
	var api *ServiceAPI
	var events bool
	// Kyma OS mode - find first
	if len(apiName) == 0 {
		for _, entry := range service.Entries {
			if entry.Type == specAPIType {
				api = &ServiceAPI{
					GatewayURL:                  entry.GatewayUrl,
					TargetURL:                   entry.TargetUrl,
					Credentials:                 convertCredentialsFromK8sType(entry.Credentials),
					RequestParametersSecretName: entry.RequestParametersSecretName,
				}
			} else if entry.Type == specEventsType {
				events = true
			} else {
				message := fmt.Sprintf("incorrect type of entry '%s' in Application Service definition", entry.Type)
				log.Error(message)
				return Service{}, apperrors.Internal(message)
			}
		}
	} else {
		// TODO Rozwiazac problem z ID
		// Management Plane mode to nie zadziała 1) get name + and Id odkleic sufix blebleble jesli nazwa jest jest postaci string_hash
		for _, entry := range service.Entries {
			if entry.Name == apiName {
				if entry.Type == specAPIType {
					api = &ServiceAPI{
						GatewayURL:                  entry.GatewayUrl,
						TargetURL:                   entry.TargetUrl,
						Credentials:                 convertCredentialsFromK8sType(entry.Credentials),
						RequestParametersSecretName: entry.RequestParametersSecretName,
					}
				} else if entry.Type == specEventsType {
					events = true
				} else {
					message := fmt.Sprintf("incorrect type of entry '%s' in Application Service definition", entry.Type)
					log.Error(message)
					return Service{}, apperrors.Internal(message)
				}
			}
		}
	}

	// r bedzie co jesli nie znajdzie?

	return Service{
		ID:                  service.ID,
		Name:                service.Name,
		DisplayName:         service.DisplayName,
		LongDescription:     service.LongDescription,
		ProviderDisplayName: service.ProviderDisplayName,
		Tags:                service.Tags,
		API:                 api, // nil!!!
		Events:              events,
	}, nil
}

var nonAlphaNumeric = regexp.MustCompile("[^A-Za-z0-9]+")

func createServiceName(serviceDisplayName, id string) string {
	// generate 5 characters suffix from the id
	sha := sha1.New()
	sha.Write([]byte(id))
	suffix := hex.EncodeToString(sha.Sum(nil))[:5]
	// remove all characters, which is not alpha numeric
	serviceDisplayName = nonAlphaNumeric.ReplaceAllString(serviceDisplayName, "-")
	// to lower
	serviceDisplayName = strings.Map(unicode.ToLower, serviceDisplayName)
	// trim dashes if exists
	serviceDisplayName = strings.TrimSuffix(serviceDisplayName, "-")
	if len(serviceDisplayName) > 57 {
		serviceDisplayName = serviceDisplayName[:57]
	}
	// add suffix
	serviceDisplayName = fmt.Sprintf("%s-%s", serviceDisplayName, suffix)
	// remove dash pre3 empty or had dash prefix
	serviceDisplayName = strings.TrimPrefix(serviceDisplayName, "-")
	return serviceDisplayName
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
