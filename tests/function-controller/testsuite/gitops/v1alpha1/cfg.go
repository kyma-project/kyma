package gitops

import (
	"github.com/kyma-project/kyma/tests/function-controller/pkg/helpers"
	"net/url"

	functionv1alpha1 "github.com/kyma-project/kyma/tests/function-controller/pkg/function/v1alpha1"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"
)

type GitopsConfig struct {
	FnName               string
	Fn                   *functionv1alpha1.Function
	RepoName             string
	GitServerImage       string
	GitServerServiceName string
	GitServerServicePort int
	GitServerRepoName    string
	GitInClusterURL      *url.URL
	InClusterURL         *url.URL
	Toolbox              shared.Container
}

const (
	gitServerServiceName = "gitserver"
	gitServerServicePort = 80
)

func NewGitopsConfig(fnName, gitServerImage, gitServerRepoName string, useProxy bool, toolbox shared.Container) (GitopsConfig, error) {
	fnURL, err := helpers.GetSvcURL(fnName, toolbox.Namespace, useProxy)
	if err != nil {
		panic(err)
	}

	gitRepoURL, err := helpers.GetGitURL(gitServerServiceName, toolbox.Namespace, gitServerRepoName, false)
	if err != nil {
		panic(err)
	}

	return GitopsConfig{
		FnName:               fnName,
		Fn:                   functionv1alpha1.NewFunction(fnName, useProxy, toolbox),
		GitServerImage:       gitServerImage,
		GitServerServicePort: gitServerServicePort,
		GitServerServiceName: gitServerServiceName,
		GitServerRepoName:    gitServerRepoName,
		InClusterURL:         fnURL,
		GitInClusterURL:      gitRepoURL,
		Toolbox:              toolbox,
	}, nil
}
