package test

import (
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8sNew"
	resourceFake "github.com/kyma-project/kyma/components/console-backend-service/internal/resource/fake"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func setupServiceWithObjects(t *testing.T, objects ...runtime.Object) *k8sNew.Resolver {
	serviceFactory, err := resourceFake.NewFakeGenericServiceFactory(v1.AddToScheme, objects...)
	require.NoError(t, err)

	service := k8sNew.New(serviceFactory)
	err = service.Enable()
	require.NoError(t, err)

	serviceFactory.InformerFactory.WaitForCacheSync(make(chan struct{}))

	return service
}

func setupEmptyService(t *testing.T) *k8sNew.Resolver {
	serviceFactory, err := resourceFake.NewFakeGenericServiceFactory(v1.AddToScheme)
	require.NoError(t, err)

	service := k8sNew.New(serviceFactory)
	err = service.Enable()
	require.NoError(t, err)

	serviceFactory.InformerFactory.WaitForCacheSync(make(chan struct{}))

	return service
}
