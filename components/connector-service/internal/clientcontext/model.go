package clientcontext

import (
	"context"
	"encoding/json"
)

const (
	ApplicationHeader     = "Application"
	ApplicationContextKey = "ApplicationContext"
	SubjectHeader         = "Client-Certificate-Subject"
	APIHostsKey           = "APIHosts"

	ClusterContextKey = "ClusterContext"
	TenantHeader      = "Tenant"
	GroupHeader       = "Group"
)

type ConnectorClientContext interface {
	GetApplication() string
	ToJSON() ([]byte, error)
	GetCommonName() string
	GetRuntimeUrls() *RuntimeURLs
}

type ContextExtender interface {
	ExtendContext(ctx context.Context) context.Context
}

type ConnectorClientReader interface {
	GetApplication() string
	GetCommonName() string
}

type ExtendedApplicationContext struct {
	ApplicationContext
	RuntimeURLs
}

type ApplicationContext struct {
	Application    string
	ClusterContext ClusterContext
}

// IsEmpty returns false if Application is set
func (appCtx ApplicationContext) IsEmpty() bool {
	return appCtx.Application == "" || appCtx.ClusterContext.IsEmpty()
}

// ToJSON parses ApplicationContext to JSON
func (appCtx ApplicationContext) ToJSON() ([]byte, error) {
	return json.Marshal(appCtx)
}

// ExtendContext extends provided context with ApplicationContext
func (appCtx ApplicationContext) ExtendContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ApplicationContextKey, appCtx)
}

// GetApplication returns Application identifier
func (appCtx ApplicationContext) GetApplication() string {
	return appCtx.Application
}

// GetCommonName returns expected Common Name value for the Application
func (appCtx ApplicationContext) GetCommonName() string {
	// TODO - adjust CN format after decision is made
	return appCtx.Application
}

func (appCtx ApplicationContext) GetRuntimeUrls() *RuntimeURLs {
	return nil
}

func (extAppCtx ExtendedApplicationContext) GetRuntimeUrls() *RuntimeURLs {
	return &extAppCtx.RuntimeURLs
}

type ClusterContext struct {
	Group  string
	Tenant string
}

// IsEmpty returns false if both Group and Tenant are set
func (clsCtx ClusterContext) IsEmpty() bool {
	return clsCtx.Group == "" || clsCtx.Tenant == ""
}

// ToJSON parses ClusterContext to JSON
func (clsCtx ClusterContext) ToJSON() ([]byte, error) {
	return json.Marshal(clsCtx)
}

// ExtendContext extends provided context with ClusterContext
func (clsCtx ClusterContext) ExtendContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ClusterContextKey, clsCtx)
}

// GetApplication returns empty string
func (clsCtx ClusterContext) GetApplication() string {
	return ""
}

// GetCommonName returns expected Common Name value for the Cluster
func (clsCtx ClusterContext) GetCommonName() string {
	// TODO - adjust CN format after decision is made
	return clsCtx.Group + clsCtx.Tenant
}

func (clsCtx ClusterContext) GetRuntimeUrls() *RuntimeURLs {
	return nil
}

type APIHosts struct {
	EventsHost   string
	MetadataHost string
}

type RuntimeURLs struct {
	EventsURL   string `json:"eventsUrl,omitempty"`
	MetadataURL string `json:"metadataUrl,omitempty"`
}

func (r APIHosts) ExtendContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, APIHostsKey, r)
}
