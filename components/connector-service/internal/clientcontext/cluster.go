package clientcontext

import (
	"context"
	"encoding/json"

	"github.com/kyma-project/kyma/components/connector-service/internal/logging"
	"github.com/sirupsen/logrus"
)

type ClientContext struct {
	Group  string
	Tenant string
	ID     string
}

// NewClusterContextExtender creates empty ClientContext
func NewClusterContextExtender() ContextExtender {
	return &ClientContext{}
}

// IsEmpty returns false if Group, Tenant and ID are set
func (clsCtx ClientContext) IsEmpty() bool {
	return clsCtx.Group == GroupEmpty || clsCtx.Tenant == TenantEmpty || clsCtx.ID == IDEmpty
}

// ExtendContext extends provided context with ClientContext
func (clsCtx ClientContext) ExtendContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ClientContextKey, clsCtx)
}

// GetLogger returns context logger with embedded context data (Group, Tenant and ID)
func (clsCtx ClientContext) GetLogger() *logrus.Entry {
	return logging.GetLogger(clsCtx.Tenant, clsCtx.Group, clsCtx.ID)
}

// GetRuntimeUrls returns nil as ClientContext does not contain RuntimeURLs
func (clsCtx ClientContext) GetRuntimeUrls() *RuntimeURLs {
	return nil
}

// GetClientContext returns ClientContext
func (clsCtx ClientContext) GetClientContext() ClientContext {
	return clsCtx
}

type ExtendedApplicationContext struct {
	ClientContext
	RuntimeURLs
}

// MarshalJSON marshals ExtendedApplicationContext to JSON as ApplicationContext
func (extAppCtx ExtendedApplicationContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(extAppCtx.ClientContext)
}

// GetRuntimeUrls returns pointer to RuntimeURLs
func (extAppCtx ExtendedApplicationContext) GetRuntimeUrls() *RuntimeURLs {
	return &extAppCtx.RuntimeURLs
}
