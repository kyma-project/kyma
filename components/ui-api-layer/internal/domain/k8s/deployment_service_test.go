package k8s_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/api/apps/v1beta2"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/k8s"
	testingUtils "github.com/kyma-project/kyma/components/ui-api-layer/internal/testing"
)

func TestDeploymentService_List(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deployment1 := fixDeployment("one", "env1", "deployment")
		deployment2 := fixDeployment("two", "env1", "function")
		deployment3 := fixDeployment("three", "env2", "deployment")
		deployment4 := fixDeployment("four", "env2", "function")

		informer := fixDeploymentInformer(deployment1, deployment2, deployment3, deployment4)
		svc := k8s.NewDeploymentService(informer)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := svc.List("env1")

		require.NoError(t, err)
		assert.Equal(t, []*v1beta2.Deployment{
			deployment1, deployment2,
		}, result)
	})

	t.Run("Not found", func(t *testing.T) {
		var expected []*v1beta2.Deployment

		informer := fixDeploymentInformer()
		svc := k8s.NewDeploymentService(informer)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := svc.List("env1")

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})
}

func TestDeploymentService_ListWithoutFunctions(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deployment1 := fixDeployment("one", "env1", "deployment")
		deployment2 := fixDeployment("two", "env1", "function")
		deployment3 := fixDeployment("three", "env2", "deployment")
		deployment4 := fixDeployment("four", "env2", "function")

		informer := fixDeploymentInformer(deployment1, deployment2, deployment3, deployment4)
		svc := k8s.NewDeploymentService(informer)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := svc.ListWithoutFunctions("env1")

		require.NoError(t, err)
		assert.Equal(t, []*v1beta2.Deployment{
			deployment1,
		}, result)
	})

	t.Run("Not found", func(t *testing.T) {
		var expected []*v1beta2.Deployment

		informer := fixDeploymentInformer()
		svc := k8s.NewDeploymentService(informer)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := svc.ListWithoutFunctions("env1")

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})
}

func fixDeployment(name, environment, kind string) *v1beta2.Deployment {
	return &v1beta2.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: environment,
			Labels: map[string]string{
				kind: "",
			},
		},
	}
}

func fixDeploymentInformer(objects ...runtime.Object) cache.SharedIndexInformer {
	client := fake.NewSimpleClientset(objects...)
	informerFactory := informers.NewSharedInformerFactory(client, 0)

	return informerFactory.Apps().V1beta2().Deployments().Informer()
}
