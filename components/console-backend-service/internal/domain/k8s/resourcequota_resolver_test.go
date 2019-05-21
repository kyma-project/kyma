package k8s

import (
	"context"
	"errors"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"

	//	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestResourceQuotaResolver_ResourceQuotasQuery(t *testing.T) {
	// GIVEN
	env := "production"
	lister := automock.NewResourceQuotaLister()
	lister.On("ListResourceQuotas", env).Return([]*v1.ResourceQuota{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "mem-default",
				Namespace: "production",
			},
		},
	}, nil)
	defer lister.AssertExpectations(t)

	resolver := newResourceQuotaResolver(lister)

	// WHEN
	result, err := resolver.ResourceQuotasQuery(context.Background(), env)

	// THEN
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, gqlschema.ResourceQuota{Name: "mem-default"}, result[0])
}

func TestResourceQuotaResolver_CreateResourceQuota(t *testing.T) {
	const (
		name                   = "mem-default"
		namespace              = "production"
		resourceLimitsMemory   = "1Gi"
		resourceRequestsMemory = "512Mi"
	)

	var (
		resourceQuotaGQL = gqlschema.ResourceQuota{
			Name: name,
			Limits: gqlschema.ResourceValues{
				Memory: ptrStr(resourceLimitsMemory),
			},
			Requests: gqlschema.ResourceValues{
				Memory: ptrStr(resourceRequestsMemory),
			},
		}
		resourceQuota = v1.ResourceQuota{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: v1.ResourceQuotaSpec{
				Hard: v1.ResourceList{
					v1.ResourceLimitsMemory:   resource.MustParse(resourceLimitsMemory),
					v1.ResourceRequestsMemory: resource.MustParse(resourceRequestsMemory),
				},
			},
		}
	)

	t.Run("Success", func(t *testing.T) {
		createdResourceQuota := resourceQuota

		lister := automock.NewResourceQuotaLister()
		lister.On("CreateResourceQuota", namespace, name, resourceLimitsMemory, resourceRequestsMemory).Return(&createdResourceQuota, nil).Once()
		defer lister.AssertExpectations(t)

		converter := automock.NewGQLResourceQuotaConverter()
		converter.On("ToGQL", &resourceQuota).Return(&resourceQuotaGQL).Once()
		defer converter.AssertExpectations(t)

		resolver := newResourceQuotaResolver(lister)
		resolver.SetResourceQuotaConverter(converter)

		result, err := resolver.CreateResourceQuota(nil, namespace, name, resourceLimitsMemory, resourceRequestsMemory)

		require.NoError(t, err)
		assert.Equal(t, &resourceQuotaGQL, result)
	})

	t.Run("ErrorCreating", func(t *testing.T) {
		expected := errors.New("fix")

		lister := automock.NewResourceQuotaLister()
		lister.On("CreateResourceQuota", namespace, name, resourceLimitsMemory, resourceRequestsMemory).Return(nil, expected).Once()
		defer lister.AssertExpectations(t)

		resolver := newResourceQuotaResolver(lister)

		result, err := resolver.CreateResourceQuota(nil, namespace, name, resourceLimitsMemory, resourceRequestsMemory)

		require.Error(t, err)
		assert.True(t, gqlerror.IsInternal(err))
		assert.Nil(t, result)
	})
}
