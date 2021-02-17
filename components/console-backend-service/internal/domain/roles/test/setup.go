package test

import (
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/roles"
	resourceFake "github.com/kyma-project/kyma/components/console-backend-service/internal/resource/fake"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func setupServiceWithObjects(t *testing.T, objects ...runtime.Object) *roles.Resolver {
	serviceFactory, err := resourceFake.NewFakeGenericServiceFactory(v1.AddToScheme, objects...)
	require.NoError(t, err)

	service := roles.New(serviceFactory)
	err = service.Enable()
	require.NoError(t, err)

	serviceFactory.InformerFactory.WaitForCacheSync(make(chan struct{}))

	return service
}
