package k8s

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	testingUtils "github.com/kyma-project/kyma/components/console-backend-service/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes/fake"
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

func TestLimitRangeResolver_CreateLimitRange(t *testing.T) {
	informer := fixLimitRangeInformer(fixLimitRange())
	client := fake.NewSimpleClientset(fixLimitRange())

	svc := newLimitRangeService(informer, client.CoreV1())
	resolver := newLimitRangeResolver(svc)

	limitRange := fixLimitRangeFromProperties("512Mi", "512Mi", "512Mi", "Container")
	result, err := resolver.CreateLimitRange(context.Background(), "testnamespace", "testlimitrangename", limitRange)

	expectedMemory := "512Mi"

	expectedResult := gqlschema.LimitRange{
		Name: "testlimitrangename",
		Limits: []gqlschema.LimitRangeItem{
			gqlschema.LimitRangeItem{
				LimitType: "Container",
				Max: gqlschema.ResourceType{
					Memory: &expectedMemory,
				},
				Default: gqlschema.ResourceType{
					Memory: &expectedMemory,
				},
				DefaultRequest: gqlschema.ResourceType{
					Memory: &expectedMemory,
				},
			},
		},
	}

	require.NoError(t, err)
	assert.Equal(t, &expectedResult, result)
}
