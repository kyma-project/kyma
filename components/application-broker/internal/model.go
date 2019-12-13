package internal

import (
	"time"
)

// ApplicationName is a Application name
type ApplicationName string

// ApplicationServiceID is an ID of Service defined in Application
type ApplicationServiceID string

// Application represents Application as defined by OSB API.
type Application struct {
	Name        ApplicationName
	Description string
	Services    []Service
	AccessLabel string
}

// Service represents service defined in the application which is mapped to service class in the service catalog.
type Service struct {
	ID                  ApplicationServiceID
	Name                string
	DisplayName         string
	Description         string
	LongDescription     string
	ProviderDisplayName string

	Tags   []string
	Labels map[string]string

	//TODO(entry-simplification): this is an accepted simplification until
	// explicit support of many APIEntry and EventEntry
	APIEntry      *APIEntry
	EventProvider bool
}

// Entry is a generic type for all type of entries.
type Entry struct {
	Type string
}

// APIEntry represents API of the application.
type APIEntry struct {
	Entry
	GatewayURL  string
	AccessLabel string
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

	// ParamsHash is an immutable hash for operation parameters
	// used to match requests.
	ParamsHash string

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
	ParamsHash    string
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
	OperationDescriptionProvisioningSucceeded   string = "provisioning succeeded"
	OperationDescriptionDeprovisioningSucceeded string = "deprovisioning succeeded"
)

// InstanceState defines the possible states of the Instance in the storage.
type InstanceState string

const (
	// InstanceStatePending is when provision is in progress
	InstanceStatePending InstanceState = "pending"
	// InstanceStateFailed is when provision was failed
	InstanceStateFailed InstanceState = "failed"
	// InstanceStateSucceeded is when provision was succeeded
	InstanceStateSucceeded InstanceState = "succeeded"
)
