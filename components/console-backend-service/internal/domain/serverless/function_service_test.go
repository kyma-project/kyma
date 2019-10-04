package serverless

import (
	"testing"
	"time"

	resourceFake "github.com/kyma-project/kyma/components/console-backend-service/internal/resource/fake"
	testingUtils "github.com/kyma-project/kyma/components/console-backend-service/internal/testing"
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFunctionService_List(t *testing.T) {
	functionA1 := fixFunction("a1", "a")
	functionA2 := fixFunction("a2", "a")
	functionB := fixFunction("b", "b")

	serviceFactory, err := resourceFake.NewFakeServiceFactory(v1alpha1.AddToScheme, functionA1, functionA2, functionB)
	require.NoError(t, err)

	service := newFunctionService(serviceFactory)
	testingUtils.WaitForInformerStartAtMost(t, time.Second, service.Informer)

	functions, err := service.List("a")
	require.NoError(t, err)
	assert.ElementsMatch(t, []*v1alpha1.Function{functionA1, functionA2}, functions)
}

func TestFunctionService_Delete(t *testing.T) {
	fixName := "a1"
	fixNamespace := "a"
	functionA1 := fixFunction(fixName, fixNamespace)

	serviceFactory, err := resourceFake.NewFakeServiceFactory(v1alpha1.AddToScheme, functionA1)
	require.NoError(t, err)

	service := newFunctionService(serviceFactory)
	testingUtils.WaitForInformerStartAtMost(t, time.Second, service.Informer)

	err = service.Delete(fixName, fixNamespace)
	require.NoError(t, err)

	_, err = service.Client.Namespace(fixNamespace).Get(fixName, v1.GetOptions{})
	assert.True(t, apiErrors.IsNotFound(err))
}

func fixFunction(name, namespace string) *v1alpha1.Function {
	return &v1alpha1.Function{
		TypeMeta: v1.TypeMeta{
			Kind:       "Function",
			APIVersion: "serverless.kyma-project.io/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}
