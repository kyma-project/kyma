package configmap

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/resource"
)

type ConfigMap struct {
	resCli      *resource.Resource
	name        string
	namespace   string
	waitTimeout time.Duration
	log         *logrus.Entry
	verbose     bool
}

func NewConfigMap(name string, c shared.Container) *ConfigMap {
	return &ConfigMap{
		resCli:      resource.New(c.DynamicCli, corev1.SchemeGroupVersion.WithResource("configmaps"), c.Namespace, c.Log, c.Verbose),
		name:        name,
		namespace:   c.Namespace,
		waitTimeout: c.WaitTimeout,
		log:         c.Log,
		verbose:     c.Verbose,
	}
}

func (c *ConfigMap) Name() string {
	return c.name
}

func (c *ConfigMap) Create(data map[string]string) error {
	function := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.name,
			Namespace: c.namespace,
		},
		Data: data,
	}

	_, err := c.resCli.Create(function)
	if err != nil {
		return errors.Wrapf(err, "while creating ConfigMap %s in namespace %s", c.name, c.namespace)
	}
	return err
}

func (c *ConfigMap) Delete() error {
	err := c.resCli.Delete(c.name, c.waitTimeout)
	if err != nil {
		return errors.Wrapf(err, "while deleting ConfigMap %s in namespace %s", c.name, c.namespace)
	}

	return nil
}
