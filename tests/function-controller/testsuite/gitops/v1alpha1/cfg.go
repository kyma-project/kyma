package gitops

import (
	"fmt"
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
	InClusterURL         *url.URL
	Toolbox              shared.Container
}

const (
	gitServerServiceName    = "gitserver"
	gitServerServicePort    = 80
	gitServerEndpointFormat = "http://%s.%s.svc.cluster.local:%v/%s.git"
)

func NewGitopsConfig(fnName, gitServerImage, gitServerRepoName string, toolbox shared.Container) (GitopsConfig, error) {
	inClusterURL, err := url.Parse(fmt.Sprintf("http://%s.%s.svc.cluster.local", fnName, toolbox.Namespace))
	if err != nil {
		return GitopsConfig{}, err
	}

	return GitopsConfig{
		FnName:               fnName,
		Fn:                   functionv1alpha1.NewFunction(fnName, toolbox),
		GitServerImage:       gitServerImage,
		GitServerServicePort: gitServerServicePort,
		GitServerServiceName: gitServerServiceName,
		GitServerRepoName:    gitServerRepoName,
		InClusterURL:         inClusterURL,
		Toolbox:              toolbox,
	}, nil
}

func (c *GitopsConfig) GetGitServerInClusterURL() string {
	return fmt.Sprintf(gitServerEndpointFormat, c.GitServerServiceName, c.Toolbox.Namespace, c.GitServerServicePort, c.GitServerRepoName)
}
