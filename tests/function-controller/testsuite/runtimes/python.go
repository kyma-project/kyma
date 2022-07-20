package runtimes

import (
	"fmt"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
)

func BasicPythonFunction(msg string, runtime serverlessv1alpha1.Runtime) serverlessv1alpha1.FunctionSpec {
	src := fmt.Sprintf(`import arrow 
def main(event, context):
	return "%s"`, msg)

	dpd := `requests==2.24.0
arrow==0.15.8`

	return serverlessv1alpha1.FunctionSpec{
		Runtime: runtime,
		Source: serverlessv1alpha1.Source{
			Inline: &serverlessv1alpha1.InlineSource{
				Source:       src,
				Dependencies: dpd,
			},
		},
	}
}

func BasicPythonFunctionWithCustomDependency(msg string, runtime serverlessv1alpha1.Runtime) serverlessv1alpha1.FunctionSpec {
	src := fmt.Sprintf(
		`import arrow
def main(event, context):
	return "%s"`, msg)

	dpd := `requests==2.24.0
arrow==0.15.8
kyma-pypi-test==1.0.0`

	return serverlessv1alpha1.FunctionSpec{
		Runtime: runtime,
		Source: serverlessv1alpha1.Source{
			Inline: &serverlessv1alpha1.InlineSource{
				Source:       src,
				Dependencies: dpd,
			},
		},
	}
}
