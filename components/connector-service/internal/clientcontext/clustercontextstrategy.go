package clientcontext

import "net/http"

type ClusterContextStrategy interface {
	ReadClusterContextFromRequest(r *http.Request) ClusterContext
	IsValidContext(clusterCtx ClientContext) bool
}

func NewClusterContextStrategy(clusterContextEnabled CtxEnabledType) ClusterContextStrategy {
	if clusterContextEnabled {
		return &clusterContextEnabledStrategy{}
	}

	return &clusterContextDisabledStrategy{}
}

type clusterContextEnabledStrategy struct{}

func (cc *clusterContextEnabledStrategy) ReadClusterContextFromRequest(r *http.Request) ClusterContext {
	clusterContext := ClusterContext{
		Tenant: r.Header.Get(TenantHeader),
		Group:  r.Header.Get(GroupHeader),
	}

	return clusterContext
}

func (cc *clusterContextEnabledStrategy) IsValidContext(clientContext ClientContext) bool {
	return !clientContext.IsEmpty()
}

type clusterContextDisabledStrategy struct{}

func (cc *clusterContextDisabledStrategy) ReadClusterContextFromRequest(r *http.Request) ClusterContext {
	return ClusterContext{}
}

func (cc *clusterContextDisabledStrategy) IsValidContext(clientContext ClientContext) bool {
	return clientContext.ID != IDEmpty && clientContext.Group == GroupEmpty && clientContext.Tenant == TenantEmpty
}
