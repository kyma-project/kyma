package configurer

import (
	"github.com/pkg/errors"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"

	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kyma-project/kyma/tests/console-backend-service/internal/waiter"
	corev1 "k8s.io/api/core/v1"
	corev1Type "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	brokerReadyTimeout = time.Second * 300
)


type TestBundleConfig struct {
	ConfigMap TestBundleConfigMap

	ClusterServiceClassExternalName string `envconfig:"default=testing"`
	ClusterServicePlanExternalNames []string `envconfig:"default=full,minimal"`
}

type TestBundleConfigMap struct {
	Name      string  `envconfig:"default=helm-broker"`
	Namespace string `envconfig:"default=kyma-system"`
	RepositoryURL string `envconfig:"default=https://github.com/kyma-project/bundles/releases/download/0.3.0/index-testing.yaml"`
	LabelKey string `envconfig:"default=helm-broker-repo"`
	LabelValue string `envconfig:"default=true"`
}

type TestBundleConfigurer struct {
	cfg TestBundleConfig

	coreCli  *corev1Type.CoreV1Client
	svcatCli *clientset.Clientset
}

func NewTestBundle(cfg TestBundleConfig, coreCli *corev1Type.CoreV1Client, svcatCli *clientset.Clientset) *TestBundleConfigurer {
	return &TestBundleConfigurer{
		cfg: cfg,
		coreCli:  coreCli,
		svcatCli: svcatCli,
	}
}

func (c *TestBundleConfigurer) Configure() error {
	cfg := c.cfg
	_, err := c.coreCli.ConfigMaps(cfg.ConfigMap.Namespace).Create(
		&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cfg.ConfigMap.Name,
				Namespace: cfg.ConfigMap.Namespace,
				Labels: map[string]string{
					cfg.ConfigMap.LabelKey: cfg.ConfigMap.LabelValue,
				},
			},
			Data: map[string]string{
				"URLs": cfg.ConfigMap.RepositoryURL,
			},
		})

	if err != nil && !apiErrors.IsAlreadyExists(err) {
		return err
	}

	return nil
}

func (c *TestBundleConfigurer) Cleanup() error {
	cfgMap := c.cfg.ConfigMap
	err := c.coreCli.ConfigMaps(cfgMap.Namespace).Delete(cfgMap.Name, &metav1.DeleteOptions{})
	if err != nil && !apiErrors.IsNotFound(err) {
		return err
	}

	return nil
}

func (c *TestBundleConfigurer) WaitForTestBundleReady() error {
	err := c.waitForClusterServiceClass()
	if err != nil {
		return errors.Wrapf(err, "while waiting for ClusterServiceClass with externalName %s", c.cfg.ClusterServiceClassExternalName)
	}

	err = c.waitForClusterServicePlans()
	if err != nil {
		return errors.Wrapf(err, "while waiting for ClusterServicePlans for ClusterServiceClass with externalName %s", c.cfg.ClusterServiceClassExternalName)
	}

	return nil
}

func (c *TestBundleConfigurer) waitForClusterServiceClass() error {
	return waiter.WaitAtMost(func() (bool, error) {
		classesList, err := c.svcatCli.ServicecatalogV1beta1().ClusterServiceClasses().List(metav1.ListOptions{})
		if err != nil {
			return false, err
		}

		for _, class := range classesList.Items {
			if class.GetExternalName() == c.cfg.ClusterServiceClassExternalName {
				return true, nil
			}
		}

		return true, nil
	}, brokerReadyTimeout)
}

func (c *TestBundleConfigurer) waitForClusterServicePlans() error {
	plansFound := map[string]bool{}

	for _, planName := range c.cfg.ClusterServicePlanExternalNames {
		plansFound[planName] = false
	}

	err := waiter.WaitAtMost(func() (bool, error) {
		planList, err := c.svcatCli.ServicecatalogV1beta1().ClusterServicePlans().List(metav1.ListOptions{})
		if err != nil {
			return false, err
		}

		for _, plan := range planList.Items {
			for key := range plansFound {
				if plan.GetExternalName() == key {
					plansFound[key] = true
				}
			}
		}

		for _, value := range plansFound {
			if !value {
				return false, nil
			}
		}

		return true, nil
	}, brokerReadyTimeout)
	if err != nil {
		return errors.Wrapf(err, "while waiting for ClusterServicePlans: %+v", plansFound)
	}

	return nil
}
