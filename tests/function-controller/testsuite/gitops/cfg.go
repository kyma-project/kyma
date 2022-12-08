package gitops

import (
	"fmt"
	"net/url"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"
)

type GitopsConfig struct {
	FnName               string
	RepoName             string
	GitServerImage       string
	GitServerServiceName string
	GitServerServicePort int
	GitServerRepoName    string
	Toolbox              shared.Container
}

const (
	gitServerServiceName    = "gitserver"
	gitServerServicePort    = 80
	gitServerEndpointFormat = "http://%s.%s.svc.cluster.local:%v/%s.git"
)

func NewGitopsConfig(fnName, gitServerImage, gitServerRepoName string, toolbox shared.Container) (GitopsConfig, error) {
	return GitopsConfig{
		FnName:               fnName,
		GitServerImage:       gitServerImage,
		GitServerServicePort: gitServerServicePort,
		GitServerServiceName: gitServerServiceName,
		GitServerRepoName:    gitServerRepoName,
		Toolbox:              toolbox,
	}, nil
}

func (c *GitopsConfig) GetGitServerURL(useProxy bool) string {
	var functionURL = ""
	if useProxy {
		functionURL = fmt.Sprintf("http://127.0.0.1:8001/api/v1/namespaces/%s/services/%s:80/proxy/%s.git", c.Toolbox.Namespace, c.GitServerServiceName, c.GitServerRepoName)
	} else {
		functionURL = c.GetGitServerInClusterURL()
	}
	parsedURL, err := url.Parse(functionURL)
	if err != nil {
		panic(err)
	}

	return parsedURL.String()
}

func (c *GitopsConfig) GetGitServerInClusterURL() string {
	functionURL := fmt.Sprintf(gitServerEndpointFormat, c.GitServerServiceName, c.Toolbox.Namespace, c.GitServerServicePort, c.GitServerRepoName)
	parsedURL, err := url.Parse(functionURL)
	if err != nil {
		panic(err)
	}

	return parsedURL.String()
}
