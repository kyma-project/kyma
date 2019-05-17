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

func TestMicrofrontendService_List(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		namespace := "test-namespace"
		Microfrontend1 := fixMicrofrontend("1")
		Microfrontend2 := fixMicrofrontend("2")
		Microfrontend3 := fixMicrofrontend("3")

		MicrofrontendInformer := fixMFInformer(Microfrontend1, Microfrontend2, Microfrontend3)
		svc := ui.NewMicrofrontendService(MicrofrontendInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, MicrofrontendInformer)

		microfrontends, err := svc.List(namespace)
		require.NoError(t, err)
		assert.Contains(t, microfrontends, Microfrontend1)
		assert.Contains(t, microfrontends, Microfrontend2)
		assert.Contains(t, microfrontends, Microfrontend3)
	})

	t.Run("NotFound", func(t *testing.T) {
		namespace := "test-namespace"
		MicrofrontendInformer := fixMFInformer()

		svc := ui.NewMicrofrontendService(MicrofrontendInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, MicrofrontendInformer)

		var emptyArray []*v1alpha1.MicroFrontend
		microfrontends, err := svc.List(namespace)
		require.NoError(t, err)
		assert.Equal(t, emptyArray, microfrontends)
	})
}

func fixMFInformer(objects ...runtime.Object) cache.SharedIndexInformer {
	client := fake.NewSimpleClientset(objects...)
	informerFactory := externalversions.NewSharedInformerFactory(client, 0)

	informer := informerFactory.Ui().V1alpha1().MicroFrontends().Informer()

	return informer
}

func fixMicrofrontend(name string) *v1alpha1.MicroFrontend {
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
