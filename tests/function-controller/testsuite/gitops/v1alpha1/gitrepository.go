package gitops

import serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"

func NoAuthRepositorySpec(url string) serverlessv1alpha1.GitRepositorySpec {
	return serverlessv1alpha1.GitRepositorySpec{
		URL: url,
	}
}
