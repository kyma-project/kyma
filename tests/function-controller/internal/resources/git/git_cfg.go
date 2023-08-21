package git

import (
	"github.com/kyma-project/kyma/tests/function-controller/internal/utils"
)

type GitopsConfig struct {
	FnName               string
	RepoName             string
	GitServerImage       string
	GitServerServiceName string
	GitServerServicePort int32
	GitServerRepoName    string
	Toolbox              utils.Container
}

const (
	gitServerServiceName       = "gitserver"
	gitServerServicePort int32 = 80
)

func NewGitopsConfig(fnName, gitServerImage, gitServerRepoName string, toolbox utils.Container) (GitopsConfig, error) {
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
	gitURL, err := utils.GetGitURL(c.GitServerServiceName, c.Toolbox.Namespace, c.GitServerRepoName, useProxy)
	if err != nil {
		panic(err)
	}
	return gitURL.String()
}

func (c *GitopsConfig) GetGitServerInClusterURL() string {
	return c.GetGitServerURL(false)

}
