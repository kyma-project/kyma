package gitops

import (
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

func GitopsFunction(repoName, baseDir, reference string, rtm serverlessv1alpha1.Runtime) serverlessv1alpha1.FunctionSpec {
	if baseDir == "" {
		baseDir = "/"
	}

	if reference == "" {
		reference = "main"
	}

	return serverlessv1alpha1.FunctionSpec{
		Type:   serverlessv1alpha1.SourceTypeGit,
		Source: repoName,
		Repository: serverlessv1alpha1.Repository{
			BaseDir:   baseDir,
			Reference: reference,
		},
		Runtime: rtm,
	}
}
