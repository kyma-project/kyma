package configmap

import (
	"github.com/kyma-project/kyma/tests/function-controller/internal/executor"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type ConfigMaps struct {
	name      string
	configMap *ConfigMap
	data      map[string]string
	log       *logrus.Entry
}

func CreateConfigMap(log *logrus.Entry, cm *ConfigMap, stepName string, data map[string]string) executor.Step {
	return &ConfigMaps{
		name:      stepName,
		data:      data,
		log:       log.WithField(executor.LogStepKey, stepName),
		configMap: cm,
	}
}

func (c ConfigMaps) Name() string {
	return c.name
}

func (c ConfigMaps) Run() error {
	return errors.Wrap(c.configMap.Create(c.data), "while creating configmap")
}

func (c ConfigMaps) Cleanup() error {
	return errors.Wrap(c.configMap.Delete(), "while deleting configmap")
}

func (c ConfigMaps) OnError() error {
	return c.configMap.LogResource()
}

var _ executor.Step = ConfigMaps{}
