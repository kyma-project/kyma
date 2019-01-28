package clientcontext

import (
	"context"
	"encoding/json"
)

const (
	ApplicationHeader     = "Application"
	ApplicationContextKey = "ApplicationContext"

	ClusterContextKey = "ClusterContext"
	TenantHeader      = "Tenant"
	GroupHeader       = "Group"
)

type ConnectorClientContext interface {
	GetApplication() string
	ToJSON() ([]byte, error)
	GetCommonName() string
}

type ContextExtender interface {
	ExtendContext(ctx context.Context) context.Context
}

type ConnectorClientReader interface {
	GetApplication() string
	GetCommonName() string
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
