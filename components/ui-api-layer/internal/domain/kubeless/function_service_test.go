package kubeless_test

import (
	"testing"
	"time"

	"github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/kubeless/kubeless/pkg/client/clientset/versioned/fake"
	"github.com/kubeless/kubeless/pkg/client/informers/externalversions"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/kubeless"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	testingUtils "github.com/kyma-project/kyma/components/ui-api-layer/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
)

func TestFunctionService_List(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		function1 := fixFunction("f1", "env1")
		function2 := fixFunction("f2", "env1")
		function3 := fixFunction("f3", "env2")

		informer := fixInformer(function1, function2, function3)

		svc := kubeless.NewFunctionService(informer)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := svc.List("env1", pager.PagingParams{})

		require.NoError(t, err)
		assert.Equal(t, []*v1beta1.Function{
			function1, function2,
		}, result)
	})

	t.Run("Not found", func(t *testing.T) {
		informer := fixInformer()
		var expected []*v1beta1.Function

		svc := kubeless.NewFunctionService(informer)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := svc.List("env1", pager.PagingParams{})

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})
}

func fixFunction(name, environment string) *v1beta1.Function {
	return &v1beta1.Function{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: environment,
		},
	}
}

func fixInformer(objects ...runtime.Object) cache.SharedIndexInformer {
	client := fake.NewSimpleClientset(objects...)
	informerFactory := externalversions.NewSharedInformerFactory(client, 0)

	return informerFactory.Kubeless().V1beta1().Functions().Informer()
}
