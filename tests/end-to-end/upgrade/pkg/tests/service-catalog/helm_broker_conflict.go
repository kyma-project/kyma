package servicecatalog

import (
	"time"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	bu "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/injector"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	externalID           = "external-id"
	instanceName         = "redis"
	secondInstanceName   = "redis-second"
	conflictInstanceName = "redis-conflicting"
)

// HelmBrokerUpgradeConflictTest tests the Helm Broker business logic after Kyma upgrade phase
type HelmBrokerUpgradeConflictTest struct {
	ServiceCatalogInterface clientset.Interface
	K8sInterface            kubernetes.Interface
	BUInterface             bu.Interface
	aInjector               *injector.Addons
}

// NewHelmBrokerTest returns new instance of the HelmBrokerUpgradeConflictTest
func NewHelmBrokerConflictTest(aInjector *injector.Addons, k8sCli kubernetes.Interface, scCli clientset.Interface, buCli bu.Interface) *HelmBrokerUpgradeConflictTest {
	return &HelmBrokerUpgradeConflictTest{
		K8sInterface:            k8sCli,
		ServiceCatalogInterface: scCli,
		aInjector:               aInjector,
		BUInterface:             buCli,
	}
}

type helmBrokerConflictFlow struct {
	baseFlow
	scInterface clientset.Interface
}

// CreateResources creates resources needed for e2e upgrade test
func (ut *HelmBrokerUpgradeConflictTest) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	if err := ut.aInjector.InjectAddonsConfiguration(namespace); err != nil {
		return errors.Wrap(err, "while injecting addons configuration")
	}
	return ut.newFlow(stop, log, namespace).CreateResources()
}

// TestResources tests resources after backup phase
func (ut *HelmBrokerUpgradeConflictTest) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	if err := ut.newFlow(stop, log, namespace).TestResources(); err != nil {
		return err
	}

	if err := ut.aInjector.CleanupAddonsConfiguration(namespace); err != nil {
		return errors.Wrap(err, "while deleting addons configuration")
	}

	return nil
}

func (ut *HelmBrokerUpgradeConflictTest) newFlow(stop <-chan struct{}, log logrus.FieldLogger, namespace string) *helmBrokerConflictFlow {
	return &helmBrokerConflictFlow{
		baseFlow: baseFlow{
			log:          log,
			stop:         stop,
			namespace:    namespace,
			k8sInterface: ut.K8sInterface,
			scInterface:  ut.ServiceCatalogInterface,
			buInterface:  ut.BUInterface,
		},

		scInterface: ut.ServiceCatalogInterface,
	}
}

func (f *helmBrokerConflictFlow) CreateResources() error {
	// iterate over steps
	for _, fn := range []func() error{
		f.createFirstRedisInstance,
		f.waitFirstRedisInstance,
	} {
		err := fn()
		if err != nil {
			f.log.Errorln(err)
			f.logReport()
			return err
		}
	}
	return nil
}

func (f *helmBrokerConflictFlow) TestResources() error {
	// iterate over steps
	for _, fn := range []func() error{
		f.createSecondRedisInstance,
		f.waitSecondRedisInstance,
		f.createConflictingRedisInstance,
		f.waitConflictingRedisInstance,
		f.deleteRedisInstances,
		f.verifyRedisInstancesRemoved,
	} {
		err := fn()
		if err != nil {
			f.log.Errorln(err)
			f.logReport()
			return err
		}
	}
	return nil
}

func (f *helmBrokerConflictFlow) logReport() {
	f.logK8SReport()
	f.logServiceCatalogAndBindingUsageReport()
}

func (f *helmBrokerConflictFlow) createFirstRedisInstance() error {
	return f.createRedisInstance(instanceName, &runtime.RawExtension{})
}
func (f *helmBrokerConflictFlow) createSecondRedisInstance() error {
	return f.createRedisInstance(secondInstanceName, &runtime.RawExtension{
		Raw: []byte("app=true"),
	})
}
func (f *helmBrokerConflictFlow) createConflictingRedisInstance() error {
	return f.createRedisInstance(conflictInstanceName, &runtime.RawExtension{
		Raw: []byte("app=false"),
	})
}
func (f *helmBrokerConflictFlow) waitFirstRedisInstance() error {
	return f.waitForInstance(instanceName)
}
func (f *helmBrokerConflictFlow) waitSecondRedisInstance() error {
	return f.waitForInstance(secondInstanceName)
}
func (f *helmBrokerConflictFlow) waitConflictingRedisInstance() error {
	return f.waitForInstanceFail(conflictInstanceName)
}

func (f *helmBrokerConflictFlow) waitForRedisInstance(name string) error {
	f.log.Infof("Waiting for Redis instance to be ready")
	return f.waitForInstance(name)
}

func (f *helmBrokerConflictFlow) createRedisInstance(name string, params *runtime.RawExtension) error {
	f.log.Infof("Creating Redis service instance")

	return wait.Poll(time.Millisecond*500, time.Second*30, func() (done bool, err error) {
		if _, err = f.scInterface.ServicecatalogV1beta1().ServiceInstances(f.namespace).Create(&v1beta1.ServiceInstance{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ServiceInstance",
				APIVersion: "servicecatalog.k8s.io/v1beta1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Spec: v1beta1.ServiceInstanceSpec{
				PlanReference: v1beta1.PlanReference{
					ServiceClassExternalName: "redis",
					ServicePlanExternalName:  "micro",
				},
				ExternalID: externalID,
				Parameters: params,
			},
		}); err != nil {
			f.log.Errorf("while creating redis instance: %v", err)
			return false, nil
		}
		return true, nil
	})
}

func (f *helmBrokerConflictFlow) deleteRedisInstances() error {
	if err := f.deleteServiceInstance(instanceName); err != nil {
		return err
	}
	if err := f.deleteServiceInstance(secondInstanceName); err != nil {
		return err
	}
	return f.deleteServiceInstance(conflictInstanceName)
}

func (f *helmBrokerConflictFlow) verifyRedisInstancesRemoved() error {
	if err := f.waitForInstanceRemoved(instanceName); err != nil {
		return err
	}
	if err := f.waitForInstanceRemoved(secondInstanceName); err != nil {
		return err
	}
	return f.waitForInstanceRemoved(conflictInstanceName)}
