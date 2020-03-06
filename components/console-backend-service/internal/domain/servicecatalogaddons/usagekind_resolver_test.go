package servicecatalogaddons_test

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	testingUtils "github.com/kyma-project/kyma/components/console-backend-service/internal/testing"
	"github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned/fake"
	"github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/informers/externalversions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsageKindResolver_ListUsageKinds(t *testing.T) {
	// GIVEN
	resourceRef := fixKubelessFunctionResourceReference()
	usageKindA := fixUsageKind("fix-A", resourceRef)
	usageKindB := fixUsageKind("fix-B", resourceRef)
	usageKinds := []*v1alpha1.UsageKind{
		usageKindA,
		usageKindB,
	}
	gqlUsageKinds := []gqlschema.UsageKind{
		*fixUsageKindGQL("fix-A", resourceRef),
		*fixUsageKindGQL("fix-B", resourceRef),
	}

	svc := automock.NewUsageKindServices()
	svc.On("List", pager.PagingParams{}).
		Return(usageKinds, nil).
		Once()
	defer svc.AssertExpectations(t)

	client := fake.NewSimpleClientset(usageKindA, usageKindB)
	informerFactory := externalversions.NewSharedInformerFactory(client, 0)

	informer := informerFactory.Servicecatalog().V1alpha1().UsageKinds().Informer()
	testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)
	resolver := servicecatalogaddons.NewUsageKindResolver(svc)

	// WHEN
	resp, err := resolver.ListUsageKinds(context.Background(), nil, nil)

	// THEN
	require.NoError(t, err)
	assert.Equal(t, gqlUsageKinds, resp)
}
