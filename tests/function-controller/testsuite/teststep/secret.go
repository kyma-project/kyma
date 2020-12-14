package teststep

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/secret"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/step"
)

type Secrets struct {
	name   string
	secret *secret.Secret
	data   map[string]string
	log    *logrus.Entry
}

func CreateSecret(log *logrus.Entry, sec *secret.Secret, stepName string, data map[string]string) step.Step {
	return &Secrets{
		name:   stepName,
		data:   data,
		log:    log.WithField(step.LogStepKey, stepName),
		secret: sec,
	}
}

func (s Secrets) Name() string {
	return s.name
}

func (s Secrets) Run() error {
	return errors.Wrap(s.secret.Create(s.data), "while creating secret")
}

func (s Secrets) Cleanup() error {
	return errors.Wrap(s.secret.Delete(), "while deleting secret")
}

func (s Secrets) OnError() error {
	return s.secret.LogResource()
}

var _ step.Step = Secrets{}
