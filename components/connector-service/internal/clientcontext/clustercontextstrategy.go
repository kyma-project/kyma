package clientcontext

import "net/http"

type ClusterContextStrategy interface {
	ReadClusterContextFromRequest(r *http.Request) ClusterContext
	IsValidContext(clusterCtx ClusterContext) bool
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

func (cc *clusterContextEnabledStrategy) IsValidContext(clusterCtx ClusterContext) bool {
	return !clusterCtx.IsEmpty()
}

type clusterContextDisabledStrategy struct{}

func (cc *clusterContextDisabledStrategy) ReadClusterContextFromRequest(r *http.Request) ClusterContext {
	return ClusterContext{}
}

func (cc *clusterContextDisabledStrategy) IsValidContext(clusterCtx ClusterContext) bool {
	return clusterCtx.Group == GroupEmpty && clusterCtx.Tenant == TenantEmpty
}
