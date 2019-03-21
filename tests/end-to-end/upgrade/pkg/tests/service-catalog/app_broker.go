package service_catalog

import (
	"time"

	"k8s.io/client-go/kubernetes"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"

	appBroker "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	appConnector "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	bu "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned"
	appConnectorApi "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
)

const (
	appEnvTester        = "app-env-tester"
	applicationName     = "application-for-testing"
	apiServiceId        = "api-service-id"
	eventsServiceId     = "events-service-id"
	apiBindingName      = "app-binding"
	apiBindingUsageName = "app-binding"
	gatewayUrl          = "https://gateway.local"
)

type AppBrokerUpgradeTest struct {
	ServiceCatalogInterface clientset.Interface
	K8sInterface            kubernetes.Interface
	BUInterface             bu.Interface
	AppBrokerInterface      appBroker.Interface
	AppConnectorInterface   appConnector.Interface
}

func NewAppBrokerUpgradeTest(scCli clientset.Interface,
	k8sCli kubernetes.Interface,
	buCli bu.Interface,
	abCli appBroker.Interface,
	acCli appConnector.Interface) *AppBrokerUpgradeTest {
	return &AppBrokerUpgradeTest{
		ServiceCatalogInterface: scCli,
		K8sInterface:            k8sCli,
		BUInterface:             buCli,
		AppBrokerInterface:      abCli,
		AppConnectorInterface:   acCli,
	}
}

type appBrokerFlow struct {
	baseFlow

	scInterface           clientset.Interface
	k8sInterface          kubernetes.Interface
	buInterface           bu.Interface
	appBrokerInterface    appBroker.Interface
	appConnectorInterface appConnector.Interface
}

func (ut *AppBrokerUpgradeTest) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	return ut.newFlow(stop, log, namespace).CreateResources()
}

func (ut *AppBrokerUpgradeTest) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	return ut.newFlow(stop, log, namespace).TestResources()
}

func (ut *AppBrokerUpgradeTest) newFlow(stop <-chan struct{}, log logrus.FieldLogger, namespace string) *appBrokerFlow {
	return &appBrokerFlow{
		baseFlow: baseFlow{
			log:       log,
			stop:      stop,
			namespace: namespace,

			buInterface:  ut.BUInterface,
			scInterface:  ut.ServiceCatalogInterface,
			k8sInterface: ut.K8sInterface,
		},

		appBrokerInterface:    ut.AppBrokerInterface,
		appConnectorInterface: ut.AppConnectorInterface,
		scInterface:           ut.ServiceCatalogInterface,
	}
}

func (f *appBrokerFlow) CreateResources() error {
	// iterate over steps
	for _, fn := range []func() error{
		f.createApplication,
		f.createApplicationMapping,
		f.deployEnvTester,
		f.waitForEnvTester,
		f.waitForClassAndPlans,
		f.createAPIServiceInstance,
		f.createEventsServiceInstance,
		f.waitForInstances,
		f.createAPIServiceBinding,
		f.createAPIServiceBindingUsage,
	} {
		err := fn()
		if err != nil {
			f.logReport()
			return err
		}
	}
	return nil
}

func (f *appBrokerFlow) TestResources() error {
	// iterate over steps
	for _, fn := range []func() error{
		f.verifyEventActivation,
		f.verifyEnvTesterHasGatewayInjected,
		f.deleteAPIServiceBindingUsage,
		f.verifyDeploymentDoesNotContainGatewayEnvs,
		f.deleteAPIServiceBinding,
		f.deleteAPIServiceInstance,
		f.deleteEventsServiceInstance,
		f.waitForInstancesDeleted,
		f.deleteApplicationMapping,
		f.deleteApplication,
		f.waitForClassesRemoved,
	} {
		err := fn()
		if err != nil {
			f.logReport()
			return err
		}
	}
	return nil
}

func (f *appBrokerFlow) logReport() {
	f.logK8SReport()
	f.logServiceCatalogAndBindingUsageReport()
	f.logApplicationReport()
}

func (f *appBrokerFlow) deleteApplication() error {
	return f.appConnectorInterface.ApplicationconnectorV1alpha1().Applications().Delete(applicationName, &metav1.DeleteOptions{})
}

