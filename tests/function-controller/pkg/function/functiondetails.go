package function

import (
	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	corev1 "k8s.io/api/core/v1"
)

type FunctionData struct {
	Body        string
	Deps        string
	MinReplicas int32
	MaxReplicas int32
	Runtime     serverlessv1alpha2.Runtime
	SourceType  serverlessv1alpha2.FunctionType
	Repository  serverlessv1alpha2.Repository
	Env         []corev1.EnvVar
}
