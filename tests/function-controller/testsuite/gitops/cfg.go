package gitops

import (
	"github.com/kyma-project/kyma/tests/function-controller/pkg/helpers"
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
	gitServerServiceName = "gitserver"
	gitServerServicePort = 80
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
	gitURL, err := helpers.GetGitURL(c.GitServerServiceName, c.Toolbox.Namespace, c.GitServerRepoName, useProxy)
	if err != nil {
		panic(err)
	}
	return gitURL.String()
}

func (c *GitopsConfig) GetGitServerInClusterURL() string {
	return c.GetGitServerURL(false)

}