func (f *appBrokerFlow) createApplication() error {
	f.log.Info("Creating Application")
	_, err := f.appConnectorInterface.ApplicationconnectorV1alpha1().Applications().Create(&v1alpha1.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Application",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: applicationName,
		},
		Spec: v1alpha1.ApplicationSpec{
			AccessLabel:      "app-access-label",
			Description:      "Application used by application acceptance test",
			SkipInstallation: true,
			Services: []v1alpha1.Service{
				{
					ID:   apiServiceId,
					Name: apiServiceId,
					Labels: map[string]string{
						"connected-app": "app-name",
					},
					ProviderDisplayName: "provider",
					DisplayName:         "Some API",
					Description:         "Application Service Class with API",
					Tags:                []string{},
					Entries: []v1alpha1.Entry{
						{
							Type:        "API",
							AccessLabel: "accessLabel",
							GatewayUrl:  gatewayUrl,
						},
					},
				},
				{
					ID:   eventsServiceId,
					Name: eventsServiceId,
					Labels: map[string]string{
						"connected-app": "app-name",
					},
					ProviderDisplayName: "provider",
					DisplayName:         "Some Events",
					Description:         "Application Service Class with Events",
					Tags:                []string{},
					Entries: []v1alpha1.Entry{
						{
							Type: "Events",
						},
					},
				},
			},
		},
	})
	return err
}

func (f *appBrokerFlow) deployEnvTester() error {
	return f.createEnvTester(appEnvTester, "GATEWAY_URL")
}

func (f *appBrokerFlow) createApplicationMapping() error {
	f.log.Info("Creating ApplicationMapping")
	_, err := f.appBrokerInterface.ApplicationconnectorV1alpha1().ApplicationMappings(f.namespace).Create(&appConnectorApi.ApplicationMapping{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ApplicationMapping",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: applicationName,
		},
	})
	return err
}

func (f *appBrokerFlow) createAPIServiceInstance() error {
	f.log.Infof("Creating API service instance")
	_, err := f.scInterface.ServicecatalogV1beta1().ServiceInstances(f.namespace).Create(&v1beta1.ServiceInstance{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceInstance",
			APIVersion: "servicecatalog.k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: apiServiceId,
		},
		Spec: v1beta1.ServiceInstanceSpec{
			PlanReference: v1beta1.PlanReference{
				ServiceClassExternalName: apiServiceId,
				ServicePlanExternalName:  "default",
			},
		},
	})
	return err
}

func (f *appBrokerFlow) createEventsServiceInstance() error {
	f.log.Infof("Creating Events service instance")
	_, err := f.scInterface.ServicecatalogV1beta1().ServiceInstances(f.namespace).Create(&v1beta1.ServiceInstance{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceInstance",
			APIVersion: "servicecatalog.k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: eventsServiceId,
		},
		Spec: v1beta1.ServiceInstanceSpec{
			PlanReference: v1beta1.PlanReference{
				ServiceClassExternalName: eventsServiceId,
				ServicePlanExternalName:  "default",
			},
		},
	})
	return err
}

func (f *appBrokerFlow) createAPIServiceBinding() error {
	return f.createBindingAndWaitForReadiness(apiBindingName, apiServiceId)
}

func (f *appBrokerFlow) createAPIServiceBindingUsage() error {
	return f.createBindingUsageAndWaitForReadiness(apiBindingUsageName, apiBindingName, appEnvTester)
}

func (f *appBrokerFlow) verifyEventActivation() error {
	f.log.Infof("Waiting for the event activation")
	_, err := f.appBrokerInterface.ApplicationconnectorV1alpha1().EventActivations(f.namespace).Get(eventsServiceId, metav1.GetOptions{})
	return err
}

func (f *appBrokerFlow) verifyEnvTesterHasGatewayInjected() error {
	f.log.Info("Waiting for GATEWAY_URL env injected")
	return f.waitForEnvInjected(appEnvTester, "GATEWAY_URL", gatewayUrl)
}

func (f *appBrokerFlow) verifyDeploymentDoesNotContainGatewayEnvs() error {
	f.log.Info("Verify deployment does not contain GATEWAY_URL environment variable")
	return f.waitForEnvNotInjected(envTesterName, "GATEWAY_URL")
}

func (f *appBrokerFlow) deleteAPIServiceBindingUsage() error {
	return f.deleteBindingUsage(apiBindingUsageName)
}

