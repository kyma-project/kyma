package teststep

import (
	"github.com/kyma-project/kyma/tests/function-controller/pkg/git"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/step"
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
	gitClient git.Client
	log       *logrus.Entry
}

var _ step.Step = commitChanges{}

func NewCommitChanges(log *logrus.Entry, stepName, repoURL string) step.Step {
	return commitChanges{
		name:      stepName,
		gitClient: git.New(repoURL),
		log:       log.WithField(step.LogStepKey, stepName),
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
