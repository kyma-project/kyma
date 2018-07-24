package storage

import (
	"github.com/Masterminds/semver"

	"k8s.io/helm/pkg/proto/hapi/chart"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
)

// Bundle is an interface that describe storage layer operations for Bundles
type Bundle interface {
	Upsert(*internal.Bundle) (replace bool, err error)
	Get(internal.BundleName, semver.Version) (*internal.Bundle, error)
	GetByID(internal.BundleID) (*internal.Bundle, error)
	Remove(internal.BundleName, semver.Version) error
	RemoveByID(internal.BundleID) error
	FindAll() ([]*internal.Bundle, error)
}

// Chart is an interface that describe storage layer operations for Charts
type Chart interface {
	Upsert(*chart.Chart) (replace bool, err error)
	Get(name internal.ChartName, ver semver.Version) (*chart.Chart, error)
	Remove(name internal.ChartName, version semver.Version) error
}

// Instance is an interface that describe storage layer operations for Instances
type Instance interface {
	Insert(*internal.Instance) error
	Get(id internal.InstanceID) (*internal.Instance, error)
	Remove(id internal.InstanceID) error
}

// InstanceOperation is an interface that describe storage layer operations for InstanceOperations
type InstanceOperation interface {
	// Insert is inserting object into storage.
	// Object is modified by setting CreatedAt.
	Insert(*internal.InstanceOperation) error
	Get(internal.InstanceID, internal.OperationID) (*internal.InstanceOperation, error)
	GetAll(internal.InstanceID) ([]*internal.InstanceOperation, error)
	UpdateState(internal.InstanceID, internal.OperationID, internal.OperationState) error
	UpdateStateDesc(internal.InstanceID, internal.OperationID, internal.OperationState, *string) error
	Remove(internal.InstanceID, internal.OperationID) error
}

// InstanceBindData is an interface that describe storage layer operations for InstanceBindData entities
type InstanceBindData interface {
	Insert(*internal.InstanceBindData) error
	Get(internal.InstanceID) (*internal.InstanceBindData, error)
	Remove(internal.InstanceID) error
}

// IsNotFoundError checks if given error is NotFound error
func IsNotFoundError(err error) bool {
	nfe, ok := err.(interface {
		NotFound() bool
	})
	return ok && nfe.NotFound()
}

// IsAlreadyExistsError checks if given error is AlreadyExist error
func IsAlreadyExistsError(err error) bool {
	aee, ok := err.(interface {
		AlreadyExists() bool
	})
	return ok && aee.AlreadyExists()
}

// IsActiveOperationInProgressError checks if given error is ActiveOperationInProgress error
func IsActiveOperationInProgressError(err error) bool {
	aee, ok := err.(interface {
		ActiveOperationInProgress() bool
	})
	return ok && aee.ActiveOperationInProgress()
}
