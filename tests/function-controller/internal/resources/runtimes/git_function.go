package runtimes

import (
	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
)

func GitopsFunction(repoURL, baseDir, reference string, rtm serverlessv1alpha2.Runtime, auth *serverlessv1alpha2.RepositoryAuth) serverlessv1alpha2.FunctionSpec {
	if baseDir == "" {
		baseDir = "/"
	}

	if reference == "" {
		reference = "main"
	}

	var minReplicas int32 = 1
	var maxReplicas int32 = 2

	gitRepo := &serverlessv1alpha2.GitRepositorySource{
		URL: repoURL,
		Repository: serverlessv1alpha2.Repository{
			BaseDir:   baseDir,
			Reference: reference,
		},
	}
	if auth != nil {
		gitRepo.Auth = auth
	}
	return serverlessv1alpha2.FunctionSpec{
		Runtime: rtm,
		Source: serverlessv1alpha2.Source{
			GitRepository: gitRepo,
		},
		ScaleConfig: &serverlessv1alpha2.ScaleConfig{
			MinReplicas: &minReplicas,
			MaxReplicas: &maxReplicas,
		},
		ResourceConfiguration: &serverlessv1alpha2.ResourceConfiguration{
			Function: &serverlessv1alpha2.ResourceRequirements{
				Profile: "M",
			},
			Build: &serverlessv1alpha2.ResourceRequirements{
				Profile: "normal",
			},
		},
	}
}
