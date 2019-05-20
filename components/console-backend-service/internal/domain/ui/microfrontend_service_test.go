package ui_test

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/common/microfrontend-client/pkg/client/clientset/versioned/fake"
	"github.com/kyma-project/kyma/common/microfrontend-client/pkg/client/informers/externalversions"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/ui"
	testingUtils "github.com/kyma-project/kyma/components/console-backend-service/internal/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
)

func TestMicroFrontendService_List(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		namespace := "test-namespace"
		MicroFrontend1 := fixMicroFrontend("1")
		MicroFrontend2 := fixMicroFrontend("2")
		MicroFrontend3 := fixMicroFrontend("3")

		MicroFrontendInformer := fixMFInformer(MicroFrontend1, MicroFrontend2, MicroFrontend3)
		svc := ui.NewMicroFrontendService(MicroFrontendInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, MicroFrontendInformer)

		microFrontends, err := svc.List(namespace)
		require.NoError(t, err)
		assert.Contains(t, microFrontends, MicroFrontend1)
		assert.Contains(t, microFrontends, MicroFrontend2)
		assert.Contains(t, microFrontends, MicroFrontend3)
	})

	t.Run("NotFound", func(t *testing.T) {
		namespace := "test-namespace"
		MicroFrontendInformer := fixMFInformer()

		svc := ui.NewMicroFrontendService(MicroFrontendInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, MicroFrontendInformer)

		var emptyArray []*v1alpha1.MicroFrontend
		microFrontends, err := svc.List(namespace)
		require.NoError(t, err)
		assert.Equal(t, emptyArray, microFrontends)
	})
}

func fixMFInformer(objects ...runtime.Object) cache.SharedIndexInformer {
	client := fake.NewSimpleClientset(objects...)
	informerFactory := externalversions.NewSharedInformerFactory(client, 0)

	informer := informerFactory.Ui().V1alpha1().MicroFrontends().Informer()

	return informer
}

func fixMicroFrontend(name string) *v1alpha1.MicroFrontend {
	namespace := "test-namespace"
	version := "v1"
	category := "test-category"
	viewBaseUrl := "http://test-viewBaseUrl.com"
	navigationNodes := []v1alpha1.NavigationNode{
		v1alpha1.NavigationNode{
			Label:            "test-mf",
			NavigationPath:   "test-path",
			ViewURL:          "/test/viewUrl",
			ShowInNavigation: true,
		},
	}

	item := v1alpha1.MicroFrontend{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.MicroFrontendSpec{
			CommonMicroFrontendSpec: v1alpha1.CommonMicroFrontendSpec{
				Version:         version,
				Category:        category,
				ViewBaseURL:     viewBaseUrl,
				NavigationNodes: navigationNodes,
			},
		},
	}

	return &item
}
