package roles

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
)

type Resolver struct {
	*resource.Module
}

func New(factory *resource.GenericServiceFactory) *Resolver {
	module := resource.NewModule("roles", factory, resource.ServiceCreators{
		roleGroupVersionResource:               NewRoleService,
		clusterRoleGroupVersionResource:        NewClusterRoleService,
		roleBindingGroupVersionResource:        NewRoleBindingService,
		clusterRoleBindingGroupVersionResource: NewClusterRoleBindingService,
	})

	return &Resolver{
		Module: module,
	}
}

func (r *Resolver) RoleService() *resource.GenericService {
	return r.Module.Service(roleGroupVersionResource)
}

func (r *Resolver) ClusterRoleService() *resource.GenericService {
	return r.Module.Service(clusterRoleGroupVersionResource)
}

func (r *Resolver) RoleBindingService() *resource.GenericService {
	return r.Module.Service(roleBindingGroupVersionResource)
}

func (r *Resolver) ClusterRoleBindingService() *resource.GenericService {
	return r.Module.Service(clusterRoleBindingGroupVersionResource)
}
