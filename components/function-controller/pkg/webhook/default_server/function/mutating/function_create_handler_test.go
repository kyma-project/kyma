package mutating

import (
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
				"DockerFileName": "dockerfile-nodejs8",
			},
			{
				"ID": "nodejs6",
				"DockerFileName": "dockerfile-nodejs6",
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
	functionCreateHandler.mutatingFunctionFn(function, rnInfo)

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
	g := gomega.NewWithT(t)
	rnInfo := runtimeConfig(t)

	// wrong serverless
	function := &serverlessv1alpha1.Function{
		Spec: serverlessv1alpha1.FunctionSpec{
			FunctionContentType: "plaintext",
			Function:            "foo",
			Size:                "S",
			Runtime:             "nodejs4",
		},
	}
	g.Expect(functionCreateHandler.validateFunctionFn(function, rnInfo)).To(gomega.MatchError("Passed Runtime: 'nodejs4', but runtime should be one of 'nodejs8, nodejs6'"))

	// wrong size
	function = &serverlessv1alpha1.Function{
		Spec: serverlessv1alpha1.FunctionSpec{
			FunctionContentType: "plaintext",
			Function:            "foo",
			Size:                "UnknownSize",
			Runtime:             "nodejs8",
		},
	}
	g.Expect(functionCreateHandler.validateFunctionFn(function, rnInfo)).To(gomega.MatchError("Passed function size: 'UnknownSize', but size should be one of 'S, M, L'"))

	// wrong functionContentType
	function = &serverlessv1alpha1.Function{
		Spec: serverlessv1alpha1.FunctionSpec{
			FunctionContentType: "UnknownFunctionContentType",
			Function:            "foo",
			Size:                "S",
			Runtime:             "nodejs8",
		},
	}
	g.Expect(functionCreateHandler.validateFunctionFn(function, rnInfo)).To(gomega.MatchError("Passed functionContetype: 'UnknownFunctionContentType', but functionContentType should be one of 'plaintetext, base64'"))

}