func (f *appBrokerFlow) deleteAPIServiceBinding() error {
	return f.deleteServiceBinding(apiBindingName)
}

func (f *appBrokerFlow) deleteAPIServiceInstance() error {
	return f.deleteServiceInstance(apiServiceId)
}

func (f *appBrokerFlow) deleteEventsServiceInstance() error {
	return f.deleteServiceInstance(eventsServiceId)
}

func (f *appBrokerFlow) deleteApplicationMapping() error {
	return f.appBrokerInterface.ApplicationconnectorV1alpha1().ApplicationMappings(f.namespace).Delete(applicationName, &metav1.DeleteOptions{})
}

func (f *appBrokerFlow) waitForInstances() error {
	f.log.Infof("Waiting for instances to be ready")
	err := f.waitForInstance(apiServiceId)
	if err != nil {
		return err
	}
	err = f.waitForInstance(eventsServiceId)

	return err
}

func (f *appBrokerFlow) waitForClassesRemoved() error {
	return f.wait(5*time.Second, func() (bool, error) {
		classes, err := f.scInterface.ServicecatalogV1beta1().ServiceClasses(f.namespace).List(metav1.ListOptions{})
		if err != nil {
			return false, err
		}
		if len(classes.Items) == 0 {
			return true, nil
		}
		return false, nil
	})
}

func (f *appBrokerFlow) waitForClassAndPlans() error {
	f.log.Infof("Waiting for classes")
	err := f.wait(45*time.Second, func() (bool, error) {
		expectedClasses := map[string]struct{}{
			apiServiceId:    {},
			eventsServiceId: {},
		}
		classes, err := f.scInterface.ServicecatalogV1beta1().ServiceClasses(f.namespace).List(metav1.ListOptions{})
		if err != nil {
			return false, err
		}
		for _, class := range classes.Items {
			delete(expectedClasses, class.Spec.ExternalName)
		}
		if len(expectedClasses) == 0 {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return err
	}
	f.log.Infof("Waiting for plans")
	return f.wait(10*time.Second, func() (bool, error) {
		expectedPlans := map[string]struct{}{
			apiServiceId:    {},
			eventsServiceId: {},
		}
		plans, err := f.scInterface.ServicecatalogV1beta1().ServicePlans(f.namespace).List(metav1.ListOptions{})
		if err != nil {
			return false, err
		}
		for _, plan := range plans.Items {
			delete(expectedPlans, plan.Spec.ServiceClassRef.Name)
		}
		if len(expectedPlans) == 0 {
			return true, nil
		}

		return false, nil
	})
}

func (f *appBrokerFlow) waitForClasses() error {
	return f.wait(2*time.Minute, func() (bool, error) {
		expectedClasses := map[string]struct{}{
			apiServiceId:    {},
			eventsServiceId: {},
		}
		classes, err := f.scInterface.ServicecatalogV1beta1().ServiceClasses(f.namespace).List(metav1.ListOptions{})
		if err != nil {
			return false, err
		}
		for _, class := range classes.Items {
			delete(expectedClasses, class.Spec.ExternalName)
		}
		if len(expectedClasses) == 0 {
			return true, nil
		}
		return false, nil
	})
}

func (f *appBrokerFlow) waitForInstancesDeleted() error {
	err := f.waitForInstanceRemoved(apiServiceId)
	if err != nil {
		return err
	}
	return f.waitForInstanceRemoved(eventsServiceId)
}

func (f *appBrokerFlow) waitForEnvTester() error {
	f.log.Info("Waiting for environment variable tester to be ready")
	return f.waitForDeployment(appEnvTester)
}

func (f *appBrokerFlow) logApplicationReport() {
	appMappings, err := f.appBrokerInterface.ApplicationconnectorV1alpha1().ApplicationMappings(f.namespace).List(metav1.ListOptions{})
	f.report("ApplicationMappings", appMappings , err)

	eventActivations, err := f.appBrokerInterface.ApplicationconnectorV1alpha1().EventActivations(f.namespace).List(metav1.ListOptions{})
	f.report("EventActivations", eventActivations , err)

	applications, err := f.appConnectorInterface.ApplicationconnectorV1alpha1().Applications().List(metav1.ListOptions{})
	f.report("Applications", applications , err)
}