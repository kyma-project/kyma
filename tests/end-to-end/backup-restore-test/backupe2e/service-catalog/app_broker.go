package service_catalog

import (
	"fmt"
	"os"
	"time"

	"github.com/avast/retry-go"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
	messagingclientv1alpha1 "knative.dev/eventing/pkg/client/clientset/versioned/typed/messaging/v1alpha1"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	sc "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	abApi "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	ab "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	appBroker "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	ao "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	appConnector "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	bu "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned"
	"github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	applicationName      = "application-for-testing"
	apiServiceId         = "api-service-id"
	eventsServiceId      = "events-service-id"
	apiBindingName       = "app-binding"
	apiBindingUsageName  = "app-binding"
	gatewayUrl           = "https://gateway.local"
	integrationNamespace = "kyma-integration"
)

type AppBrokerTest struct {
	serviceCatalogInterface clientset.Interface
	k8sInterface            kubernetes.Interface
	buInterface             bu.Interface
	appBrokerInterface      appBroker.Interface
	appConnectorInterface   appConnector.Interface
	messagingInterface      messagingclientv1alpha1.MessagingV1alpha1Interface
}

type appBrokerFlow struct {
	brokersFlow

	scInterface           clientset.Interface
	k8sInterface          kubernetes.Interface
	buInterface           bu.Interface
	appBrokerInterface    appBroker.Interface
	appConnectorInterface appConnector.Interface
	messagingInterface    messagingclientv1alpha1.MessagingV1alpha1Interface
}

func NewAppBrokerTest() (AppBrokerTest, error) {
	kubeconfig := os.Getenv("KUBECONFIG")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return AppBrokerTest{}, err
	}
	k8sCS, err := kubernetes.NewForConfig(config)
	if err != nil {
		return AppBrokerTest{}, err
	}
	scCS, err := sc.NewForConfig(config)
	if err != nil {
		return AppBrokerTest{}, err
	}
	buSc, err := bu.NewForConfig(config)
	if err != nil {
		return AppBrokerTest{}, err
	}
	appConnectorCS, err := ao.NewForConfig(config)
	if err != nil {
		return AppBrokerTest{}, err
	}
	appBrokerCS, err := ab.NewForConfig(config)
	if err != nil {
		return AppBrokerTest{}, err
	}
	msgCS, err := messagingclientv1alpha1.NewForConfig(config)
	if err != nil {
		return AppBrokerTest{}, err
	}

	return AppBrokerTest{
		k8sInterface:            k8sCS,
		appBrokerInterface:      appBrokerCS,
		appConnectorInterface:   appConnectorCS,
		buInterface:             buSc,
		serviceCatalogInterface: scCS,
		messagingInterface:      msgCS,
	}, nil
}

func (t AppBrokerTest) CreateResources(namespace string) {
	t.newFlow(namespace).createResources()
}

func (t AppBrokerTest) TestResources(namespace string) {
	t.newFlow(namespace).testResources()
}

func (t *AppBrokerTest) newFlow(namespace string) *appBrokerFlow {
	return &appBrokerFlow{
		brokersFlow: brokersFlow{
			namespace: namespace,
			log:       logrus.WithField("test", "app-broker"),

			buInterface:  t.buInterface,
			scInterface:  t.serviceCatalogInterface,
			k8sInterface: t.k8sInterface,
		},

		appBrokerInterface:    t.appBrokerInterface,
		appConnectorInterface: t.appConnectorInterface,
		scInterface:           t.serviceCatalogInterface,
		messagingInterface:    t.messagingInterface,
	}
}

func (f *appBrokerFlow) createResources() {
	for _, fn := range []func() error{
		f.createChannel,
		f.waitForChannel,
		f.createApplication,
		f.createApplicationMapping,
		f.deployEnvTester,
		f.waitForClassAndPlans,
		f.createAPIServiceInstance,
		f.createEventsServiceInstance,
		f.waitForAppInstances,
		f.createAPIServiceBindingAndWaitForReadiness,
		f.createAPIServiceBindingUsageAndWaitForReadiness,
	} {
		err := fn()
		if err != nil {
			f.logReport()
		}
		So(err, ShouldBeNil)
	}
}

func (f *appBrokerFlow) testResources() {
	for _, fn := range []func() error{
		// channels are not backed up, so we need to create the required one in this testcase
		f.createChannel,
		f.waitForChannel,
		f.verifyApplication,
		f.waitForClassAndPlans,
		f.waitForAppInstances,
		f.verifyEventActivation,
		f.verifyEnvTesterHasGatewayInjected,
		f.deleteAPIServiceBindingUsage,
		f.verifyEnvTesterHasGatewayNotInjected,

		// we create again APIServiceBindingUsage to restore it after the tests
		f.createAPIServiceBindingUsage,
	} {
		err := fn()
		if err != nil {
			f.logReport()
		}
		So(err, ShouldBeNil)
	}
}

func (f *appBrokerFlow) deployEnvTester() error {
	return f.createEnvTester("GATEWAY_URL")
}

