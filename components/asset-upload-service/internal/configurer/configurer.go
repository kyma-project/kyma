package configurer

import (
	"github.com/kyma-project/kyma/components/asset-upload-service/internal/bucket"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
)

type Config struct {
	Name string `envconfig:"asset-upload-service"`
	Namespace string `envconfig:"default=kyma-system"`
}

type SharedAppConfig struct {
	SystemBuckets bucket.SystemBucketNames
}

type Configurer struct {
	client corev1.CoreV1Interface
	cfg    Config
}

//TODO: tests

func New(restConfig *restclient.Config, cfg Config) (*Configurer, error) {
	client, err := typedcorev1.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating K8S Client")
	}

	return &Configurer{
		client: client,
		cfg:    cfg,
	}, nil
}

func (c *Configurer) LoadIfExists() (*SharedAppConfig, bool, error) {
	configMap, err := c.client.ConfigMaps(c.cfg.Namespace).Get(c.cfg.Name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, false, nil
		}
		return nil, false, errors.Wrapf(err, "while loading ConfigMap %s from %s", c.cfg.Name, c.cfg.Namespace)
	}

	sharedAppConfig := c.fromConfigMap(configMap)

	return sharedAppConfig, true, nil
}

func (c *Configurer) Save(config SharedAppConfig) error {
	_, err := c.client.ConfigMaps(c.cfg.Namespace).Create(c.convertToConfigMap(config))
	if err != nil {
		return errors.Wrapf(err, "while creating ConfigMap %s in namespace %s", c.cfg.Name, c.cfg.Namespace)
	}

	return nil
}

func (c *Configurer) convertToConfigMap(cfg SharedAppConfig) *v1.ConfigMap {
	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: c.cfg.Name,
			Namespace: c.cfg.Namespace,
		},
		Data: map[string]string{
			"private": cfg.SystemBuckets.Private,
			"public": cfg.SystemBuckets.Public,
		},
	}

	return configMap
}

func (c *Configurer) fromConfigMap(configMap *v1.ConfigMap) *SharedAppConfig {
	appConfig := &SharedAppConfig{
		SystemBuckets: bucket.SystemBucketNames{
			Public:  configMap.Data["private"],
			Private: configMap.Data["public"],
		},
	}

	return appConfig
}
