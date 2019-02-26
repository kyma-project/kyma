package clientcontext

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kyma-project/kyma/components/connector-service/internal/logging"
	"github.com/sirupsen/logrus"
)

type ExtendedApplicationContext struct {
	ApplicationContext
	RuntimeURLs `json:"-"`
}

type ApplicationContext struct {
	Application string `json:"application"`
	ClusterContext
}

func NewApplicationContextExtender() ContextExtender {
	return &ApplicationContext{}
}

// IsEmpty returns false if Application is set
func (appCtx ApplicationContext) IsEmpty() bool {
	return appCtx.Application == ApplicationEmpty
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
	if appCtx.ClusterContext.IsEmpty() {
		return appCtx.Application
	}

	return fmt.Sprintf("%s%s%s%s%s", appCtx.ClusterContext.Tenant, SubjectCNSeparator,
		appCtx.ClusterContext.Group, SubjectCNSeparator, appCtx.Application)
}

func (appCtx ApplicationContext) GetRuntimeUrls() *RuntimeURLs {
	return nil
}

func (appCtx ApplicationContext) GetLogger() *logrus.Entry {
	return logging.GetApplicationLogger(appCtx.Application, appCtx.ClusterContext.Tenant, appCtx.ClusterContext.Group)
}

func (extAppCtx ExtendedApplicationContext) GetRuntimeUrls() *RuntimeURLs {
	return &extAppCtx.RuntimeURLs
}

func (appCtx ApplicationContext) FillPlaceholders(format string) string {
	filledFormat := strings.Replace(format, TenantPlaceholder, appCtx.ClusterContext.Tenant, 1)
	filledFormat = strings.Replace(filledFormat, GroupPlaceholder, appCtx.ClusterContext.Group, 1)
	filledFormat = strings.Replace(filledFormat, ApplicationPlaceholder, appCtx.Application, 1)
	return filledFormat
}
