package defaultserver

import (
	"bytes"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/runtime"

	"io"
	"net/http"
	"net/http/httputil"
	"os"

	"crypto/tls"
	"encoding/json"

	"github.com/onsi/gomega"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var c client.Client

const timeout = time.Second * 10
const webhookURL = "https://localhost:9876/mutating-create-function"

// Integration test for webhook
// Spin up the webhook server and issue an admission request against it with an invalid function
func TestWebHook(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	g.Expect(createCertificates(t)).NotTo(gomega.HaveOccurred())

	// create manager
	mgr, err := manager.New(cfg, manager.Options{})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	// add webhook to manager
	g.Expect(Add(mgr)).NotTo(gomega.HaveOccurred())

	// start manager
	stopMgr, mgrStopped := StartTestManager(mgr, g)
	defer func() {
		close(stopMgr)
		mgrStopped.Wait()
	}()

	// create admission request
	var admissionReview = getAdmissionRequest()
	jsonRequest, err := json.Marshal(admissionReview)
	if err != nil {
		t.Errorf("Error encoding admission review: %v", err)
	}

	// wait for webserver to be reachable
	g.Eventually(func() error {
		_, err := getInsecureClient().Post(webhookURL, "application/json", bytes.NewBuffer(jsonRequest))
		return err
	}, timeout).Should(gomega.Succeed())

	// get admission response from webhook
	admissionRequest, err := getInsecureClient().Post(webhookURL, "application/json", bytes.NewBuffer(jsonRequest))
	if err != nil {
		t.Errorf("Could not get result from webhook: %v", err)
	}
	defer admissionRequest.Body.Close()
	g.Expect(admissionRequest.StatusCode).To(gomega.Equal(200))

	// print admission request
	byts, _ := httputil.DumpResponse(admissionRequest, true)
	t.Logf("response: %v", string(byts))

	// decode admission review
	var response admissionv1beta1.AdmissionReview
	g.Expect(json.NewDecoder(admissionRequest.Body).Decode(&response)).NotTo(gomega.HaveOccurred())

	// ensure function got rejected
	g.Expect(response.Response.Allowed).To(gomega.BeFalse())
	// due to invalid function size
	g.Expect(response.Response.Result.Message).To(gomega.Equal("size should be one of 'S,M,L,XL'"))
}

// We are using a certificate signed by an unknown CA
// Since this is just a local integration test, we ignore TLS verifcation
func getInsecureClient() *http.Client {
	insecureTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &http.Client{Transport: insecureTransport}
}

// Create an admission request
func getAdmissionRequest() admissionv1beta1.AdmissionReview {
	var admissionReview = admissionv1beta1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admission.k8s.io/v1beta1",
			Kind:       "AdmissionReview",
		},
		Request: &admissionv1beta1.AdmissionRequest{
			UID: "e9137d7d-c318-12e8-bbad-025654321111",
			Kind: metav1.GroupVersionKind{
				Group:   "runtime.kyma-project.io",
				Kind:    "Function",
				Version: "v1alpha1",
			},
			Resource: metav1.GroupVersionResource{
				Group:    "runtime.kyma-project.io",
				Resource: "Functions",
				Version:  "v1alpha1",
			},
			Name:      "invalid-function-size",
			Operation: admissionv1beta1.Create,
			UserInfo:  authenticationv1.UserInfo{},
			Object: runtime.RawExtension{
				Raw: []byte(`
				{
					"metadata": {
						"name": "invalid-function",
						"uid": "e9137d7d-c318-12e8-bbad-025654321111",
						"creationTimestamp": "2019-06-07T12:33:39Z"
					},
					"spec": {
						"function": "",
						"functionContentType": "plaintext",
						"size": "foo"
					}
				}`),
			},
		},
		Response: &admissionv1beta1.AdmissionResponse{},
	}
	return admissionReview
}

// Create the certificates to be used by the webhook https server
// Certificates have been created with `mkcert`
func createCertificates(t *testing.T) error {
	var err error
	var srcFileKey *os.File
	var srcFileCert *os.File
	var destFileKey *os.File
	var destFileCert *os.File

	// create directory if not existing yet
	_ = os.Mkdir("/tmp/cert", os.ModePerm)

	// open src files
	if srcFileKey, err = os.Open("../../../test/certs/localhost+2-key.pem"); err != nil {
		return err
	}
	defer srcFileKey.Close()
	if srcFileCert, err = os.Open("../../../test/certs/localhost+2.pem"); err != nil {
		return err
	}
	defer srcFileCert.Close()

	// open dest files
	if destFileKey, err = os.Create("/tmp/cert/key.pem"); err != nil {
		return err
	}
	defer destFileKey.Close()
	if destFileCert, err = os.Create("/tmp/cert/cert.pem"); err != nil {
		return err
	}
	defer destFileCert.Close()

	// copy key
	if _, err := io.Copy(destFileKey, srcFileKey); err != nil {
		return err
	}
	// copy cert
	if _, err := io.Copy(destFileCert, srcFileCert); err != nil {
		return err
	}

	return nil
}
