package gitops

import (
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/function"
)

func GitopsFunction(repoName, baseDir string, rtm serverlessv1alpha1.Runtime) *function.FunctionData {
	if baseDir == "" {
		baseDir = "/"
	}
	return &function.FunctionData{
		SourceType: serverlessv1alpha1.SourceTypeGit,
		Body:       repoName,
		Repository: serverlessv1alpha1.Repository{
			BaseDir:   baseDir,
			Reference: "main",
		},
		MinReplicas: 1,
		MaxReplicas: 2,
		Runtime:     rtm,
	}
}
