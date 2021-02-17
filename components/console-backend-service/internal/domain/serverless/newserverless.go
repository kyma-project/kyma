package serverless

import "github.com/kyma-project/kyma/components/console-backend-service/internal/resource"

type NewResolver struct {
	*resource.Module
}

// Temporary name - change it when old implementation will be rewritten
func NewR(factory *resource.GenericServiceFactory) *NewResolver {
	module := resource.NewModule("serverless", factory, resource.ServiceCreators{
		gitRepositoriesGroupVersionResource: newGitRepositoryService,
	})

	return &NewResolver{
		Module: module,
	}
}

func (r *NewResolver) GitRepositoryService() *resource.GenericService {
	return r.Module.Service(gitRepositoriesGroupVersionResource)
}
