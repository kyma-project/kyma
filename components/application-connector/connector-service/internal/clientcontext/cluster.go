package clientcontext

import (
	"context"

	"github.com/kyma-project/kyma/components/connector-service/internal/logging"
	"github.com/sirupsen/logrus"
)

type ClusterContext struct {
	Group  string `json:"group,omitempty"`
	Tenant string `json:"tenant,omitempty"`
}

// NewClusterContextExtender creates empty ClusterContext
func NewClusterContextExtender() ContextExtender {
	return &ClusterContext{}
}

// IsEmpty returns false if both Group and Tenant are set
func (clsCtx ClusterContext) IsEmpty() bool {
	return clsCtx.Group == GroupEmpty || clsCtx.Tenant == TenantEmpty
}

// ExtendContext extends provided context with ClusterContext
func (clsCtx ClusterContext) ExtendContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ClusterContextKey, clsCtx)
}

// GetLogger returns context logger with embedded context data (Group and Tenant)
func (clsCtx ClusterContext) GetLogger() *logrus.Entry {
	return logging.GetClusterLogger(clsCtx.Tenant, clsCtx.Group)
}

// GetRuntimeUrls returns nil as ClusterContext does not contain RuntimeURLs
func (clsCtx ClusterContext) GetRuntimeUrls() *RuntimeURLs {
	return nil
}
