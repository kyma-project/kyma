package k8s_test

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s"
	testingUtils "github.com/kyma-project/kyma/components/console-backend-service/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
)

func TestKymaVersionService_FindDeployment(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		deployment := fixDeploymentWithImage()

		informer := fixKymaVersionInformer(deployment)
		svc, err := k8s.NewKymaVersionService(informer)
		require.NoError(t, err)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := svc.FindDeployment()

		require.NoError(t, err)
		assert.Equal(t, deployment, result)
	})

	t.Run("NotFound", func(t *testing.T) {
		informer := fixKymaVersionInformer()
		svc, err := k8s.NewKymaVersionService(informer)
		require.NoError(t, err)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

		result, err := svc.FindDeployment()
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}


func fixDeploymentWithImage() *apps.Deployment {
	return &apps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kyma-installer",
			Namespace: "kyma-installer",
		},
		Spec: apps.DeploymentSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Image: "image",
						},
					},
				},
			},
		},
	}
}

func fixKymaVersionInformer(objects ...runtime.Object) cache.SharedIndexInformer {
	client := fake.NewSimpleClientset(objects...)
	informerFactory := informers.NewSharedInformerFactory(client, 0)

	return informerFactory.Apps().V1().Deployments().Informer()
}