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
	resourceQuotaName = "test-resource-quota"
)

type resourceQuotas struct {
	ResourceQuotas []resourceQuota `json:"resourceQuotas"`
}

type resourceQuotaSpec struct {
	Hard resourceQuotaHard `json:"hard"`
}

type resourceQuotaHard struct {
	Pods     string         `json:"pods"`
	Limits   resourceValues `json:"limits"`
	Requests resourceValues `json:"requests"`
}

type resourceQuota struct {
	Name string            `json:"name"`
	Spec resourceQuotaSpec `json:"spec"`
}

type resourceValues struct {
	Memory string `json:"memory"`
	Cpu    string `json:"cpu"`
}

func TestResourceQuotaQuery(t *testing.T) {
	c, err := graphql.New()
	require.NoError(t, err)

	k8sClient, _, err := client.NewClientWithConfig()
	require.NoError(t, err)

	_, err = k8sClient.ResourceQuotas(testNamespace).Create(fixResourceQuota())
	require.NoError(t, err)

	err = waiter.WaitAtMost(func() (bool, error) {
		_, err := k8sClient.ResourceQuotas(testNamespace).Get(resourceQuotaName, metav1.GetOptions{})
		if err == nil {
			return true, nil
		}
		return false, err
	}, time.Minute)
	require.NoError(t, err)

	var listResult resourceQuotas

	err = c.Do(fixListResourceQuotasQuery(), &listResult)
	require.NoError(t, err)
	assert.Contains(t, listResult.ResourceQuotas, fixListResourceQuotasResponse())

	t.Log("Checking authorization directives...")
	ops := &auth.OperationsInput{
		auth.List: {fixListResourceQuotasQuery()},
	}
	AuthSuite.Run(t, ops)
}

func fixListResourceQuotasQuery() *graphql.Request {
	query := `query($namespace: String!) {
				resourceQuotas(namespace: $namespace) {
					name
					spec {
						hard {
							limits {
								memory				
							}
							requests {
								memory				
							}
							pods
						}
					}
				}
			}`
	r := graphql.NewRequest(query)
	r.SetVar("namespace", testNamespace)

	return r
}

func fixListResourceQuotasResponse() resourceQuota {
	return resourceQuota{
		Name: resourceQuotaName,
		Spec: resourceQuotaSpec{
			Hard: resourceQuotaHard{
				Pods: "10",
				Limits: resourceValues{
					Memory: "1Gi",
				},
				Requests: resourceValues{
					Memory: "512Mi",
				},
			},
		},
	}
}

func fixResourceQuota() *v1.ResourceQuota {
	return &v1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourceQuotaName,
			Namespace: testNamespace,
		},
		Spec: v1.ResourceQuotaSpec{
			Hard: v1.ResourceList{
				v1.ResourcePods:           resource.MustParse("10"),
				v1.ResourceLimitsMemory:   resource.MustParse("1Gi"),
				v1.ResourceRequestsMemory: resource.MustParse("512Mi"),
			},
		},
	}
}
