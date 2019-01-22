package middlewares

const (
	ApplicationHeader     = "Application"
	ApplicationContextKey = "ApplicationContext"

	ClusterContextKey = "ClusterContext"
	TenantHeader      = "Tenant"
	GroupHeader       = "Group"
)

type ApplicationContext struct {
	Application string
}

// IsEmpty returns false if both Group and Tenant are set
func (context ApplicationContext) IsEmpty() bool {
	return context.Application == ""
}

type ClusterContext struct {
	Group  string
	Tenant string
}

// IsEmpty returns false if both Group and Tenant are set
func (context ClusterContext) IsEmpty() bool {
	return context.Group == "" || context.Tenant == ""
}
