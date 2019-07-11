package clientcontext

import (
	"context"

	"github.com/kyma-project/kyma/components/connector-service/internal/logging"
	"github.com/sirupsen/logrus"
)

type ClusterContext struct {
	Group     string `json:"group,omitempty"`
	Tenant    string `json:"tenant,omitempty"`
	RuntimeID string `json:"runtimeId,omitempty"`
}

// NewClusterContextExtender creates empty ClusterContext
func NewClusterContextExtender() ContextExtender {
	return &ClusterContext{}
}

// IsEmpty returns false if Group, Tenant and RuntimeID are set
func (clsCtx ClusterContext) IsEmpty() bool {
	return clsCtx.Group == GroupEmpty || clsCtx.Tenant == TenantEmpty || clsCtx.RuntimeID == RuntimeIDEmpty
}

// ExtendContext extends provided context with ClusterContext
func (clsCtx ClusterContext) ExtendContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ClusterContextKey, clsCtx)
}

// GetLogger returns context logger with embedded context data (Group, Tenant and RuntimeID)
func (clsCtx ClusterContext) GetLogger() *logrus.Entry {
	return logging.GetClusterLogger(clsCtx.Tenant, clsCtx.Group, clsCtx.RuntimeID)
}

// GetRuntimeID returns RuntimeID
func (clsCtx ClusterContext) GetRuntimeID() *string {
	return &clsCtx.RuntimeID
}

// GetRuntimeUrls returns nil as ClusterContext does not contain RuntimeURLs
func (clsCtx ClusterContext) GetRuntimeUrls() *RuntimeURLs {
	return nil
}
