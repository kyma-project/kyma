package backupe2e

import (
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	. "github.com/smartystreets/goconvey/convey"

	uiV1alpha1v "github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
	mfClient "github.com/kyma-project/kyma/common/microfrontend-client/pkg/client/clientset/versioned"
)

type microfrontendTest struct {
	microforntendName, uuid string
	mfClient                *mfClient.Clientset
	coreClient              *kubernetes.Clientset
}

func NewMicrofrontendTest() (microfrontendTest, error) {

	kubeconfig := os.Getenv("KUBECONFIG")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return microfrontendTest{}, err
	}

	mfClient, err := mfClient.NewForConfig(config)
	if err != nil {
		return microfrontendTest{}, err
	}

	coreClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return microfrontendTest{}, err
	}
	return microfrontendTest{
		mfClient:          mfClient,
		coreClient:        coreClient,
		microforntendName: "test-mf",
		uuid:              uuid.New().String(),
	}, nil
}

func (t microfrontendTest) CreateResources(namespace string) {
	_, err := t.createMicrofrontend(namespace)
	So(err, ShouldBeNil)
}

func (t microfrontendTest) TestResources(namespace string) {
	mfs, err := t.getMicrofrontends(namespace, 2*time.Minute)
	So(err, ShouldBeNil)
	mfValue := mfs.Items[0]
	So(mfValue.Name, ShouldEqual, t.microforntendName)
	So(mfValue.Spec.Category, ShouldEqual, "Test")
	So(mfValue.Spec.Version, ShouldEqual, "1")
	So(mfValue.Spec.ViewBaseURL, ShouldEqual, "/mf-test")

	So(len(mfValue.Spec.NavigationNodes), ShouldEqual, 1)
	navNode := mfValue.Spec.NavigationNodes[0]

	So(navNode.Label, ShouldEqual, "testMF")
	So(navNode.NavigationPath, ShouldEqual, "/test/path")
	So(navNode.ViewURL, ShouldEqual, "resourcePath")
	So(navNode.ShowInNavigation, ShouldEqual, true)

}

func (t microfrontendTest) createMicrofrontend(namespace string) (*uiV1alpha1v.MicroFrontend, error) {
	microfrontend := &uiV1alpha1v.MicroFrontend{
		ObjectMeta: metav1.ObjectMeta{
			Name: t.microforntendName,
		},
		Spec: uiV1alpha1v.MicroFrontendSpec{
			CommonMicroFrontendSpec: uiV1alpha1v.CommonMicroFrontendSpec{
				Version:     "1",
				Category:    "Test",
				ViewBaseURL: "/mf-test",
				NavigationNodes: []uiV1alpha1v.NavigationNode{
					uiV1alpha1v.NavigationNode{
						Label:            "testMF",
						NavigationPath:   "/test/path",
						ViewURL:          "/resourcePath",
						ShowInNavigation: true,
					},
				},
			},
		},
	}
	return t.mfClient.UiV1alpha1().MicroFrontends(namespace).Create(microfrontend)
}

func (t microfrontendTest) getMicrofrontends(namespace string, waitmax time.Duration) (*uiV1alpha1v.MicroFrontendList, error) {
	timeout := time.After(waitmax)
	tick := time.Tick(2 * time.Second)
	for {
		select {
		case <-timeout:
			mfs, err := t.mfClient.UiV1alpha1().MicroFrontends(namespace).List(metav1.ListOptions{})
			if err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("Microfrontend not availiable within given time  %v: %+v", waitmax, mfs)
		case <-tick:
			mfs, err := t.mfClient.UiV1alpha1().MicroFrontends(namespace).List(metav1.ListOptions{})
			So(err, ShouldBeNil)
			if 1 != int32(len(mfs.Items)) {
				return nil, fmt.Errorf("Expected only one microfrontend, but got %v", len(mfs.Items))
			}
			return mfs, err
		}
	}
}
