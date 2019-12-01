package ui

import (
	"github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
	mfClient "github.com/kyma-project/kyma/common/microfrontend-client/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MicrofrontendUpgradeTest tests the creation of a microfrontend
type MicrofrontendUpgradeTest struct {
	microfrontendName string
	stop              <-chan struct{}
	mfClient          *mfClient.Clientset
}

// NewMicrofrontendUpgradeTest returns new instance of the MicrofrontendUpgradeTest
func NewMicrofrontendUpgradeTest(mfClient *mfClient.Clientset) *MicrofrontendUpgradeTest {
	return &MicrofrontendUpgradeTest{
		microfrontendName: "mf-name",
		mfClient:          mfClient,
	}
}

// CreateResources creates resources needed for e2e upgrade test
func (t *MicrofrontendUpgradeTest) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	log.Println("MicrofrontendUpgradeTest creating resources")
	t.stop = stop

	err := t.createMicrofrontend(namespace)
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
func (t *MicrofrontendUpgradeTest) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	log.Println("MicrofrontendUpgradeTest testing resources")
	t.stop = stop

	_, err := t.mfClient.UiV1alpha1().MicroFrontends(namespace).Get(t.microfrontendName, metav1.GetOptions{})

	if err != nil {
		return errors.Wrapf(err, "while checking if microfrontend %q still exists", t.microfrontendName)
	}

	return nil
}

func (t *MicrofrontendUpgradeTest) createMicrofrontend(namespace string) error {
	microfrontend := &v1alpha1.MicroFrontend{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ui.kyma-project.io/v1alpha1",
			Kind:       "MicroFrontend",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: t.microfrontendName,
		},
		Spec: v1alpha1.MicroFrontendSpec{
			CommonMicroFrontendSpec: v1alpha1.CommonMicroFrontendSpec{
				Version:     "1",
				Category:    "Test",
				ViewBaseURL: "https://test.kyma.cx/mf-test",
				NavigationNodes: []v1alpha1.NavigationNode{
					v1alpha1.NavigationNode{
						Label:            "testMF",
						NavigationPath:   "test",
						ViewURL:          "/resourcePath",
						ShowInNavigation: true,
						RequiredPermissions: []v1alpha1.RequiredPermission{},
					},
					v1alpha1.NavigationNode{
						Label:            "testMF Child",
						NavigationPath:   "test/child",
						ViewURL:          "/resourcePath/child",
						ShowInNavigation: true,
						RequiredPermissions: []v1alpha1.RequiredPermission{},
					},
				},
			},
		},
	}
	_, err := t.mfClient.UiV1alpha1().MicroFrontends(namespace).Create(microfrontend)
	return err
}