func (f *appBrokerFlow) createApplication() error {
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

// createChannel creates a Knative channel that would normally be installed by the application controller upon helm-chart installation. As we skip the installation part,
// we need to ensure that the channel exists as it is required by the knative broker.
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
	if errors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

func (f *appBrokerFlow) waitForChannel() error {
	return retry.Do(
		func() error {
			channel, err := f.messagingInterface.Channels(integrationNamespace).Get(applicationName, metav1.GetOptions{})
			if err != nil {
				return err
			}

			if !channel.Status.IsReady() {
				return fmt.Errorf("channel not ready: %+v", channel.Status)
			}
			return nil
		},
	)
}

func (f *appBrokerFlow) deleteChannel() error {
	return f.messagingInterface.Channels(integrationNamespace).Delete(applicationName, &metav1.DeleteOptions{})
}

func (f *appBrokerFlow) createApplicationMapping() error {
	_, err := f.appBrokerInterface.ApplicationconnectorV1alpha1().ApplicationMappings(f.namespace).Create(&abApi.ApplicationMapping{
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

func (f *appBrokerFlow) createAPIServiceBindingAndWaitForReadiness() error {
	return f.createBindingAndWaitForReadiness(apiBindingName, apiServiceId)
}

func (f *appBrokerFlow) createAPIServiceBindingUsage() error {
	return f.createBindingUsage(apiBindingUsageName, apiBindingName)
}

func (f *appBrokerFlow) createAPIServiceBindingUsageAndWaitForReadiness() error {
	return f.createBindingUsageAndWaitForReadiness(apiBindingUsageName, apiBindingName)
}

func (f *appBrokerFlow) verifyApplication() error {
	f.log.Infof("Verifying if Application exists %s", applicationName)
	return f.wait(time.Minute, func() (done bool, err error) {
		_, e := f.appConnectorInterface.ApplicationconnectorV1alpha1().Applications().Get(applicationName, metav1.GetOptions{})
		if e != nil {
			return false, e
		}
		return true, nil
	})
}

func (f *appBrokerFlow) verifyEventActivation() error {
	f.log.Infof("Verifying if EventActivation exists %s", applicationName)
	return f.wait(time.Minute, func() (done bool, err error) {
		_, e := f.appBrokerInterface.ApplicationconnectorV1alpha1().ApplicationMappings(f.namespace).Get(applicationName, metav1.GetOptions{})
		if e != nil {
			return false, e
		}
		return true, nil
	})
}

func (f *appBrokerFlow) verifyEnvTesterHasGatewayInjected() error {
	return f.waitForEnvInjected("GATEWAY_URL", gatewayUrl)
}

func (f *appBrokerFlow) verifyEnvTesterHasGatewayNotInjected() error {
	return f.waitForEnvNotInjected("GATEWAY_URL")
}

func (f *appBrokerFlow) deleteAPIServiceBindingUsage() error {
	return f.deleteBindingUsage(apiBindingUsageName)
}

func (f *appBrokerFlow) waitForAppInstances() error {
	f.log.Infof("Waiting for instances to be ready")
	err := f.waitForInstance(apiServiceId)
	if err != nil {
		return err
	}
	err = f.waitForInstance(eventsServiceId)

	return err
}

func (f *appBrokerFlow) waitForClassAndPlans() error {
	f.log.Infof("Waiting for classes")
	if err := f.wait(2*time.Minute, func() (bool, error) {
		expectedClasses := map[string]struct{}{
			apiServiceId:    {},
			eventsServiceId: {},
		}
		classes, err := f.scInterface.ServicecatalogV1beta1().ServiceClasses(f.namespace).List(metav1.ListOptions{})
		if err != nil {
			return false, err
		}
		if len(classes.Items) != 2 {
			f.log.Warnf("should have 2 classes not %v", len(classes.Items))
			return false, nil
		}
		for _, class := range classes.Items {
			if _, ok := expectedClasses[class.Spec.ExternalName]; !ok {
				f.log.Warnf("following class is not defined: %s", class.Spec.ExternalName)
				return false, nil
			}
		}
		return true, nil
	}); err != nil {
		return err
	}

	f.log.Infof("Waiting for plans")
	return f.wait(15*time.Second, func() (bool, error) {
		expectedPlans := map[string]struct{}{
			apiServiceId:    {},
			eventsServiceId: {},
		}
		plans, err := f.scInterface.ServicecatalogV1beta1().ServicePlans(f.namespace).List(metav1.ListOptions{})
		if err != nil {
			return false, err
		}

		if len(plans.Items) != 2 {
			f.log.Warnf("should have 2 plan not %v", len(plans.Items))
			return false, nil
		}
		for _, plan := range plans.Items {
			if _, ok := expectedPlans[plan.Spec.ServiceClassRef.Name]; !ok {
				f.log.Warnf("following class is not defined: %s", plan.Spec.ServiceClassRef.Name)
				return false, nil
			}
		}

		return true, nil
	})
}

func (f *appBrokerFlow) logReport() {
	f.logK8SReport()
	f.logServiceCatalogAndBindingUsageReport()
	f.logApplicationReport()
}

func (f *appBrokerFlow) logApplicationReport() {
	appMappings, err := f.appBrokerInterface.ApplicationconnectorV1alpha1().ApplicationMappings(f.namespace).List(metav1.ListOptions{})
	f.report("ApplicationMappings", appMappings, err)

	eventActivations, err := f.appBrokerInterface.ApplicationconnectorV1alpha1().EventActivations(f.namespace).List(metav1.ListOptions{})
	f.report("EventActivations", eventActivations, err)

	applications, err := f.appConnectorInterface.ApplicationconnectorV1alpha1().Applications().List(metav1.ListOptions{})
	f.report("Applications", applications, err)
}
