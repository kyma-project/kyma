package mutating

import (
	"testing"

	"github.com/appscode/jsonpatch"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"

	"github.com/onsi/gomega/gstruct"

	runtimev1alpha1 "github.com/kyma-project/kyma/components/knative-function-controller/pkg/apis/runtime/v1alpha1"
	"github.com/onsi/gomega"

	"context"

	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

var functionCreateHandler = FunctionCreateHandler{}

// Test that an empty function gets all default values set
func TestMutation(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	function := &runtimev1alpha1.Function{
		ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "default"},
		Spec: runtimev1alpha1.FunctionSpec{
			FunctionContentType: "plaintext",
			Function:            "foo",
		},
	}

	// mutate function
	functionCreateHandler.mutatingFunctionFn(function)

	// ensure defaults are set
	g.Expect(function.Spec).To(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
		"Size":                gomega.BeEquivalentTo("S"),
		"Timeout":             gomega.BeEquivalentTo(180),
		"Runtime":             gomega.BeEquivalentTo("nodejs8"),
		"FunctionContentType": gomega.BeEquivalentTo("plaintext"),
	}))
}

// Test that all values get validated
func TestValidation(t *testing.T) {
	g := gomega.NewWithT(t)

	// wrong runtime
	function := &runtimev1alpha1.Function{
		Spec: runtimev1alpha1.FunctionSpec{
			FunctionContentType: "plaintext",
			Function:            "foo",
			Size:                "S",
			Runtime:             "nodejs4",
		},
	}
	g.Expect(functionCreateHandler.validateFunctionFn(function)).To(gomega.MatchError("runtime should be one of 'nodejs6,nodejs8'"))

	// wrong size
	function = &runtimev1alpha1.Function{
		Spec: runtimev1alpha1.FunctionSpec{
			FunctionContentType: "plaintext",
			Function:            "foo",
			Size:                "UnknownSize",
			Runtime:             "nodejs8",
		},
	}
	g.Expect(functionCreateHandler.validateFunctionFn(function)).To(gomega.MatchError("size should be one of 'S,M,L,XL'"))

	// wrong functionContentType
	function = &runtimev1alpha1.Function{
		Spec: runtimev1alpha1.FunctionSpec{
			FunctionContentType: "UnknownFunctionContentType",
			Function:            "foo",
			Size:                "S",
			Runtime:             "nodejs8",
		},
	}
	g.Expect(functionCreateHandler.validateFunctionFn(function)).To(gomega.MatchError("functionContentType should be one of 'plaintext,base64'"))

}

// Check that a function with invalid parameter values get's rejected by the webhook
// other value permutations are already covered by unit test TestValidation
func TestHandleInvalid(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	admissionDecoder, err := admission.NewDecoder(scheme.Scheme)
	if err != nil {
		t.Error("Could not create admission decoder")
	}

	functionCreateHandler := FunctionCreateHandler{
		Decoder: admissionDecoder,
	}

	// create an admission request
	req := &admissionv1beta1.AdmissionRequest{
		Operation: admissionv1beta1.Create,
		Kind: metav1.GroupVersionKind{
			Group:   "kyma-project.io",
			Version: "v1alpha1",
			Kind:    "Function",
		},
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
}
`),
		},
	}

	// call the handler - this is the method called by the webhook server
	response := functionCreateHandler.Handle(context.TODO(), types.Request{AdmissionRequest: req})

	// ensure function got rejected
	g.Expect(response.Response.Allowed).To(gomega.BeFalse())

	// ensure that object has not been declined due to decoding issues (4XX error code)
	// reject functions should return a 500 error code
	g.Expect(response.Response.Result.Code).To(gomega.Equal(int32(500)))

}

// Check that the default values are added for optional parameters
func TestHandleDefaults(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	// these are the expected patches for an empty function
	expectedPatches := []jsonpatch.Operation{
		{
			Operation: "replace",
			Path:      "/spec/functionContentType",
			Value:     "plaintext",
		},
		{
			Operation: "replace",
			Path:      "/spec/size",
			Value:     "S",
		},
		{
			Operation: "replace",
			Path:      "/spec/runtime",
			Value:     "nodejs8",
		},
		{
			Operation: "add",
			Path:      "/spec/timeout",
			Value:     float64(180),
		},
	}

	admissionDecoder, err := admission.NewDecoder(scheme.Scheme)
	if err != nil {
		t.Error("could not created admission decoder")
	}

	functionCreateHandler := FunctionCreateHandler{
		Decoder: admissionDecoder,
	}

	// create admission request
	req := &admissionv1beta1.AdmissionRequest{
		Operation: admissionv1beta1.Create,
		Kind: metav1.GroupVersionKind{
			Group:   "kyma-project.io",
			Version: "v1alpha1",
			Kind:    "Function",
		},
	}

	// call handler
	response := functionCreateHandler.Handle(context.TODO(), types.Request{AdmissionRequest: req})

	// ensure request got accepted and check patches later
	g.Expect(response.Response.Allowed).To(gomega.BeTrue())

	// check that each received patch matches at least one expected patch
	for _, actualPatch := range response.Patches {
		g.Expect(expectedPatches).To(gomega.ContainElement(gomega.BeEquivalentTo(actualPatch)))
	}

}
