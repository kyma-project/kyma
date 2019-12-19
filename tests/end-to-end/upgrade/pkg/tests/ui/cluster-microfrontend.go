package ui

import (
	"github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
	mfClient "github.com/kyma-project/kyma/common/microfrontend-client/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterMicrofrontendUpgradeTest tests the creation of a cluster microfrontend
type ClusterMicrofrontendUpgradeTest struct {
	clusterMicrofrontendName string
	stop                     <-chan struct{}
	mfClient                 *mfClient.Clientset
}

// NewClusterMicrofrontendUpgradeTest returns new instance of the ClusterMicrofrontendUpgradeTest
func NewClusterMicrofrontendUpgradeTest(mfClient *mfClient.Clientset) *ClusterMicrofrontendUpgradeTest {
	return &ClusterMicrofrontendUpgradeTest{
		clusterMicrofrontendName: "cmf-name",
		mfClient:                 mfClient,
	}
}

// CreateResources creates resources needed for e2e upgrade test
func (t *ClusterMicrofrontendUpgradeTest) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	log.Println("ClusterMicrofrontendUpgradeTest creating resources")
	t.stop = stop

	err := t.createClusterMicrofrontend()
	if err != nil {
		return err
	}

	// Ensure resources works
	err = t.TestResources(stop, log, namespace)
	if err != nil {
		return errors.Wrap(err, "first call to TestResources() failed.")
	}
	return nil
}

// TestResources tests resources after the upgrade test
func (t *ClusterMicrofrontendUpgradeTest) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	log.Println("ClusterMicrofrontendUpgradeTest testing resources")
	t.stop = stop

	_, err := t.mfClient.UiV1alpha1().ClusterMicroFrontends().Get(t.clusterMicrofrontendName, metav1.GetOptions{})

	if err != nil {
		return errors.Wrapf(err, "while checking if cluster microfrontend %q still exists", t.clusterMicrofrontendName)
	}

	return nil
}

func (t *ClusterMicrofrontendUpgradeTest) createClusterMicrofrontend() error {
	clusterMicrofrontend := &v1alpha1.ClusterMicroFrontend{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ui.kyma-project.io/v1alpha1",
			Kind:       "ClusterMicroFrontend",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: t.clusterMicrofrontendName,
		},
		Spec: v1alpha1.ClusterMicroFrontendSpec{
			Placement: "cluster",
			CommonMicroFrontendSpec: v1alpha1.CommonMicroFrontendSpec{
				Version:     "1",
				Category:    "Test",
				ViewBaseURL: "https://test.kyma.cx/mf-test",
				NavigationNodes: []v1alpha1.NavigationNode{
					v1alpha1.NavigationNode{
						Label:            "testCMF",
						NavigationPath:   "test",
						ViewURL:          "/resourcePath",
						ShowInNavigation: true,
						RequiredPermissions: []v1alpha1.RequiredPermission{},
					},
					v1alpha1.NavigationNode{
						Label:            "testCMF child",
						NavigationPath:   "child",
						ViewURL:          "/resourcePath/child",
						ShowInNavigation: true,
						RequiredPermissions: []v1alpha1.RequiredPermission{},
					},
				},
			},
		},
	}
	_, err := t.mfClient.UiV1alpha1().ClusterMicroFrontends().Create(clusterMicrofrontend)
	return err
}
