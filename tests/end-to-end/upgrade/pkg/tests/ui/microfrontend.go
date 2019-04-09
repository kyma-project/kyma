package ui

import (
	"net/http"
	"strings"
	"time"

	"crypto/tls"
	"net/http/cookiejar"

	"github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
	mfClient "github.com/kyma-project/kyma/common/microfrontend-client/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MicrofrontendUpgradeTest tests the creation of a kubeless microfrontend and execute a http request to the exposed api of the microfrontend after Kyma upgrade phase
type MicrofrontendUpgradeTest struct {
	microfrontendName string
	namespace         string
	stop              <-chan struct{}
	httpClient        *http.Client
	mfClient          *mfClient.Clientset
}

// NewMicrofrontendUpgradeTest returns new instance of the MicrofrontendUpgradeTest
func NewMicrofrontendUpgradeTest(mfClient *mfClient.Clientset) *MicrofrontendUpgradeTest {
	namespace := strings.ToLower("MicrofrontendUpgradeTest")
	httpCli, err := getHTTPClient(true)
	if err != nil {
		logrus.Fatal(errors.Wrap(err, "failed on getting the http client."))
	}
	return &MicrofrontendUpgradeTest{
		microfrontendName: "mf-name",
		namespace:         namespace,
		httpClient:        httpCli,
		mfClient:          mfClient,
	}
}

// CreateResources creates resources needed for e2e upgrade test
func (t *MicrofrontendUpgradeTest) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	log.Println("MicrofrontendUpgradeTest creating resources")
	t.namespace = namespace
	t.stop = stop

	err := t.createMicrofrontend()
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

	_, err := t.mfClient.UiV1alpha1().MicroFrontends(t.namespace).Get(t.microfrontendName, metav1.GetOptions{})

	if err != nil {
		return errors.Wrapf(err, "while checking if microfrontend %q still exists", t.microfrontendName)
	}

	return nil
}

func (t MicrofrontendUpgradeTest) createMicrofrontend() error {
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
						NavigationPath:   "/test/path",
						ViewURL:          "/resourcePath",
						ShowInNavigation: true,
					},
				},
			},
		},
	}
	_, err := t.mfClient.UiV1alpha1().MicroFrontends(t.namespace).Create(microfrontend)
	return err
}

func getHTTPClient(skipVerify bool) (*http.Client, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipVerify},
	}

	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	return &http.Client{Timeout: 15 * time.Second, Transport: tr, Jar: cookieJar}, nil
}

func int32Ptr(i int32) *int32 { return &i }
