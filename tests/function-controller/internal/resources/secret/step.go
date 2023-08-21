package secret

import (
	"github.com/kyma-project/kyma/tests/function-controller/internal/executor"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Secrets struct {
	name   string
	secret *Secret
	data   map[string]string
	log    *logrus.Entry
}

func CreateSecret(log *logrus.Entry, sec *Secret, stepName string, data map[string]string) executor.Step {
	return &Secrets{
		name:   stepName,
		data:   data,
		log:    log.WithField(executor.LogStepKey, stepName),
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
	return nil
}

var _ executor.Step = Secrets{}
