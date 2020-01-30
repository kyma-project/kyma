package webhook

import (
	"strings"
	"testing"

	"github.com/onsi/gomega/gstruct"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/onsi/gomega"

	runtimeUtil "github.com/kyma-project/kyma/components/function-controller/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var functionCreateHandler = FunctionCreateHandler{}

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
	functionCreateHandler.mutatingFunctionFn(function, rnInfo)

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
	testCases := []struct {
		name    string
		tweakFn func(fn *serverlessv1alpha1.Function)
		numErrs int
	}{
		{
			name:    "valid function",
			tweakFn: func(fn *serverlessv1alpha1.Function) {},
			numErrs: 0,
		},
		{
			name: "missing namespace",
			tweakFn: func(fn *serverlessv1alpha1.Function) {
				fn.Namespace = ""
			},
			numErrs: 1,
		},
		{
			name: "invalid namespace",
			tweakFn: func(fn *serverlessv1alpha1.Function) {
				fn.Namespace = "-123"
			},
			numErrs: 1,
		},
		{
			name: "missing name",
			tweakFn: func(fn *serverlessv1alpha1.Function) {
				fn.Name = ""
			},
			numErrs: 1,
		},
		{
			name: "invalid name",
			tweakFn: func(fn *serverlessv1alpha1.Function) {
				fn.Name = "-123"
			},
			numErrs: 1,
		},
		{
			name: "too long name",
			tweakFn: func(fn *serverlessv1alpha1.Function) {
				fn.Name = strings.Repeat("a", 64)
			},
			numErrs: 1,
		},
		{
			name: "invalid generateName",
			tweakFn: func(fn *serverlessv1alpha1.Function) {
				fn.GenerateName = "-123"
			},
			numErrs: 1,
		},
		{
			name: "too long generateName",
			tweakFn: func(fn *serverlessv1alpha1.Function) {
				fn.GenerateName = strings.Repeat("a", 64)
			},
			numErrs: 1,
		},
		{
			name: "invalid runtime",
			tweakFn: func(fn *serverlessv1alpha1.Function) {
				fn.Spec.Runtime = "nodejs4"
			},
			numErrs: 1,
		},
		{
			name: "invalid function size",
			tweakFn: func(fn *serverlessv1alpha1.Function) {
				fn.Spec.Size = "UnknownSize"
			},
			numErrs: 1,
		},
		{
			name: "invalid functionContentType",
			tweakFn: func(fn *serverlessv1alpha1.Function) {
				fn.Spec.FunctionContentType = "UnknownFunctionContentType"
			},
			numErrs: 1,
		},
		{
			name: "multiple errors",
			tweakFn: func(fn *serverlessv1alpha1.Function) {
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
		tc.tweakFn(fn)
		errs := functionCreateHandler.validateFunctionFn(fn, rnInfo)

		if len(errs) != tc.numErrs {
			g.Expect(errs).To(gomega.HaveLen(tc.numErrs))
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
