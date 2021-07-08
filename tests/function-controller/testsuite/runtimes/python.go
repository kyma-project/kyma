package runtimes

import (
	"fmt"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/function"
)

func BasicPythonFunction(msg string, runtime serverlessv1alpha1.Runtime) *function.FunctionData {
	return &function.FunctionData{
		Body: fmt.Sprintf(
			`import arrow
def main(event, context):
	return "%s"`, msg),
		Deps: `requests==2.24.0
arrow==0.15.8`,
		MinReplicas: 1,
		MaxReplicas: 1,
		Runtime:     runtime,
	}
}

func BasicPythonFunctionWithCustomDependency(msg string, runtime serverlessv1alpha1.Runtime) *function.FunctionData {
	return &function.FunctionData{
		Body: fmt.Sprintf(
			`import arrow
def main(event, context):
	return "%s"`, msg),
		Deps: `requests==2.24.0
arrow==0.15.8
kyma-pypi-test==1.0.0`,
		MinReplicas: 1,
		MaxReplicas: 1,
		Runtime:     runtime,
	}
}
