package clientcontext

import (
	"context"
	"fmt"
	"strings"

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

// GetCommonName returns expected Common Name value for the Cluster
func (clsCtx ClusterContext) GetCommonName() string {
	return fmt.Sprintf("%s%s%s", clsCtx.Tenant, SubjectCNSeparator, clsCtx.Group)
}

// GetLogger returns context logger with embedded context data (Group and Tenant)
func (clsCtx ClusterContext) GetLogger() *logrus.Entry {
	return logging.GetClusterLogger(clsCtx.Tenant, clsCtx.Group)
}

// GetRuntimeUrls returns nil as ClusterContext does not contain RuntimeURLs
func (clsCtx ClusterContext) GetRuntimeUrls() *RuntimeURLs {
	return nil
}

// FillPlaceholders replaces placeholders {TENANT}, {GROUP} with values from the context
func (clsCtx ClusterContext) FillPlaceholders(format string) string {
	filledFormat := strings.Replace(format, TenantPlaceholder, clsCtx.Tenant, 1)
	filledFormat = strings.Replace(filledFormat, GroupPlaceholder, clsCtx.Group, 1)
	return filledFormat
}
