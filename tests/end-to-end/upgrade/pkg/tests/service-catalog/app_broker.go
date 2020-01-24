package servicecatalog

import (
	"time"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	appConnectorApi "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	appBroker "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	appConnector "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	bu "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	messagingclientv1alpha1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/messaging/v1alpha1"
)

const (
	appEnvTester         = "app-env-tester"
	applicationName      = "application-for-testing"
	apiServiceID         = "api-service-id"
	eventsServiceID      = "events-service-id"
	apiBindingName       = "app-binding"
	apiBindingUsageName  = "app-binding"
	gatewayURL           = "https://gateway.local"
	integrationNamespace = "kyma-integration"
)

// AppBrokerUpgradeTest tests the Helm Broker business logic after Kyma upgrade phase
type AppBrokerUpgradeTest struct {
	ServiceCatalogInterface clientset.Interface
	K8sInterface            kubernetes.Interface
	BUInterface             bu.Interface
	AppBrokerInterface      appBroker.Interface
	AppConnectorInterface   appConnector.Interface
	MessagingInterface      messagingclientv1alpha1.MessagingV1alpha1Interface
}

// NewAppBrokerUpgradeTest returns new instance of the AppBrokerUpgradeTest
func NewAppBrokerUpgradeTest(scCli clientset.Interface,
	k8sCli kubernetes.Interface,
	buCli bu.Interface,
	abCli appBroker.Interface,
	acCli appConnector.Interface,
	msgCli messagingclientv1alpha1.MessagingV1alpha1Interface) *AppBrokerUpgradeTest {
	return &AppBrokerUpgradeTest{
		ServiceCatalogInterface: scCli,
		K8sInterface:            k8sCli,
		BUInterface:             buCli,
		AppBrokerInterface:      abCli,
		AppConnectorInterface:   acCli,
		MessagingInterface:      msgCli,
	}
}

type appBrokerFlow struct {
	baseFlow

	scInterface           clientset.Interface
	k8sInterface          kubernetes.Interface
	buInterface           bu.Interface
	appBrokerInterface    appBroker.Interface
	appConnectorInterface appConnector.Interface
	messagingInterface    messagingclientv1alpha1.MessagingV1alpha1Interface
}

// CreateResources creates resources needed for e2e upgrade test
func (ut *AppBrokerUpgradeTest) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	return ut.newFlow(stop, log, namespace).CreateResources()
}

// TestResources tests resources after backup phase
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
		messagingInterface:    ut.MessagingInterface,
	}
}

func (f *appBrokerFlow) CreateResources() error {
	// iterate over steps
	for _, fn := range []func() error{
		f.createChannel,
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
			f.log.Errorln(err)
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
		f.deleteChannel,
		f.waitForClassesRemoved,
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

func (f *appBrokerFlow) logReport() {
	f.logK8SReport()
	f.logServiceCatalogAndBindingUsageReport()
	f.logApplicationReport()
}

func (f *appBrokerFlow) deleteApplication() error {
	return f.appConnectorInterface.ApplicationconnectorV1alpha1().Applications().Delete(applicationName, &metav1.DeleteOptions{})
}

func (f *appBrokerFlow) createChannel() error {
	channel := &messagingv1alpha1.Channel{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Channel",
			APIVersion: "messaging.knative.dev/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: applicationName,
			Labels: map[string]string{
				"application-name": applicationName,
			},
		},
	}

	_, err := f.messagingInterface.Channels(integrationNamespace).Create(channel)
	return err
}

func (f *appBrokerFlow) deleteChannel() error {
	return f.messagingInterface.Channels(integrationNamespace).Delete(applicationName, &metav1.DeleteOptions{})
}

func (f *appBrokerFlow) createApplication() error {
	f.log.Info("Creating Application")
	return wait.Poll(time.Millisecond*500, time.Second*30, func() (done bool, err error) {
		if _, err = f.appConnectorInterface.ApplicationconnectorV1alpha1().Applications().Create(&v1alpha1.Application{
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
						ID:   apiServiceID,
						Name: apiServiceID,
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
								GatewayUrl:  gatewayURL,
							},
						},
					},
					{
						ID:   eventsServiceID,
						Name: eventsServiceID,
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
		}); err != nil {
			f.log.Errorf("while creating application: %v", err)
			return false, nil
		}
		return true, nil
	})
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
			Name: apiServiceID,
		},
		Spec: v1beta1.ServiceInstanceSpec{
			PlanReference: v1beta1.PlanReference{
				ServiceClassExternalName: apiServiceID,
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
			Name: eventsServiceID,
		},
		Spec: v1beta1.ServiceInstanceSpec{
			PlanReference: v1beta1.PlanReference{
				ServiceClassExternalName: eventsServiceID,
				ServicePlanExternalName:  "default",
			},
		},
	})
	return err
}

