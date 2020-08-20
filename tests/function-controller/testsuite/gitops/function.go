package gitops

import (
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/function"
)

func GitopsFunction(repoName string, rtm serverlessv1alpha1.Runtime) *function.FunctionData {
	return &function.FunctionData{
		SourceType: serverlessv1alpha1.SourceTypeGit,
		Body:       repoName,
		Repository: serverlessv1alpha1.Repository{
			BaseDir:   "/",
			Reference: "master",
		},
		MinReplicas: 1,
		MaxReplicas: 2,
		Runtime:     rtm,
	}
}
