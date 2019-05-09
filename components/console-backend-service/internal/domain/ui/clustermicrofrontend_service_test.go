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

func TestClusterMicrofrontendService_List(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ClusterMicrofrontend1 := fixClusterMicrofrontend("1")
		ClusterMicrofrontend2 := fixClusterMicrofrontend("2")
		ClusterMicrofrontend3 := fixClusterMicrofrontend("3")

		ClusterMicrofrontendInformer := fixCMFInformer(ClusterMicrofrontend1, ClusterMicrofrontend2, ClusterMicrofrontend3)
		svc := ui.NewClusterMicrofrontendService(ClusterMicrofrontendInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, ClusterMicrofrontendInformer)

		clustermicrofrontends, err := svc.List()
		require.NoError(t, err)
		assert.Contains(t, clustermicrofrontends, ClusterMicrofrontend1)
		assert.Contains(t, clustermicrofrontends, ClusterMicrofrontend2)
		assert.Contains(t, clustermicrofrontends, ClusterMicrofrontend3)
	})

	t.Run("NotFound", func(t *testing.T) {
		ClusterMicrofrontendInformer := fixCMFInformer()

		svc := ui.NewClusterMicrofrontendService(ClusterMicrofrontendInformer)

		testingUtils.WaitForInformerStartAtMost(t, time.Second, ClusterMicrofrontendInformer)

		var emptyArray []*v1alpha1.ClusterMicroFrontend
		clustermicrofrontends, err := svc.List()
		require.NoError(t, err)
		assert.Equal(t, emptyArray, clustermicrofrontends)
	})
}

func fixCMFInformer(objects ...runtime.Object) cache.SharedIndexInformer {
	client := fake.NewSimpleClientset(objects...)
	informerFactory := externalversions.NewSharedInformerFactory(client, 0)

	informer := informerFactory.Ui().V1alpha1().ClusterMicroFrontends().Informer()

	return informer
}

func fixClusterMicrofrontend(name string) *v1alpha1.ClusterMicroFrontend {
	version := "v1"
	category := "test-category"
	viewBaseUrl := "http://test-viewBaseUrl.com"
	placement := "cluster"
	navigationNodes := []v1alpha1.NavigationNode{
		v1alpha1.NavigationNode{
			Label:            "test-mf",
			NavigationPath:   "test-path",
			ViewURL:          "/test/viewUrl",
			ShowInNavigation: true,
		},
	}

	item := v1alpha1.ClusterMicroFrontend{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.ClusterMicroFrontendSpec{
			Placement: placement,
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
