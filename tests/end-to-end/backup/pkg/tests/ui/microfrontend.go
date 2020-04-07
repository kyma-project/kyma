package ui

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	uiV1alpha1v "github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
	mfClient "github.com/kyma-project/kyma/common/microfrontend-client/pkg/client/clientset/versioned"

	"github.com/kyma-project/kyma/tests/end-to-end/backup/pkg/config"
)

type microfrontendTest struct {
	microforntendName string
	mfClient          *mfClient.Clientset
	coreClient        *kubernetes.Clientset
}

func NewMicrofrontendTest() (microfrontendTest, error) {
	restConfig, err := config.NewRestClientConfig()
	if err != nil {
		return microfrontendTest{}, err
	}

	mfClient, err := mfClient.NewForConfig(restConfig)
	if err != nil {
		return microfrontendTest{}, err
	}

	coreClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return microfrontendTest{}, err
	}

	return microfrontendTest{
		mfClient:          mfClient,
		coreClient:        coreClient,
		microforntendName: "test-mf",
	}, nil
}

func (mft microfrontendTest) CreateResources(t *testing.T, namespace string) {
	_, err := mft.createMicrofrontend(namespace)
	require.NoError(t, err)
}

func (mft microfrontendTest) TestResources(t *testing.T, namespace string) {
	mfs, err := mft.getMicrofrontends(t, namespace, 2*time.Minute)
	require.NoError(t, err)
	mfValue := mfs.Items[0]
	require.Equal(t, mfValue.Name, mft.microforntendName)
	require.Equal(t, mfValue.Spec.Category, "Test")
	require.Equal(t, mfValue.Spec.Version, "1")
	require.Equal(t, mfValue.Spec.ViewBaseURL, "https://test.kyma.cx/mf-test")

	require.Len(t, mfValue.Spec.NavigationNodes, 1)
	navNode := mfValue.Spec.NavigationNodes[0]

	require.Equal(t, navNode.Label, "testMF")
	require.Equal(t, navNode.NavigationPath, "path")
	require.Equal(t, navNode.ViewURL, "/resourcePath")
	require.Equal(t, navNode.ShowInNavigation, true)
}

func (mft microfrontendTest) createMicrofrontend(namespace string) (*uiV1alpha1v.MicroFrontend, error) {
	microfrontend := &uiV1alpha1v.MicroFrontend{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ui.kyma-project.io/v1alpha1",
			Kind:       "MicroFrontend",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: mft.microforntendName,
		},
		Spec: uiV1alpha1v.MicroFrontendSpec{
			CommonMicroFrontendSpec: uiV1alpha1v.CommonMicroFrontendSpec{
				Version:     "1",
				Category:    "Test",
				ViewBaseURL: "https://test.kyma.cx/mf-test",
				NavigationNodes: []uiV1alpha1v.NavigationNode{
					uiV1alpha1v.NavigationNode{
						Label:               "testMF",
						NavigationPath:      "path",
						ViewURL:             "/resourcePath",
						ShowInNavigation:    true,
						RequiredPermissions: []uiV1alpha1v.RequiredPermission{},
					},
				},
			},
		},
	}
	return mft.mfClient.UiV1alpha1().MicroFrontends(namespace).Create(microfrontend)
}

func (mft microfrontendTest) getMicrofrontends(t *testing.T, namespace string, waitmax time.Duration) (*uiV1alpha1v.MicroFrontendList, error) {
	mfs, err := mft.mfClient.UiV1alpha1().MicroFrontends(namespace).List(metav1.ListOptions{})
	require.NoError(t, err)
	if 1 != int32(len(mfs.Items)) {
		return nil, fmt.Errorf("Expected only one microfrontend, but got %v", len(mfs.Items))
	}
	return mfs, err
}
