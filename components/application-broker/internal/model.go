package internal

import (
	"fmt"
	"time"

	"github.com/kyma-project/kyma/components/application-gateway/pkg/proxyconfig"
)

// ApplicationName is a Application name
type ApplicationName string

// ApplicationServiceID is an ID of Service defined in Application
type ApplicationServiceID string

type CompassMetadata struct {
	ApplicationID string
}

// Application represents Application as defined by OSB API.
type Application struct {
	Name                ApplicationName
	Description         string
	Services            []Service
	CompassMetadata     CompassMetadata
	DisplayName         string
	ProviderDisplayName string
	LongDescription     string
	Labels              map[string]string
	Tags                []string

	// Deprecated, remove in https://github.com/kyma-project/kyma/issues/7415
	AccessLabel string
}

// Service represents service defined in the application which is mapped to service class in the service catalog.
type Service struct {
	ID                                   ApplicationServiceID
	Name                                 string
	DisplayName                          string
	Description                          string
	Entries                              []Entry
	EventProvider                        bool
	ServiceInstanceCreateParameterSchema map[string]interface{}

	// Deprecated, remove in https://github.com/kyma-project/kyma/issues/7415
	LongDescription string
	// Deprecated, remove in https://github.com/kyma-project/kyma/issues/7415
	ProviderDisplayName string
	// Deprecated, remove in https://github.com/kyma-project/kyma/issues/7415
	Tags []string
	// Deprecated, remove in https://github.com/kyma-project/kyma/issues/7415
	Labels map[string]string
}

func (s *Service) IsBindable() bool {
	for _, e := range s.Entries {
		if e.Type == APIEntryType {
			return true
		}
	}
	return false
}

// Entry is a generic type for all type of entries.
type Entry struct {
	Type string
	*APIEntry
}

// APIEntry represents API of the application.
type APIEntry struct {
	Name      string
	TargetURL string
	ID        string

	// Deprecated, remove in https://github.com/kyma-project/kyma/issues/7415
	GatewayURL string
	// Deprecated, remove in https://github.com/kyma-project/kyma/issues/7415
	AccessLabel string
}

func (a *APIEntry) String() string {
	if a == nil {
		return "APIEntry: nil"
	}
	return fmt.Sprintf("APIEntry{Name: %s, TargetURL: %s, GateywaURL:%s, AccessLabel: %s}",
		a.Name,
		a.TargetURL,
		a.GatewayURL,
		a.AccessLabel,
	)
}

// InstanceID is a service instance identifier.
type InstanceID string

// IsZero checks if InstanceID equals zero.
func (id InstanceID) IsZero() bool { return id == InstanceID("") }

// OperationID is used as binding operation identifier.
type OperationID string

// IsZero checks if OperationID equals zero
func (id OperationID) IsZero() bool { return id == OperationID("") }

// InstanceOperation represents single operation.
type InstanceOperation struct {
	InstanceID       InstanceID
	OperationID      OperationID
	Type             OperationType
	State            OperationState
	StateDescription *string

	// CreatedAt points to creation time of the operation.
	// Field should be treated as immutable and is responsibility of storage implementation.
	// It should be set by storage Insert method.
	CreatedAt time.Time
}

// ServiceID is an ID of the Service exposed via Service Catalog.
type ServiceID string

// ServicePlanID is an ID of the Plan of Service exposed via Service Catalog.
type ServicePlanID string

// Namespace is the name of namespace in k8s
type Namespace string

// Instance contains info about Service exposed via Service Catalog.
type Instance struct {
	ID            InstanceID
	ServiceID     ServiceID
	ServicePlanID ServicePlanID
	Namespace     Namespace
	State         InstanceState
}

// InstanceCredentials are created when we bind a service instance.
type InstanceCredentials map[string]string

// InstanceBindData contains data about service instance and it's credentials.
type InstanceBindData struct {
	InstanceID  InstanceID
	Credentials InstanceCredentials
}

// OperationState defines the possible states of an asynchronous request to a broker.
type OperationState string

// String returns state of the operation.
func (os OperationState) String() string {
	return string(os)
}

// APIPackageCredential holds all information necessary to auth with a given Application API
// service.
type APIPackageCredential struct {
	ID     string
	Type   proxyconfig.AuthType
	Config proxyconfig.Configuration
}

const (
	// OperationStateInProgress means that operation is in progress
	OperationStateInProgress OperationState = "in progress"
	// OperationStateSucceeded means that request succeeded
	OperationStateSucceeded OperationState = "succeeded"
	// OperationStateFailed means that request failed
	OperationStateFailed OperationState = "failed"
)

// OperationType defines the possible types of an asynchronous operation to a broker.
type OperationType string

const (
	// OperationTypeCreate means creating OperationType
	OperationTypeCreate OperationType = "create"
	// OperationTypeRemove means removing OperationType
	OperationTypeRemove OperationType = "remove"
	// OperationTypeUndefined means undefined OperationType
	OperationTypeUndefined OperationType = ""
)

const (
	// OperationDescriptionProvisioningSucceeded means that the provisioning succeeded
	OperationDescriptionProvisioningSucceeded string = "provisioning succeeded"
	// OperationDescriptionDeprovisioningSucceeded means that the deprovisioning succeeded
	OperationDescriptionDeprovisioningSucceeded string = "deprovisioning succeeded"
)

// InstanceState defines the possible states of the Instance in the storage.
type InstanceState string

const (
	// InstanceStatePending is when provision is in progress
	InstanceStatePending InstanceState = "pending"
	// InstanceStatePendingDeletion is when deprovision is in progress
	InstanceStatePendingDeletion InstanceState = "removing"
	// InstanceStateFailed is when provision was failed
	InstanceStateFailed InstanceState = "failed"
	// InstanceStateSucceeded is when provision was succeeded
	InstanceStateSucceeded InstanceState = "succeeded"
)

const (
	APIEntryType   = "API"
	EventEntryType = "Events"
)
