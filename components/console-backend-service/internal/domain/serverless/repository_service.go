package serverless

import (
	"github.com/kyma-project/kyma/components/console-backend-service/internal/resource"
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var gitRepositoryKind = "GitRepository"

var gitRepositoriesGroupVersionResource = schema.GroupVersionResource{
	Version:  v1alpha1.GroupVersion.Version,
	Group:    v1alpha1.GroupVersion.Group,
	Resource: "gitrepositories",
}

func newGitRepositoryService(serviceFactory *resource.GenericServiceFactory) (*resource.GenericService, error) {
	return serviceFactory.ForResource(gitRepositoriesGroupVersionResource), nil
}
