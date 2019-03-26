package installer

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

type TestBundleConfigMap struct {
	name      string
	namespace string
	labels    map[string]string
	data      map[string]string
}

type TestBundleInstaller struct {
	configMap                       TestBundleConfigMap
	clusterServiceClassExternalName string
	clusterServicePlanExternalNames []string

	coreCli  *corev1Type.CoreV1Client
	svcatCli *clientset.Clientset
}

func NewTestBundle(name, namespace, urls string, coreCli *corev1Type.CoreV1Client, svcatCli *clientset.Clientset) *TestBundleInstaller {
	return &TestBundleInstaller{
		configMap: TestBundleConfigMap{
			name:      name,
			namespace: namespace,
			labels: map[string]string{
				"helm-broker-repo": "true",
			},
			data: map[string]string{
				"URLs": urls,
			},
		},
		clusterServiceClassExternalName: "testing",
		clusterServicePlanExternalNames: []string{"minimal", "full"},

		coreCli:  coreCli,
		svcatCli: svcatCli,
	}
}

func (t *TestBundleInstaller) Install() error {
	_, err := t.coreCli.ConfigMaps(t.configMap.namespace).Create(
		&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      t.configMap.name,
				Namespace: t.configMap.namespace,
				Labels:    t.configMap.labels,
			},
			Data: t.configMap.data,
		})

	if err != nil && !apiErrors.IsAlreadyExists(err) {
		return err
	}

	return nil
}

func (t *TestBundleInstaller) Uninstall() error {
	err := t.coreCli.ConfigMaps(t.configMap.namespace).Delete(t.configMap.name, &metav1.DeleteOptions{})
	if err != nil && !apiErrors.IsNotFound(err) {
		return err
	}

	return nil
}

func (t *TestBundleInstaller) WaitForTestBundleReady() error {
	err := t.waitForClusterServiceClass()
	if err != nil {
		return errors.Wrapf(err, "while waiting for ClusterServiceClass with externalName %s", t.clusterServiceClassExternalName)
	}

	err = t.waitForClusterServicePlans()
	if err != nil {
		return errors.Wrapf(err, "while waiting for ClusterServicePlans for ClusterServiceClass with externalName %s", t.clusterServiceClassExternalName)
	}

	return nil
}

func (t *TestBundleInstaller) waitForClusterServiceClass() error {
	return waiter.WaitAtMost(func() (bool, error) {
		classesList, err := t.svcatCli.ServicecatalogV1beta1().ClusterServiceClasses().List(metav1.ListOptions{})
		if err != nil {
			return false, err
		}

		for _, class := range classesList.Items {
			if class.GetExternalName() == t.clusterServiceClassExternalName {
				return true, nil
			}
		}

		return true, nil
	}, brokerReadyTimeout)
}

func (t *TestBundleInstaller) waitForClusterServicePlans() error {
	plansFound := map[string]bool{}

	for _, planName := range t.clusterServicePlanExternalNames {
		plansFound[planName] = false
	}

	err := waiter.WaitAtMost(func() (bool, error) {
		planList, err := t.svcatCli.ServicecatalogV1beta1().ClusterServicePlans().List(metav1.ListOptions{})
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
