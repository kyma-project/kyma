package configurer

import (
	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/asset-upload-service/internal/bucket"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type Config struct {
	Name      string `envconfig:"default=asset-upload-service"`
	Namespace string `envconfig:"default=kyma-system"`
	Enabled   bool   `envconfig:"default=true"`
}

type SharedAppConfig struct {
	SystemBuckets bucket.SystemBucketNames
}

type Configurer struct {
	client corev1.CoreV1Interface
	cfg    Config
}

func New(client corev1.CoreV1Interface, cfg Config) *Configurer {
	return &Configurer{
		client: client,
		cfg:    cfg,
	}
}

func (c *Configurer) Load() (*SharedAppConfig, error) {
	if !c.cfg.Enabled {
		return nil, nil
	}

	configMap, err := c.client.ConfigMaps(c.cfg.Namespace).Get(c.cfg.Name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "while loading ConfigMap %s from %s", c.cfg.Name, c.cfg.Namespace)
	}

	sharedAppConfig := c.fromConfigMap(configMap)
	glog.Infof("Config successfully loaded from ConfigMap %s in namespace %s", c.cfg.Name, c.cfg.Namespace)

	return sharedAppConfig, nil
}

func (c *Configurer) Save(config SharedAppConfig) error {
	if !c.cfg.Enabled {
		return nil
	}

	_, err := c.client.ConfigMaps(c.cfg.Namespace).Create(c.convertToConfigMap(config))
	if err != nil {
		return errors.Wrapf(err, "while creating ConfigMap %s in namespace %s", c.cfg.Name, c.cfg.Namespace)
	}

	glog.Infof("Config successfully saved to ConfigMap %s in namespace %s", c.cfg.Name, c.cfg.Namespace)

	return nil
}

func (c *Configurer) convertToConfigMap(cfg SharedAppConfig) *v1.ConfigMap {
	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.cfg.Name,
			Namespace: c.cfg.Namespace,
		},
		Data: map[string]string{
			"private": cfg.SystemBuckets.Private,
			"public":  cfg.SystemBuckets.Public,
		},
	}

	return configMap
}

func (c *Configurer) fromConfigMap(configMap *v1.ConfigMap) *SharedAppConfig {
	appConfig := &SharedAppConfig{
		SystemBuckets: bucket.SystemBucketNames{
			Public:  configMap.Data["public"],
			Private: configMap.Data["private"],
		},
	}

	return appConfig
}
