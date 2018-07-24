package k8s

import (
	"context"
	"testing"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/k8s/automock"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestResourceQuotaResolver_ResourceQuotasQuery(t *testing.T) {
	// GIVEN
	env := "production"
	lister := automock.NewResourceQuotaLister()
	lister.On("List", env).Return([]*v1.ResourceQuota{
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
