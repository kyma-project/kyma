// +build acceptance

package k8s

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/tests/console-backend-service/internal/client"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/domain/shared/auth"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/graphql"
	"github.com/kyma-project/kyma/tests/console-backend-service/pkg/waiter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	limitRangeName      = "test-limit-range"
	limitRangeNamespace = "console-backend-service-lr"
)

func TestLimitRangeQuery(t *testing.T) {
	c, err := graphql.New()
	require.NoError(t, err)

	client, _, err := client.NewClientWithConfig()
	require.NoError(t, err)

	t.Log("Creating namespace...")
	_, err = client.Namespaces().Create(fixNamespace(limitRangeNamespace))
	require.NoError(t, err)

	defer func() {
		t.Log("Deleting namespace...")
		err = client.Namespaces().Delete(limitRangeNamespace, &metav1.DeleteOptions{})
		require.NoError(t, err)
	}()

	_, err = client.LimitRanges(limitRangeNamespace).Create(fixLimitRange())
	require.NoError(t, err)

	err = waiter.WaitAtMost(func() (bool, error) {
		_, err := client.LimitRanges(limitRangeNamespace).Get(limitRangeName, metav1.GetOptions{})
		if err == nil {
			return true, nil
		}
		return false, err
	}, time.Minute)
	require.NoError(t, err)

	t.Log("Querying for Limit Ranges...")
	var res limitRangeQueryResponse
	err = c.Do(fixLimitRangeQuery(), &res)
	require.NoError(t, err)

	assert.Equal(t, fixLimitRangeQueryResponse(), res)

	t.Log("Checking authorization directives...")
	ops := &auth.OperationsInput{
		auth.List: {fixLimitRangeQuery()},
	}
	AuthSuite.Run(t, ops)
}

type limitRangeQueryResponse struct {
	LimitRange []limitRange `json:"limitRanges"`
}

type limitRange struct {
	Name   string           `json:"name"`
	Limits []limitRangeItem `json:"limits"`
}

type limitRangeItem struct {
	LimitType      limitType    `json:"limitType"`
	DefaultRequest resourceType `json:"defaultRequest"`
	Default        resourceType `json:"default"`
	Max            resourceType `json:"max"`
}

type resourceType struct {
	Memory string `json:"memory"`
	Cpu    string `json:"cpu"`
}

type limitType string

const (
	limitTypeContainer limitType = "Container"
)

func fixLimitRangeQuery() *graphql.Request {
	query := `query ($namespace: String!) {
				limitRanges(namespace: $namespace) {
					name
					limits {
						limitType
						max {
							memory
						}
						default {
							memory
						}
						defaultRequest {
							memory
						}
					}
				}
			}`
	req := graphql.NewRequest(query)
	req.SetVar("namespace", limitRangeNamespace)

	return req
}

func fixLimitRangeQueryResponse() limitRangeQueryResponse {
	return limitRangeQueryResponse{
		LimitRange: []limitRange{
			{
				Name: limitRangeName,
				Limits: []limitRangeItem{
					{
						LimitType: limitTypeContainer,
						Max: resourceType{
							Memory: "1Gi",
						},
						Default: resourceType{
							Memory: "96Mi",
						},
						DefaultRequest: resourceType{
							Memory: "32Mi",
						},
					},
				},
			},
		},
	}
}

func fixLimitRange() *v1.LimitRange {
	return &v1.LimitRange{
		ObjectMeta: metav1.ObjectMeta{
			Name: limitRangeName,
		},
		Spec: v1.LimitRangeSpec{
			Limits: []v1.LimitRangeItem{
				{
					Type: v1.LimitTypeContainer,
					Max: v1.ResourceList{
						v1.ResourceMemory: resource.MustParse("1Gi"),
					},
					Default: v1.ResourceList{
						v1.ResourceMemory: resource.MustParse("96Mi"),
					},
					DefaultRequest: v1.ResourceList{
						v1.ResourceMemory: resource.MustParse("32Mi"),
					},
				},
			},
		},
	}
}
