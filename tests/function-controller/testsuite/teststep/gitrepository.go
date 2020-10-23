package teststep

import (
	serverlessv1alhpa1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/gitrepository"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/step"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type createGitRepository struct {
	name string
	spec serverlessv1alhpa1.GitRepositorySpec
	repo *gitrepository.GitRepository
	log  *logrus.Entry
}

var _ step.Step = createGitRepository{}

func NewCreateGitRepository(log *logrus.Entry, repo *gitrepository.GitRepository, stepName string, spec serverlessv1alhpa1.GitRepositorySpec) step.Step {
	return createGitRepository{
		name: stepName,
		spec: spec,
		repo: repo,
		log:  log.WithField(step.LogStepKey, stepName),
	}
}

func (r createGitRepository) Name() string {
	return r.name
}

func (r createGitRepository) Run() error {
	return errors.Wrapf(r.repo.Create(r.spec), "while creating GitRepository: %s", r.name)
}

func (r createGitRepository) Cleanup() error {
	return errors.Wrapf(r.repo.Delete(), "while deleting GitRepository: %s", r.name)
}

func (r createGitRepository) OnError() error {
	return r.repo.LogResource()
}
