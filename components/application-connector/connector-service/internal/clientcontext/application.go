package clientcontext

import (
	"context"
	"encoding/json"

	"github.com/kyma-project/kyma/components/application-connector/connector-service/internal/logging"
	"github.com/sirupsen/logrus"
)

type ExtendedApplicationContext struct {
	ApplicationContext
	RuntimeURLs
}

// MarshalJSON marshals ExtendedApplicationContext to JSON as ApplicationContext
func (extAppCtx ExtendedApplicationContext) MarshalJSON() ([]byte, error) {
	return json.Marshal(extAppCtx.ApplicationContext)
}

// GetRuntimeUrls returns pointer to RuntimeURLs
func (extAppCtx ExtendedApplicationContext) GetRuntimeUrls() *RuntimeURLs {
	return &extAppCtx.RuntimeURLs
}

type ApplicationContext struct {
	Application string `json:"application"`
	ClusterContext
}

// NewApplicationContextExtender returns empty ApplicationContext
func NewApplicationContextExtender() ContextExtender {
	return &ApplicationContext{}
}

// IsEmpty returns false if Application is set
func (appCtx ApplicationContext) IsEmpty() bool {
	return appCtx.Application == ApplicationEmpty
}

// ExtendContext extends provided context with ApplicationContext
func (appCtx ApplicationContext) ExtendContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ApplicationContextKey, appCtx)
}

// GetApplication returns Application identifier
func (appCtx ApplicationContext) GetApplication() string {
	return appCtx.Application
}

// GetRuntimeUrls returns nil as ApplicationContext does not contain RuntimeURLs
func (appCtx ApplicationContext) GetRuntimeUrls() *RuntimeURLs {
	return nil
}

// GetLogger returns context logger with embedded context data (Application, Group and Tenant)
func (appCtx ApplicationContext) GetLogger() *logrus.Entry {
	return logging.GetApplicationLogger(appCtx.Application, appCtx.ClusterContext.Tenant, appCtx.ClusterContext.Group)
}
