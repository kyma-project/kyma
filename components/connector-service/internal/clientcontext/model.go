package clientcontext

import (
	"context"

	"github.com/sirupsen/logrus"
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

	// ApplicationHeader is key representing Application in headers
	ApplicationHeader = "Application"

	// ApplicationContextKey is the key value for storing Application in context
	ApplicationContextKey clientContextKey = "ApplicationContext"

	// SubjectHeader is key representing client certificate subject set in headers
	SubjectHeader = "Client-Certificate-Subject"

	// APIHostsKey is the key value for storing API hosts in context
	APIHostsKey clientContextKey = "APIHosts"

	// ClusterContextKey is the key value for storing cluster data in context
	ClusterContextKey clientContextKey = "ClusterContext"

	// TenantHeader is key representing Tenant in headers
	TenantHeader = "Tenant"

	// GroupHeader is key representing Group in headers
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
