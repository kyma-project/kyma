package webhook

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"testing"
	"time"

	// "k8s.io/apimachinery/pkg/types"

	"gomodules.xyz/jsonpatch/v2"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/onsi/gomega"
	"golang.org/x/net/context"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	authenticationv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var c client.Client

const timeout = time.Second * 10

var webhookURL = fmt.Sprintf("https://localhost:%d/%s", port, webhookEndpoint)

var fnConfig = &corev1.ConfigMap{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "fn-config",
		Namespace: "default",
	},
	Data: map[string]string{
		"dockerRegistry":     "test",
		"serviceAccountName": "build-bot",
		"defaults": `{
			"size": "S",
			"runtime": "nodejs8",
			"timeOut": 180,
			"funcContentType": "plaintext",
		}`,
		"runtimes": `[
			{
				"ID": "nodejs8",
				"dockerfileName": "dockerfile-nodejs8",
			},
			{
				"ID": "nodejs6",
				"dockerfileName": "dockerfile-nodejs6",
			}
		]`,
		"funcSizes": `[
			{"size": "S"},
			{"size": "M"},
			{"size": "L"},
		]`,
		"funcTypes": `[
			{"type": "plaintext"},
			{"type": "base64"}
		]`,
	},
}

// Integration test for webhook
// Spin up the webhook server and issue an admission request against it with an invalid function
func TestWebHook(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	g.Expect(createCertificates(t)).NotTo(gomega.HaveOccurred())

	// create manager
	mgr, err := manager.New(cfg, manager.Options{
		MetricsBindAddress: "0",
	})
	g.Expect(err).NotTo(gomega.HaveOccurred())

	c = mgr.GetClient()

	// add webhook to manager
	Add(mgr)

	// start manager
	stopMgr, mgrStopped := StartTestManager(mgr, g)
	defer func() {
		close(stopMgr)
		mgrStopped.Wait()
	}()

	g.Expect(c.Create(context.TODO(), fnConfig)).NotTo(gomega.HaveOccurred())

	testInvalidFunc(t)
	testHandleDefaults(t)
}

func testInvalidFunc(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
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
	g.Expect(response.Response.Result.Message).NotTo(gomega.BeEmpty())
	g.Expect(response.Response.Result.Code).To(gomega.Equal(int32(400)))
}

// We are using a certificate signed by an unknown CA
// Since this is just a local integration test, we ignore TLS verifcation
func getInsecureClient() *http.Client {
	insecureTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &http.Client{Transport: insecureTransport}
}

func testHandleDefaults(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// these are the expected patches for an empty function
	g.Expect(createCertificates(t)).NotTo(gomega.HaveOccurred())

	expectedPatches := []jsonpatch.Operation{
		{
			Operation: "add",
			Path:      "/spec/functionContentType",
			Value:     "plaintext",
		},
		{
			Operation: "add",
			Path:      "/spec/size",
			Value:     "S",
		},
		{
			Operation: "add",
			Path:      "/spec/runtime",
			Value:     "nodejs8",
		},
		{
			Operation: "add",
			Path:      "/spec/timeout",
			Value:     float64(180),
		},
		{
			Operation: "add",
			Path:      "/status",
			Value:     map[string]interface{}{},
		},
	}

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
			Name:      "foo",
			Operation: admissionv1beta1.Create,
			UserInfo:  authenticationv1.UserInfo{},
			Object: runtime.RawExtension{
				Raw: []byte(`
				{
					"metadata": {
						"name": "foo",
						"namespace": "default",
						"uid": "e9137d7d-c318-12e8-bbad-025654321111",
						"creationTimestamp": "2019-06-07T12:33:39Z"
					},
					"spec": {
						"function": "foo()"
					}
				}`),
			},
		},
		Response: &admissionv1beta1.AdmissionResponse{},
	}

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
	g.Expect(response.Response.Allowed).To(gomega.BeTrue())

	responsePatch := []jsonpatch.Operation{}
	err = json.Unmarshal(response.Response.Patch, &responsePatch)
	if err != nil {
		t.Errorf("Error while unmarshalling response patch: %v", err)
	}

	// check that each received patch matches at least one expected patch
	for _, expectedPatch := range expectedPatches {
		g.Expect(responsePatch).To(gomega.ContainElement(gomega.BeEquivalentTo(expectedPatch)))
	}

	// check that each received patch matches at least one expected patch
	// for _, actualPatch := range responsePatch {
	// 	g.Expect(expectedPatches).To(gomega.ContainElement(gomega.BeEquivalentTo(actualPatch)))
	// }
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
				Group:   "serverless.kyma-project.io",
				Kind:    "Function",
				Version: "v1alpha1",
			},
			Resource: metav1.GroupVersionResource{
				Group:    "serverless.kyma-project.io",
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
	var srcCaCrt *os.File
	var srcTlsCrt *os.File
	var srcTlsKey *os.File
	var destCaCrt *os.File
	var destTlsCrt *os.File
	var destTlsKey *os.File

	dir := "/tmp/k8s-webhook-server/serving-certs"

	// create directory if not existing yet
	_ = os.Mkdir("/tmp/k8s-webhook-server", os.ModePerm)
	_ = os.Mkdir(dir, os.ModePerm)

	// open src files
	if srcCaCrt, err = os.Open("../../test/certs/ca.crt"); err != nil {
		return err
	}
	defer srcCaCrt.Close()
	if srcTlsCrt, err = os.Open("../../test/certs/tls.crt"); err != nil {
		return err
	}
	defer srcTlsCrt.Close()
	if srcTlsKey, err = os.Open("../../test/certs/tls.key"); err != nil {
		return err
	}
	defer srcTlsKey.Close()

	// open dest files
	if destCaCrt, err = os.Create(fmt.Sprintf("%s/%s", dir, "ca.crt")); err != nil {
		return err
	}
	defer destCaCrt.Close()
	if destTlsCrt, err = os.Create(fmt.Sprintf("%s/%s", dir, "tls.crt")); err != nil {
		return err
	}
	defer destTlsCrt.Close()
	if destTlsKey, err = os.Create(fmt.Sprintf("%s/%s", dir, "tls.key")); err != nil {
		return err
	}
	defer destTlsKey.Close()

	// copy ca.crt
	if _, err := io.Copy(destCaCrt, srcCaCrt); err != nil {
		return err
	}
	// copy tls.crt
	if _, err := io.Copy(destTlsCrt, srcTlsCrt); err != nil {
		return err
	}
	// copy tls.key
	if _, err := io.Copy(destTlsKey, srcTlsKey); err != nil {
		return err
	}

	return nil
}
