package runtimes

import (
	"fmt"

	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
)

func BasicPythonFunction(msg string, runtime serverlessv1alpha2.Runtime) serverlessv1alpha2.FunctionSpec {
	src := fmt.Sprintf(`import arrow 
def main(event, context):
	return "%s"`, msg)

	dpd := `requests==2.24.0
arrow==0.15.8`

	return serverlessv1alpha2.FunctionSpec{
		Runtime:     runtime,
		Source: serverlessv1alpha2.Source{
			Inline: serverlessv1alpha2.InlineSource{
				Source:       src,
				Dependencies: dpd,
			},
		},
	}
}

func BasicPythonFunctionWithCustomDependency(msg string, runtime serverlessv1alpha2.Runtime) serverlessv1alpha2.FunctionSpec {
	src := fmt.Sprintf(
		`import arrow
def main(event, context):
	return "%s"`, msg)

	dpd := `requests==2.24.0
arrow==0.15.8
kyma-pypi-test==1.0.0`

	return serverlessv1alpha2.FunctionSpec{
		Runtime:     runtime,
		Source: serverlessv1alpha2.Source{
			Inline: serverlessv1alpha2.InlineSource{
				Source:       src,
				Dependencies: dpd,
			},
		},
	}
}
