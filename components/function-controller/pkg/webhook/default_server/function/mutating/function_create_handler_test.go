package mutating

import (
	"fmt"
	"strings"
	"testing"

	"k8s.io/client-go/rest"

	"github.com/onsi/gomega/gstruct"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/onsi/gomega"

	runtimeUtil "github.com/kyma-project/kyma/components/function-controller/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var functionCreateHandler = FunctionCreateHandler{}
var cfg *rest.Config

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
			"timeOut": 10,
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
			{"type": "plaintetext"},
			{"type": "base64"}
		]`,
	},
}

func runtimeConfig(t *testing.T) *runtimeUtil.RuntimeInfo {
	g := gomega.NewGomegaWithT(t)

	rnInfo, err := runtimeUtil.New(fnConfig)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	return rnInfo
}

// Test that an empty function gets all default values set
func TestMutation(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	rnInfo := runtimeConfig(t)

	function := &serverlessv1alpha1.Function{
		ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "default"},
		Spec: serverlessv1alpha1.FunctionSpec{
			FunctionContentType: "plaintext",
			Function:            "foo",
		},
	}

	// mutate function
	functionCreateHandler.mutatingFunction(function, rnInfo)

	// ensure defaults are set
	g.Expect(function.Spec).To(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
		"Size":                gomega.BeEquivalentTo("S"),
		"Timeout":             gomega.BeEquivalentTo(10),
		"Runtime":             gomega.BeEquivalentTo("nodejs8"),
		"FunctionContentType": gomega.BeEquivalentTo("plaintext"),
	}))
}

// Test that all values get validated
func TestValidation(t *testing.T) {
	testCases := []struct {
		name       string
		tweakSvcFn func(fn *serverlessv1alpha1.Function)
		numErrs    int
	}{
		{
			name:     "valid function",
			tweakSvc: func(fn *serverlessv1alpha1.Function) {},
			numErrs:  0,
		},
		{
			name: "missing namespace",
			tweakSvc: func(fn *serverlessv1alpha1.Function) {
				fn.Namespace = ""
			},
			numErrs: 1,
		},
		{
			name: "invalid namespace",
			tweakSvc: func(fn *serverlessv1alpha1.Function) {
				fn.Namespace = "-123"
			},
			numErrs: 1,
		},
		{
			name: "missing name",
			tweakSvc: func(fn *serverlessv1alpha1.Function) {
				fn.Name = ""
			},
			numErrs: 1,
		},
		{
			name: "invalid name",
			tweakSvc: func(fn *serverlessv1alpha1.Function) {
				fn.Name = "-123"
			},
			numErrs: 1,
		},
		{
			name: "too long name",
			tweakSvc: func(fn *serverlessv1alpha1.Function) {
				fn.Name = strings.Repeat("a", 64)
			},
			numErrs: 1,
		},
		{
			name: "invalid generateName",
			tweakSvc: func(fn *serverlessv1alpha1.Function) {
				fn.GenerateName = "-123"
			},
			numErrs: 1,
		},
		{
			name: "too long generateName",
			tweakSvc: func(fn *serverlessv1alpha1.Function) {
				fn.GenerateName = strings.Repeat("a", 64)
			},
			numErrs: 1,
		},
		{
			name: "invalid runtime",
			tweakSvc: func(fn *serverlessv1alpha1.Function) {
				fn.Spec.Runtime = "nodejs4"
			},
			numErrs: 1,
		},
		{
			name: "invalid function size",
			tweakSvc: func(fn *serverlessv1alpha1.Function) {
				fn.Spec.Size = "UnknownSize"
			},
			numErrs: 1,
		},
		{
			name: "invalid functionContentType",
			tweakSvc: func(fn *serverlessv1alpha1.Function) {
				fn.Spec.FunctionContentType = "UnknownFunctionContentType"
			},
			numErrs: 1,
		},
		{
			name: "multiple errors",
			tweakSvc: func(fn *serverlessv1alpha1.Function) {
				fn.Name = "123"
				fn.Spec.Runtime = "nodejs4"
				fn.Spec.Size = "UnknownSize"
				fn.Spec.FunctionContentType = "UnknownFunctionContentType"
			},
			numErrs: 4,
		},
	}

	g := gomega.NewWithT(t)
	rnInfo := runtimeConfig(t)

	for _, tc := range testCases {
		fn := fixValidFunction()
		tc.tweakSvc(fn)
		errs := functionCreateHandler.validateFunction(fn, rnInfo)

		if len(errs) != tc.numErrs {
			g.Expect(errs).To(gomega.HaveLen(tc.numErrs), fmt.Printf("Unexpected error list for case %q: %v", tc.name, errs.ToAggregate()))
		}
	}
}

func fixValidFunction() *serverlessv1alpha1.Function {
	return &serverlessv1alpha1.Function{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-123",
			Namespace: "ns",
		},
		Spec: serverlessv1alpha1.FunctionSpec{
			FunctionContentType: "plaintext",
			Function:            "foo",
			Size:                "S",
			Runtime:             "nodejs6",
		},
	}
}
