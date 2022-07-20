package runtimes

import (
	"fmt"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

var (
	minReplicas int32 = 1
	maxReplicas int32 = 2
)

func BasicPythonFunction(msg string, runtime serverlessv1alpha1.Runtime) serverlessv1alpha1.FunctionSpec {
	src := fmt.Sprintf(`import arrow 
def main(event, context):
	return "%s"`, msg)

	dpd := `requests==2.24.0
arrow==0.15.8`

	return serverlessv1alpha1.FunctionSpec{
		Source:      src,
		Deps:        dpd,
		Runtime:     runtime,
		MinReplicas: &minReplicas,
		MaxReplicas: &maxReplicas,
	}
}
