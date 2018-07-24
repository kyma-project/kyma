package k8s

import (
	"context"
	"testing"
	"time"

	testingUtils "github.com/kyma-project/kyma/components/ui-api-layer/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLimitRangeResolver_LimitRangeQuery(t *testing.T) {
	// GIVEN
	informer := fixLimitRangeInformer(fixLimitRange())

	svc := newLimitRangeService(informer)
	testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

	resolver := newLimitRangeResolver(svc)

	// WHEN
	result, err := resolver.LimitRangesQuery(context.Background(), fixLimitRangeNamespace())

	// THEN
	require.NoError(t, err)
	assert.Contains(t, result, *fixGQLLimitRange())
	assert.Len(t, result, 1)
}
