package gitops

import (
	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
)

func GitopsFunction(repoURL, baseDir, reference string, rtm serverlessv1alpha2.Runtime) serverlessv1alpha2.FunctionSpec {
	if baseDir == "" {
		baseDir = "/"
	}

	if reference == "" {
		reference = "main"
	}

	var minReplicas int32 = 1
	var maxReplicas int32 = 2

	return serverlessv1alpha2.FunctionSpec{
		Runtime: rtm,
		Source: serverlessv1alpha2.Source{
			GitRepository: &serverlessv1alpha2.GitRepositorySource{
				URL: repoURL,
				Repository: serverlessv1alpha2.Repository{
					BaseDir:   baseDir,
					Reference: reference,
				},
			},
		},
		MinReplicas: &minReplicas,
		MaxReplicas: &maxReplicas,
	}
}
