package servicecatalog

import (
	"bufio"
	"time"

	appConnectorApi "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	appBroker "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	appConnector "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	bu "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	k8sCoreTypes "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
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
}

// NewAppBrokerUpgradeTest returns new instance of the AppBrokerUpgradeTest
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

// CreateResources creates resources needed for e2e upgrade test
func (ut *AppBrokerUpgradeTest) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	return ut.newFlow(stop, log, namespace).CreateResources()
}

// TestResources tests resources after upgrade
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
		k8sInterface:          ut.K8sInterface,
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
		f.undeployEnvTester,
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
	report := NewReport(f.namespace, f.log)

	f.logK8SReport(report)
	f.logServiceCatalogAndBindingUsageReport(report)
	f.logApplicationReport(report)
	f.reportLogs(report)

	report.Print()
}

func (f *appBrokerFlow) deleteApplication() error {
	f.log.Infof("Removing Application %s", applicationName)
	return f.appConnectorInterface.ApplicationconnectorV1alpha1().Applications().Delete(applicationName, &metav1.DeleteOptions{})
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
	return f.waitForEnvInjected(appEnvTester, gatewayURL)
}

func (f *appBrokerFlow) verifyDeploymentDoesNotContainGatewayEnvs() error {
	f.log.Info("Verify deployment does not contain GATEWAY_URL environment variable")
	return f.waitForEnvNotInjected(appEnvTester)
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
	f.log.Infof("Removing Application Mapping %s", applicationName)
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

func (f *appBrokerFlow) undeployEnvTester() error {
	f.log.Infof("Removing deployment %s", appEnvTester)
	return f.deleteDeployment(appEnvTester)
}

func (f *appBrokerFlow) logApplicationReport(report *Report) {
	appMappings, err := f.appBrokerInterface.ApplicationconnectorV1alpha1().ApplicationMappings(f.namespace).List(metav1.ListOptions{})
	report.AddApplicationMappings(appMappings, err)

	eventActivations, err := f.appBrokerInterface.ApplicationconnectorV1alpha1().EventActivations(f.namespace).List(metav1.ListOptions{})
	report.AddEventActivations(eventActivations, err)

	applications, err := f.appConnectorInterface.ApplicationconnectorV1alpha1().Applications().List(metav1.ListOptions{})
	report.AddApplications(applications, err)
}

func (f *appBrokerFlow) logFromPod(namespace, labelSelector, container string) ([]string, error) {
	logs := make([]string, 0)

	pods, err := f.k8sInterface.CoreV1().Pods(namespace).List(metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return logs, errors.Wrapf(err, "unable to fetch logs for label %s", labelSelector)
	}
	if len(pods.Items) == 0 {
		return logs, errors.Wrapf(err, "could not find pod for label %s", labelSelector)
	}

	for _, p := range pods.Items {
		func() {
			req := f.k8sInterface.CoreV1().Pods(namespace).
				GetLogs(p.Name, &k8sCoreTypes.PodLogOptions{
					Container: container,
				})
			readCloser, err := req.Stream()
			if err != nil {
				f.log.Warnf("cannot get stream for pod %s: %s", p.Name, err)
				return
			}
			defer func() {
				err = readCloser.Close()
				if err != nil {
					f.log.Warnf("cannot close stream for pod %s: %s", p.Name, err)
				}
			}()

			rd := bufio.NewReader(readCloser)
			lineCounter := 0
			for {
				lineCounter++
				line, _, err := rd.ReadLine()
				if err != nil {
					f.log.Warnf("cannot read line %d from stream for pod %s: %s", lineCounter, p.Name, err)
					return
				}
				logs = append(logs, string(line))
			}
		}()
	}

	return logs, nil
}

func (f *appBrokerFlow) reportLogs(report *Report) {
	logs, err := f.logFromPod(integrationNamespace, "app=application-broker", "ctrl")
	report.AddLogs("ApplicationBroker", logs, []string{apiServiceID, eventsServiceID, applicationName}, err)

	logs, err = f.logFromPod("kyma-system", "app=service-catalog-addons-service-binding-usage-controller", "service-binding-usage-controller")
	report.AddLogs("ServiceBindingUsageController", logs, []string{apiBindingUsageName}, err)

	logs, err = f.logFromPod("kyma-system", "app=service-catalog-catalog-controller-manager", "controller-manager")
	report.AddLogs("ServiceCatalog", logs, []string{apiServiceID, eventsServiceID, apiBindingName}, err)
}