func (f *appBrokerFlow) createAPIServiceBinding() error {
	return f.createBindingAndWaitForReadiness(apiBindingName, apiServiceID)
}

func (f *appBrokerFlow) createAPIServiceBindingUsage() error {
	return f.createBindingUsageAndWaitForReadiness(apiBindingUsageName, apiBindingName, appEnvTester)
}

func (f *appBrokerFlow) verifyEventActivation() error {
	f.log.Infof("Waiting for the event activation")
	_, err := f.appBrokerInterface.ApplicationconnectorV1alpha1().EventActivations(f.namespace).Get(eventsServiceID, metav1.GetOptions{})
	return err
}

func (f *appBrokerFlow) verifyEnvTesterHasGatewayInjected() error {
	f.log.Info("Waiting for GATEWAY_URL env injected")
	return f.waitForEnvInjected(appEnvTester, "GATEWAY_URL", gatewayURL)
}

func (f *appBrokerFlow) verifyDeploymentDoesNotContainGatewayEnvs() error {
	f.log.Info("Verify deployment does not contain GATEWAY_URL environment variable")
	return f.waitForEnvNotInjected(appEnvTester, "GATEWAY_URL")
}

func (f *appBrokerFlow) deleteAPIServiceBindingUsage() error {
	return f.deleteBindingUsage(apiBindingUsageName)
}

func (f *appBrokerFlow) deleteAPIServiceBinding() error {
	return f.deleteServiceBinding(apiBindingName)
}

func (f *appBrokerFlow) deleteAPIServiceInstance() error {
	return f.deleteServiceInstance(apiServiceID)
}

func (f *appBrokerFlow) deleteEventsServiceInstance() error {
	return f.deleteServiceInstance(eventsServiceID)
}

func (f *appBrokerFlow) deleteApplicationMapping() error {
	return f.appBrokerInterface.ApplicationconnectorV1alpha1().ApplicationMappings(f.namespace).Delete(applicationName, &metav1.DeleteOptions{})
}

func (f *appBrokerFlow) waitForInstances() error {
	f.log.Infof("Waiting for instances to be ready")
	err := f.waitForInstance(apiServiceID)
	if err != nil {
		return err
	}
	err = f.waitForInstance(eventsServiceID)

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
			apiServiceID:    {},
			eventsServiceID: {},
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
			apiServiceID:    {},
			eventsServiceID: {},
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
			apiServiceID:    {},
			eventsServiceID: {},
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
	err := f.waitForInstanceRemoved(apiServiceID)
	if err != nil {
		return err
	}
	return f.waitForInstanceRemoved(eventsServiceID)
}

func (f *appBrokerFlow) waitForEnvTester() error {
	f.log.Info("Waiting for environment variable tester to be ready")
	return f.waitForDeployment(appEnvTester)
}

func (f *appBrokerFlow) logApplicationReport() {
	appMappings, err := f.appBrokerInterface.ApplicationconnectorV1alpha1().ApplicationMappings(f.namespace).List(metav1.ListOptions{})
	f.report("ApplicationMappings", appMappings, err)

	eventActivations, err := f.appBrokerInterface.ApplicationconnectorV1alpha1().EventActivations(f.namespace).List(metav1.ListOptions{})
	f.report("EventActivations", eventActivations, err)

	applications, err := f.appConnectorInterface.ApplicationconnectorV1alpha1().Applications().List(metav1.ListOptions{})
	f.report("Applications", applications, err)
}
