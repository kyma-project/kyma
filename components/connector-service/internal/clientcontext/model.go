package clientcontext

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kyma-project/kyma/components/connector-service/internal/logging"
	"github.com/sirupsen/logrus"
	"strings"
)

type clientContextKey string

// CtxRequiredType type defines if context is mandatory
type CtxRequiredType bool

// HeadersRequiredType type defines if headers must be specified
type HeadersRequiredType bool

const (
	TenantPlaceholder      = "{TENANT}"
	GroupPlaceholder       = "{GROUP}"
	ApplicationPlaceholder = "{APPLICATION}"
  
	// ApplicationHeader is key represeting Application in headers
	ApplicationHeader = "Application"

	// ApplicationContextKey is the key value for storing Application in context
	ApplicationContextKey clientContextKey = "ApplicationContext"

	// SubjectHeader is key represeting client certificate subject set in headers
	SubjectHeader = "Client-Certificate-Subject"

	// APIHostsKey is the key value for storing API hosts in context
	APIHostsKey clientContextKey = "APIHosts"

	// ClusterContextKey is the key value for storing cluster data in context
	ClusterContextKey clientContextKey = "ClusterContext"

	// TenantHeader is key represeting Tenant in headers
	TenantHeader = "Tenant"

	// GroupHeader is key represeting Group in headers
	GroupHeader = "Group"

	// SubjectCNSeparator holds separator for values packed in CN of Subject
	SubjectCNSeparator = ";"

	// GroupEmpty represents empty value for Group
	GroupEmpty = ""

	// TenantEmpty represents empty value for Tenant
	TenantEmpty = ""

	// ApplicationEmpty represents empty value for Application
	ApplicationEmpty = ""

	// CtxRequired represents value for required context
	CtxRequired CtxRequiredType = true

	// CtxNotRequired represents value for not required context
	CtxNotRequired CtxRequiredType = false

	// HeadersRequired represents value for required headers
	HeadersRequired HeadersRequiredType = true

	// HeadersNotRequired represents value for not required headers
	HeadersNotRequired HeadersRequiredType = false
)

type ClientContextService interface {
	ToJSON() ([]byte, error)
	GetCommonName() string
	GetRuntimeUrls() *RuntimeURLs
	GetLogger() *logrus.Entry
	FillPlaceholders(format string) string
}

type ContextExtender interface {
	ExtendContext(ctx context.Context) context.Context
}

func NewClusterContextExtender() ContextExtender {
	return &ClusterContext{}
}

func NewApplicationContextExtender() ContextExtender {
	return &ApplicationContext{}
}

type ConnectorClientReader interface {
	GetApplication() string
	GetCommonName() string
}

type ExtendedApplicationContext struct {
	ApplicationContext
	RuntimeURLs `json:"-"`
}

type ApplicationContext struct {
	Application string `json:"application"`
	ClusterContext
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

func (appCtx ApplicationContext) FillPlaceholders(format string) string {
	filledFormat := strings.Replace(format, TenantPlaceholder, appCtx.ClusterContext.Tenant, 1)
	filledFormat = strings.Replace(filledFormat, GroupPlaceholder, appCtx.ClusterContext.Group, 1)
	filledFormat = strings.Replace(filledFormat, ApplicationPlaceholder, appCtx.Application, 1)
	return filledFormat
}

func (extAppCtx ExtendedApplicationContext) GetRuntimeUrls() *RuntimeURLs {
	return &extAppCtx.RuntimeURLs
}

type ClusterContext struct {
	Group  string `json:"group"`
	Tenant string `json:"tenant"`
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

func (clsCtx ClusterContext) FillPlaceholders(format string) string {
	filledFormat := strings.Replace(format, TenantPlaceholder, clsCtx.Tenant, 1)
	filledFormat = strings.Replace(filledFormat, GroupPlaceholder, clsCtx.Group, 1)
	return filledFormat
}

func (clsCtx ClusterContext) GetRuntimeUrls() *RuntimeURLs {
	return nil
}

type APIHosts struct {
	EventsHost   string
	MetadataHost string
}

type RuntimeURLs struct {
	EventsURL   string `json:"eventsUrl"`
	MetadataURL string `json:"metadataUrl"`
}

func (r APIHosts) ExtendContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, APIHostsKey, r)
}
