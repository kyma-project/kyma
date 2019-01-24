package httpcontext

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

// TODO - look for better name - maybe client context
type KymaContext interface {
	GetApplication() string
	ToJSON() ([]byte, error)
	GetCommonName() string
}

type Serializer interface {
	ToJSON() ([]byte, error)
}

type ContextExtender interface {
	ExtendContext(ctx context.Context) context.Context
}

type ContextReader interface {
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

func (appCtx ApplicationContext) GetApplication() string {
	return appCtx.Application
}

func (appCtx ApplicationContext) GetCommonName() string {
	// TODO - construct CN
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

func (clsCtx ClusterContext) GetApplication() string {
	return ""
}

func (clsCtx ClusterContext) GetCommonName() string {
	// TODO - construct CN
	return clsCtx.Group + clsCtx.Tenant
}
