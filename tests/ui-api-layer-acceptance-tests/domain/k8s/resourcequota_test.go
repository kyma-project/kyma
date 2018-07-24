// +build acceptance

package k8s

import (
	"fmt"
	"testing"
	"time"

	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/graphql"
	"github.com/kyma-project/kyma/tests/ui-api-layer-acceptance-tests/k8s"
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
	ResourceQuotas []resourceQuota
}

type resourceQuota struct {
	Name     string
	Pods     string
	Limits   resourceValues
	Requests resourceValues
}

type resourceValues struct {
	Memory string `json:"memory"`
	Cpu    string `json:"cpu"`
}

func TestResourceQuotaQuery(t *testing.T) {
	// GIVEN
	k8sClient, _, err := k8s.NewClientWithConfig()
	require.NoError(t, err)

	_, err = k8sClient.Namespaces().Create(fixNamespace(resourceQuotaNamespace))
	require.NoError(t, err)
	defer func() {
		err := k8sClient.Namespaces().Delete(resourceQuotaNamespace, &metav1.DeleteOptions{})
		require.NoError(t, err)
	}()

	_, err = k8sClient.ResourceQuotas(resourceQuotaNamespace).Create(fixResourceQuota())
	require.NoError(t, err)

	query := fmt.Sprintf(`
	query {
	  resourceQuotas(environment: "%s") {
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
	}`, resourceQuotaNamespace)
	c, err := graphql.New()
	require.NoError(t, err)

	waiter.WaitAtMost(func() (bool, error) {
		_, err := k8sClient.ResourceQuotas(resourceQuotaNamespace).Get(resourceQuotaName, metav1.GetOptions{})
		if err == nil {
			return true, nil
		}
		return false, err
	}, time.Minute)

	// WHEN
	var result resourceQuotas
	err = c.DoQuery(query, &result)

	// THEN
	require.NoError(t, err)
	assert.Contains(t, result.ResourceQuotas, resourceQuota{
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
	})
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
