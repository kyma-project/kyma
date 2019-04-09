package ui

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"crypto/tls"
	"net/http/cookiejar"

	"github.com/google/uuid"
	"github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"
	mfClient "github.com/kyma-project/kyma/common/microfrontend-client/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// MicrofrontendUpgradeTest tests the creation of a kubeless microfrontend and execute a http request to the exposed api of the microfrontend after Kyma upgrade phase
type MicrofrontendUpgradeTest struct {
	microfrontendName, uuid string
	coreClient              kubernetes.Interface
	namespace               string
	stop                    <-chan struct{}
	httpClient              *http.Client
	mfClient                *mfClient.Clientset
	hostName                string
}

// NewMicrofrontendUpgradeTest returns new instance of the MicrofrontendUpgradeTest
func NewMicrofrontendUpgradeTest(k8sCli kubernetes.Interface) *MicrofrontendUpgradeTest {
	domainName := os.Getenv("DOMAIN")
	if len(domainName) == 0 {
		logrus.Fatal("Environment variable DOMAIN is not found.")
	}
	namespace := strings.ToLower("MicrofrontendUpgradeTest")
	hostName := fmt.Sprintf("%s-%s.%s", "hello", namespace, domainName)
	httpCli, err := getHTTPClient(true)
	if err != nil {
		logrus.Fatal(errors.Wrap(err, "failed on getting the http client."))
	}
	return &MicrofrontendUpgradeTest{
		coreClient:        k8sCli,
		microfrontendName: "hello",
		uuid:              uuid.New().String(),
		namespace:         namespace,
		httpClient:        httpCli,
		hostName:          hostName,
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

	host := fmt.Sprintf("https://%s", t.hostName)

	value, err := t.getMicrofrontendOutput(host, 1*time.Minute, log)
	if err != nil {
		return errors.Wrapf(err, "failed request to host %s.", host)
	}

	if !strings.Contains(value, t.uuid) {
		return fmt.Errorf("could not get expected microfrontend output:\n %v\n output:\n %v", t.uuid, value)
	}

	return nil
}

func (t *MicrofrontendUpgradeTest) getMicrofrontendOutput(host string, waitmax time.Duration, log logrus.FieldLogger) (string, error) {
	log.Println("MicrofrontendUpgradeTest microfrontend output")
	log.Printf("\nHost: %s", host)

	tick := time.Tick(2 * time.Second)
	timeout := time.After(waitmax)
	messages := ""

	for {
		select {
		case <-tick:

			resp, err := t.httpClient.Post(host, "text/plain", bytes.NewBufferString(t.uuid))
			if err != nil {
				messages += fmt.Sprintf("%+v\n", err)
				break
			}
			if resp.StatusCode == http.StatusOK {
				bodyBytes, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return "", err
				}
				return string(bodyBytes), nil
			}
			messages += fmt.Sprintf("%+v", err)

		case <-timeout:
			return "", fmt.Errorf("could not get microfrontend output:\n %v", messages)
		case <-t.stop:
			return "", fmt.Errorf("can't be possible to get a response from the http request to the microfrontend")
		}
	}

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
