package git

import (
	"github.com/kyma-project/kyma/tests/function-controller/internal/executor"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	filePath = "handler.js"
	oldValue = "GITOPS 1"
	newValue = "GITOPS 2"
)

type commitChanges struct {
	name      string
	gitClient Client
	log       *logrus.Entry
}

var _ executor.Step = commitChanges{}

func NewCommitChanges(log *logrus.Entry, stepName, repoURL string) executor.Step {
	return commitChanges{
		name:      stepName,
		gitClient: NewGitClient(repoURL),
		log:       log.WithField(executor.LogStepKey, stepName),
	}
}

func (c commitChanges) Name() string {
	return c.name
}

func (c commitChanges) Run() error {
	err := c.gitClient.ReplaceInRemoteFile(filePath, oldValue, newValue)
	return errors.Wrap(err, "while replacing file content in git repository")
}

func (c commitChanges) OnError() error {
	out, err := c.gitClient.PullRemote(filePath)
	if err != nil {
		return errors.Wrap(err, "while pulling from remote repository")
	}
	c.log.Infof("Code from git repository: %s", out)
	return nil
}

func (c commitChanges) Cleanup() error {
	return nil
}
