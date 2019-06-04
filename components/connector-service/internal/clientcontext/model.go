package clientcontext

import (
	"context"

	"github.com/kyma-project/kyma/components/connector-service/internal/certificates"
	"github.com/sirupsen/logrus"
)

type clientContextKey string

// CtxEnabledType type defines if context is enabled
type CtxEnabledType bool

// LookupEnabledType type defines if headers must be specified
type LookupEnabledType bool

const (
	// ApplicationHeader is key representing Application in headers
	ApplicationHeader = "Application"

	// ApplicationContextKey is the key value for storing Application in context
	ApplicationContextKey clientContextKey = "ApplicationContext"

	// ApiURLsKey is the key value for storing API hosts in context
	ApiURLsKey clientContextKey = "ApiURLs"

	// ClusterContextKey is the key value for storing cluster data in context
	ClusterContextKey clientContextKey = "ClusterContext"

	// TenantHeader is key representing Tenant in headers
	TenantHeader = "Tenant"

	// GroupHeader is key representing Group in headers
	GroupHeader = "Group"

	// GroupEmpty represents empty value for Group
	GroupEmpty = ""

	// TenantEmpty represents empty value for Tenant
	TenantEmpty = ""

	// ApplicationEmpty represents empty value for Application
	ApplicationEmpty = ""

	// LookupEnabled represents value for required fetch from Runtime
	LookupEnabled LookupEnabledType = true

	// LookupDisabled represents value for not required fetch from Runtime
	LookupDisabled LookupEnabledType = false
)

type ClientContextService interface {
	GetRuntimeUrls() *RuntimeURLs
	GetLogger() *logrus.Entry
}

type ClientCertContextService interface {
	ClientContextService
	ClientContext() ClientContextService
	GetSubject() certificates.CSRSubject
}

type ContextExtender interface {
	ExtendContext(ctx context.Context) context.Context
}

type RuntimeURLs struct {
	EventsInfoURL string `json:"eventsInfoUrl"`
	EventsURL     string `json:"eventsUrl"`
	MetadataURL   string `json:"metadataUrl"`
}

type ApiURLs struct {
	EventsBaseURL   string
	MetadataBaseURL string
}

// ExtendContext extends provided context with ApiURLs
func (r ApiURLs) ExtendContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ApiURLsKey, r)
}
