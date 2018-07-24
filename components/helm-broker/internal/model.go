package internal

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/Masterminds/semver"
	"github.com/alecthomas/jsonschema"
	"github.com/fatih/structs"
	"github.com/pkg/errors"
)

// BundleID is a Bundle identifier as defined by Open Service Broker API.
type BundleID string

// BundleName is a Bundle name as defined by Open Service Broker API.
type BundleName string

// BundlePlanID is an identifier of Bundle plan as defined by Open Service Broker API.
type BundlePlanID string

// BundlePlanName is the name of the Bundle plan as defined by Open Service Broker API
type BundlePlanName string

// PlanSchemaType describes type of the schema file.
type PlanSchemaType string

// PlanSchema is schema definition used for creating parameters
type PlanSchema jsonschema.Schema

const (
	// SchemaTypeBind represents 'bind' schema plan
	SchemaTypeBind PlanSchemaType = "bind"
	// SchemaTypeProvision represents 'provision' schema plan
	SchemaTypeProvision PlanSchemaType = "provision"
	// SchemaTypeUpdate represents 'update' schema plan
	SchemaTypeUpdate PlanSchemaType = "update"
)

// ChartName is a type expressing name of the chart
type ChartName string

// ChartRef provide reference to bundle's chart
type ChartRef struct {
	Name    ChartName
	Version semver.Version
}

// GobDecode is decoding chart info
func (cr *ChartRef) GobDecode(in []byte) error {
	var dto struct {
		Name    ChartName
		Version string
	}

	buf := bytes.NewReader(in)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(&dto); err != nil {
		return errors.Wrap(err, "while decoding")
	}

	cr.Name = dto.Name

	ver, _ := semver.NewVersion(dto.Version)
	cr.Version = *ver

	return nil
}

// GobEncode implements GobEncoder for custom encoding
func (cr ChartRef) GobEncode() ([]byte, error) {
	dto := struct {
		Name    ChartName
		Version string
	}{
		Name:    cr.Name,
		Version: cr.Version.String(),
	}

	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(&dto); err != nil {
		return []byte{}, errors.Wrap(err, "while encoding")
	}

	return buf.Bytes(), nil
}

// ChartValues are used as container for chart's values.
// It's currently populated from yaml file or request parameters.
// TODO: switch to more concrete type
type ChartValues map[string]interface{}

// BundlePlanBindTemplate represents template used for helm chart installation
type BundlePlanBindTemplate []byte

// BundlePlan is a container for whole data of bundle plan.
// Each bundle needs to have at least one plan.
type BundlePlan struct {
	ID           BundlePlanID
	Name         BundlePlanName
	Description  string
	Schemas      map[PlanSchemaType]PlanSchema
	ChartRef     ChartRef
	ChartValues  ChartValues
	Metadata     BundlePlanMetadata
	Bindable     *bool
	BindTemplate BundlePlanBindTemplate
}

// BundlePlanMetadata provides metadata of the bundle.
type BundlePlanMetadata struct {
	DisplayName string
}

// ToMap function is converting Metadata to format compatible with YAML encoder.
func (b BundlePlanMetadata) ToMap() map[string]interface{} {
	type mapped struct {
		DisplayName string `structs:"displayName"`
	}

	return structs.Map(mapped(b))
}

// BundleTag is a Tag attached to Bundle.
type BundleTag string

// Bundle represents bundle as defined by OSB API.
type Bundle struct {
	ID          BundleID
	Name        BundleName
	Version     semver.Version
	Description string
	Plans       map[BundlePlanID]BundlePlan
	Metadata    BundleMetadata
	Tags        []BundleTag
	Bindable    bool
}

// BundleMetadata holds bundle metadata as defined by OSB API.
type BundleMetadata struct {
	DisplayName         string
	ProviderDisplayName string
	LongDescription     string
	DocumentationURL    string
	SupportURL          string
	// ImageURL is graphical representation of the bundle.
	// Currently SVG is required.
	ImageURL string
}

// ToMap collect data from BundleMetadata to format compatible with YAML encoder.
func (b BundleMetadata) ToMap() map[string]interface{} {
	type mapped struct {
		DisplayName         string `structs:"displayName"`
		ProviderDisplayName string `structs:"providerDisplayName"`
		LongDescription     string `structs:"longDescription"`
		DocumentationURL    string `structs:"documentationURL"`
		SupportURL          string `structs:"supportURL"`
		ImageURL            string `structs:"imageURL"`
	}
	return structs.Map(mapped(b))
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

// ReleaseName is the name of the Helm (Tiller) release.
type ReleaseName string

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
	ReleaseName   ReleaseName
	Namespace     Namespace
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
