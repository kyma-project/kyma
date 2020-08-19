package k8s

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/automock"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceQuotaStatusResolver_ResourceQuotaStatus_HappyPath(t *testing.T) {
	// GIVEN
	statusChecker := automock.NewResourceQuotaStatusChecker()
	statusChecker.On("CheckResourceQuotaStatus", fixNamespaceName()).Return(&gqlschema.ResourceQuotasStatus{Exceeded: false}, nil)

	resolver := newResourceQuotaStatusResolver(statusChecker)

	// WHEN
	status, err := resolver.ResourceQuotasStatus(context.Background(), fixNamespaceName())

	// THEN
	require.NoError(t, err)
	assert.False(t, status.Exceeded)
}

func TestResourceQuotaStatusResolver_ResourceQuotaStatus_Error(t *testing.T) {
	// GIVEN
	statusChecker := automock.NewResourceQuotaStatusChecker()
	statusChecker.On("CheckResourceQuotaStatus", fixNamespaceName()).Return(&gqlschema.ResourceQuotasStatus{}, errors.New("something went wrong"))

	resolver := newResourceQuotaStatusResolver(statusChecker)

	// WHEN
	status, err := resolver.ResourceQuotasStatus(context.Background(), fixNamespaceName())

	// THEN
	require.Error(t, err)
	assert.Nil(t, status)
}
