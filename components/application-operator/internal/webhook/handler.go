package webhook

import (
	"encoding/json"
	"fmt"
	"github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	alpha1 "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	v1alpha12 "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"net/http"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
)

type Handler struct {
	applicationClient v1alpha12.ApplicationInterface
}

type patch struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func (s Handler) handle(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		log.Errorf("contentType=%s, expect application/json", contentType)
		return
	}

	log.Info(fmt.Sprintf("handling request: %s", body))

	// The AdmissionReview that was sent to the webhook
	requestedAdmissionReview := v1beta1.AdmissionReview{}

	// The AdmissionReview that will be returned
	responseAdmissionReview := v1beta1.AdmissionReview{}

	deserializer := codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(body, nil, &requestedAdmissionReview); err != nil {
		log.Error(err)
		responseAdmissionReview.Response = toAdmissionResponse(err)
	} else {
		responseAdmissionReview.Response = s.mutate(requestedAdmissionReview)
	}

	responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID

	log.Info(fmt.Sprintf("sending response: %v", responseAdmissionReview.Response))

	respBytes, err := json.Marshal(responseAdmissionReview)
	if err != nil {
		log.Error(err)
	}
	if _, err := w.Write(respBytes); err != nil {
		log.Error(err)
	}
}

func (s Handler) mutate(admissionReview v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	var mapping v1alpha1.ApplicationMapping

	if err := json.Unmarshal(admissionReview.Request.Object.Raw, &mapping); err != nil {
		log.Errorf("Could not unmarshal raw application mapping: %s", err.Error())
		return toAdmissionResponse(err)
	}

	application, err := s.applicationClient.Get(mapping.Name, metav1.GetOptions{})

	if err != nil {
		log.Errorf("Could not get Application %s: %s", mapping.Name, err.Error())
		return toAdmissionResponse(err)
	}

	patch, err := createPatch(application)

	if err != nil {
		log.Errorf("Failed to marshal patch: %s", err.Error())
		return toAdmissionResponse(err)
	}

	return &v1beta1.AdmissionResponse{
		Allowed: true,
		PatchType: func() *v1beta1.PatchType {
			pt := v1beta1.PatchTypeJSONPatch
			return &pt
		}(),
		Patch: patch,
	}
}

func createPatch(application *alpha1.Application) ([]byte, error) {
	controller := true

	reference := metav1.OwnerReference{
		APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		Kind:       "Application",
		Name:       application.Name,
		UID:        application.UID,
		Controller: &controller,
	}

	patch := patch{
		Op:    "add",
		Path:  "/metadata/ownerReferences",
		Value: reference,
	}

	marshal, err := json.Marshal(patch)

	if err != nil {
		return nil, err
	}

	return marshal, nil
}

func toAdmissionResponse(err error) *v1beta1.AdmissionResponse {
	return &v1beta1.AdmissionResponse{
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}
