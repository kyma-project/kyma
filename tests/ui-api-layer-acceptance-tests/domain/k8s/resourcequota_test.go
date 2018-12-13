// +build acceptance

package k8s

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/client"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/dex"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/graphql"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/waiter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	resourceQuotaName      = "test-resource-quota"
	resourceQuotaNamespace = "test-resource-quota-ns"
)

type resourceQuotas struct {
	ResourceQuotas []resourceQuota `json:"resourceQuotas"`
}

type resourceQuota struct {
	Name     string         `json:"name"`
	Pods     string         `json:"pods"`
	Limits   resourceValues `json:"limits"`
	Requests resourceValues `json:"requests"`
}

type resourceValues struct {
	Memory string `json:"memory"`
	Cpu    string `json:"cpu"`
}

type resourceQuotaStatus struct {
	Exceeded bool   `json:"exceeded"`
	Message  string `json:"message"`
}

func TestResourceQuotaQuery(t *testing.T) {
	if dex.IsSCIEnabled() {
		t.Skip("SCI Enabled")
	}

	k8sClient, _, err := client.NewClientWithConfig()
	require.NoError(t, err)

	_, err = k8sClient.Namespaces().Create(fixNamespace(resourceQuotaNamespace))
	require.NoError(t, err)
	defer func() {
		err := k8sClient.Namespaces().Delete(resourceQuotaNamespace, &metav1.DeleteOptions{})
		assert.NoError(t, err)
	}()

	_, err = k8sClient.ResourceQuotas(resourceQuotaNamespace).Create(fixResourceQuota())
	require.NoError(t, err)

	c, err := graphql.New()
	require.NoError(t, err)

	waiter.WaitAtMost(func() (bool, error) {
		_, err := k8sClient.ResourceQuotas(resourceQuotaNamespace).Get(resourceQuotaName, metav1.GetOptions{})
		if err == nil {
			return true, nil
		}
		return false, err
	}, time.Minute)

	var listResult resourceQuotas
	var statusResult resourceQuotaStatus

	err = c.Do(fixListResourceQuotasQuery(), &listResult)
	require.NoError(t, err)
	assert.Contains(t, listResult.ResourceQuotas, fixListResourceQuotasResponse())

	err = c.Do(fixResourceQuotasStatusQuery(), &statusResult)
	require.NoError(t, err)
	assert.False(t, statusResult.Exceeded)

}

func fixListResourceQuotasQuery() *graphql.Request {
	query := `query($environment: String!) {
				resourceQuotas(environment: $environment) {
					name
					pods
					limits {
					  memory
					  cpu
					}
					requests {
					  memory
					  cpu
					}
				}
			}`
	r := graphql.NewRequest(query)
	r.SetVar("environment", resourceQuotaNamespace)

	return r
}

func fixResourceQuotasStatusQuery() *graphql.Request {
	query := `query($environment: String!) {
				  resourceQuotasStatus(environment: $environment) {
					exceeded
				  }
				}`
	r := graphql.NewRequest(query)
	r.SetVar("environment", resourceQuotaNamespace)

	return r
}

func fixListResourceQuotasResponse() resourceQuota {
	return resourceQuota{
		Name: resourceQuotaName,
		Pods: "10",
		Limits: resourceValues{
			Cpu:    "900m",
			Memory: "1Gi",
		},
		Requests: resourceValues{
			Cpu:    "500m",
			Memory: "512Mi",
		},
	}
}

func fixNamespace(name string) *v1.Namespace {
	return &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func fixResourceQuota() *v1.ResourceQuota {
	return &v1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourceQuotaName,
			Namespace: resourceQuotaNamespace,
		},
		Spec: v1.ResourceQuotaSpec{
			Hard: v1.ResourceList{
				v1.ResourcePods:           resource.MustParse("10"),
				v1.ResourceLimitsCPU:      resource.MustParse("900m"),
				v1.ResourceLimitsMemory:   resource.MustParse("1Gi"),
				v1.ResourceRequestsCPU:    resource.MustParse("500m"),
				v1.ResourceRequestsMemory: resource.MustParse("512Mi"),
			},
		},
	}
}
