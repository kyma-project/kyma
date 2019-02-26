package clientcontext

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kyma-project/kyma/components/connector-service/internal/logging"
	"github.com/sirupsen/logrus"
)

type ClusterContext struct {
	Group  string `json:"group"`
	Tenant string `json:"tenant"`
}

// NewClusterContextExtender creates ClusterContext
func NewClusterContextExtender() ContextExtender {
	return &ClusterContext{}
}

// IsEmpty returns false if both Group and Tenant are set
func (clsCtx ClusterContext) IsEmpty() bool {
	return clsCtx.Group == GroupEmpty || clsCtx.Tenant == TenantEmpty
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
	return ApplicationEmpty
}

// GetCommonName returns expected Common Name value for the Cluster
func (clsCtx ClusterContext) GetCommonName() string {
	return fmt.Sprintf("%s%s%s", clsCtx.Tenant, SubjectCNSeparator, clsCtx.Group)
}

func (clsCtx ClusterContext) GetLogger() *logrus.Entry {
	return logging.GetClusterLogger(clsCtx.Tenant, clsCtx.Group)

}

func (clsCtx ClusterContext) GetRuntimeUrls() *RuntimeURLs {
	return nil
}

func (clsCtx ClusterContext) FillPlaceholders(format string) string {
	filledFormat := strings.Replace(format, TenantPlaceholder, clsCtx.Tenant, 1)
	filledFormat = strings.Replace(filledFormat, GroupPlaceholder, clsCtx.Group, 1)
	return filledFormat
}
