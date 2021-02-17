package configmap

import (
	"time"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/sirupsen/logrus"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/helpers"
	"github.com/kyma-project/kyma/tests/function-controller/pkg/shared"

	"github.com/kyma-project/kyma/tests/function-controller/pkg/resource"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.name,
			Namespace: c.namespace,
		},
		Data: data,
	}

	_, err := c.resCli.Create(cm)
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

func (c *ConfigMap) Get() (*corev1.ConfigMap, error) {
	u, err := c.resCli.Get(c.name)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting ConfigMap %s in namespace %s", c.name, c.namespace)
	}
	cm := corev1.ConfigMap{}
	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &cm); err != nil {
		return nil, errors.Wrap(err, "while constructing ConfigMap from unstructured")
	}

	return &cm, nil
}
func (c *ConfigMap) LogResource() error {
	cm, err := c.Get()
	if err != nil {
		return errors.Wrap(err, "while getting ConfigMap")
	}

	out, err := helpers.PrettyMarshall(cm)
	if err != nil {
		return err
	}

	c.log.Infof("ConfigMap: %s", out)
	return nil
}
