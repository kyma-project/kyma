package function

import (
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

type FunctionData struct {
	Body        string
	Deps        string
	MinReplicas int32
	MaxReplicas int32
	Runtime     serverlessv1alpha1.Runtime
	SourceType  serverlessv1alpha1.SourceType
	Repository  serverlessv1alpha1.Repository
	Env         []corev1.EnvVar
}
