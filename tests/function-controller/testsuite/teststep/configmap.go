package teststep

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/configmap"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/step"
)

type ConfigMaps struct {
	name      string
	configMap *configmap.ConfigMap
	data      map[string]string
	log       *logrus.Entry
}

func CreateConfigMap(log *logrus.Entry, cm *configmap.ConfigMap, stepName string, data map[string]string) step.Step {
	return &ConfigMaps{
		name:      stepName,
		data:      data,
		log:       log,
		configMap: cm,
	}
}

func (c ConfigMaps) Name() string {
	return c.name
}

func (c ConfigMaps) Run() error {
	return errors.Wrap(c.configMap.Create(c.data), "while checking if configmap is ready")
}

func (c ConfigMaps) Cleanup() error {
	return errors.Wrap(c.configMap.Delete(), "while deleting configmap")
}

var _ step.Step = ConfigMaps{}
