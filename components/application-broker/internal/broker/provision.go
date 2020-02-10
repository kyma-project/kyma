package broker

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/pkg/errors"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/sirupsen/logrus"

	istioauthenticationalpha1 "istio.io/api/authentication/v1alpha1"
	istiov1alpha1 "istio.io/client-go/pkg/apis/authentication/v1alpha1"
	istioversionedclient "istio.io/client-go/pkg/clientset/versioned"

	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/internal/access"
	"github.com/kyma-project/kyma/components/application-broker/internal/knative"
	"github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	v1client "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
)

const (
	integrationNamespace     = "kyma-integration"
	serviceCatalogAPIVersion = "servicecatalog.k8s.io/v1beta1"

	// knativeEventingInjectionLabelKey used for enabling Knative eventing default broker for a given namespace
	knativeEventingInjectionLabelKey          = "knative-eventing-injection"
	knativeEventingInjectionLabelValueEnabled = "enabled"

	// applicationNameLabelKey is used to selected Knative Channels and Subscriptions
	applicationNameLabelKey = "application-name"

	// brokerNamespaceLabelKey is used to selected Knative Subscriptions
	brokerNamespaceLabelKey = "broker-namespace"

	// knSubscriptionNamePrefix is the prefix used for the generated Knative Subscription name
	knSubscriptionNamePrefix = "brokersub"

	// brokerTargetSelectorName used for targeting the default-broker svc while creating an istio policy
	brokerTargetSelectorName = "default-broker"
	// filterTargetSelectorName used for targeting the default-broker-filter svc while creating an istio policy
	filterTargetSelectorName = "default-broker-filter"

	istioMtlsPermissiveMode = 1

	policyNameSuffix = "-broker"
)

// NewProvisioner creates provisioner
func NewProvisioner(instanceInserter instanceInserter, instanceGetter instanceGetter,
	instanceStateGetter instanceStateGetter, operationInserter operationInserter, operationUpdater operationUpdater,
	accessChecker access.ProvisionChecker, appSvcFinder appSvcFinder, serviceInstanceGetter serviceInstanceGetter,
	eaClient v1client.ApplicationconnectorV1alpha1Interface, knClient knative.Client,
	istioClient istioversionedclient.Interface, iStateUpdater instanceStateUpdater,
	operationIDProvider func() (internal.OperationID, error), log logrus.FieldLogger) *ProvisionService {
	return &ProvisionService{
		instanceInserter:      instanceInserter,
		instanceGetter:        instanceGetter,
		instanceStateGetter:   instanceStateGetter,
		instanceStateUpdater:  iStateUpdater,
		operationInserter:     operationInserter,
		operationUpdater:      operationUpdater,
		operationIDProvider:   operationIDProvider,
		accessChecker:         accessChecker,
		appSvcFinder:          appSvcFinder,
		eaClient:              eaClient,
		knClient:              knClient,
		istioClient:           istioClient,
		serviceInstanceGetter: serviceInstanceGetter,
		maxWaitTime:           time.Minute,
		log:                   log.WithField("service", "provisioner"),
	}
}

// ProvisionService performs provisioning action
type ProvisionService struct {
	instanceInserter      instanceInserter
	instanceGetter        instanceGetter
	instanceStateUpdater  instanceStateUpdater
	operationInserter     operationInserter
	operationUpdater      operationUpdater
	instanceStateGetter   instanceStateGetter
	operationIDProvider   func() (internal.OperationID, error)
	appSvcFinder          appSvcFinder
	eaClient              v1client.ApplicationconnectorV1alpha1Interface
	accessChecker         access.ProvisionChecker
	serviceInstanceGetter serviceInstanceGetter
	knClient              knative.Client
	istioClient           istioversionedclient.Interface

	mu sync.Mutex

	maxWaitTime time.Duration
	log         logrus.FieldLogger
	asyncHook   func()
}

// Provision action
func (svc *ProvisionService) Provision(ctx context.Context, osbCtx osbContext, req *osb.ProvisionRequest) (*osb.ProvisionResponse, *osb.HTTPStatusCodeError) {
	if len(req.Parameters) > 0 {
		return nil, &osb.HTTPStatusCodeError{StatusCode: http.StatusBadRequest, ErrorMessage: strPtr("application-broker does not support configuration options for provisioning")}
	}
	if !req.AcceptsIncomplete {
		return nil, &osb.HTTPStatusCodeError{StatusCode: http.StatusBadRequest, ErrorMessage: strPtr("asynchronous operation mode required")}
	}

	svc.mu.Lock()
	defer svc.mu.Unlock()

	iID := internal.InstanceID(req.InstanceID)

	switch state, err := svc.instanceStateGetter.IsProvisioned(iID); true {
	case err != nil:
		return nil, &osb.HTTPStatusCodeError{StatusCode: http.StatusInternalServerError, ErrorMessage: strPtr(fmt.Sprintf("while checking if instance is already provisioned: %v", err))}
	case state:
		return &osb.ProvisionResponse{Async: false}, nil
	}

	switch opIDInProgress, inProgress, err := svc.instanceStateGetter.IsProvisioningInProgress(iID); true {
	case err != nil:
		return nil, &osb.HTTPStatusCodeError{StatusCode: http.StatusInternalServerError, ErrorMessage: strPtr(fmt.Sprintf("while checking if instance is being provisioned: %v", err))}
	case inProgress:
		opKeyInProgress := osb.OperationKey(opIDInProgress)
		return &osb.ProvisionResponse{Async: true, OperationKey: &opKeyInProgress}, nil
	}

	id, err := svc.operationIDProvider()
	if err != nil {
		return nil, &osb.HTTPStatusCodeError{StatusCode: http.StatusInternalServerError, ErrorMessage: strPtr(fmt.Sprintf("while generating ID for operation: %v", err))}
	}

	op := internal.InstanceOperation{
		InstanceID:  iID,
		OperationID: id,
		Type:        internal.OperationTypeCreate,
		State:       internal.OperationStateInProgress,
	}

	if err := svc.operationInserter.Insert(&op); err != nil {
		return nil, &osb.HTTPStatusCodeError{StatusCode: http.StatusInternalServerError, ErrorMessage: strPtr(fmt.Sprintf("while inserting instance operation to storage: %v", err))}
	}

	svcID := internal.ServiceID(req.ServiceID)
	svcPlanID := internal.ServicePlanID(req.PlanID)

	app, err := svc.appSvcFinder.FindOneByServiceID(internal.ApplicationServiceID(req.ServiceID))
	if err != nil {
		return nil, &osb.HTTPStatusCodeError{StatusCode: http.StatusInternalServerError, ErrorMessage: strPtr(fmt.Sprintf("while getting application with id: %s to storage: %v", req.ServiceID, err))}
	}

	namespace := internal.Namespace(osbCtx.BrokerNamespace)

	service, err := getSvcByID(app.Services, internal.ApplicationServiceID(req.ServiceID))
	if err != nil {
		return nil, &osb.HTTPStatusCodeError{StatusCode: http.StatusBadRequest, ErrorMessage: strPtr(fmt.Sprintf("while getting service [%s] from Application [%s]: %v", req.ServiceID, app.Name, err))}
	}

	i := internal.Instance{
		ID:            iID,
		Namespace:     namespace,
		ServiceID:     svcID,
		ServicePlanID: svcPlanID,
		State:         internal.InstanceStatePending,
	}

	if err = svc.instanceInserter.Insert(&i); err != nil {
		return nil, &osb.HTTPStatusCodeError{StatusCode: http.StatusBadRequest, ErrorMessage: strPtr(fmt.Sprintf("while inserting instance to storage: %v", err))}
	}

	opKey := osb.OperationKey(op.OperationID)
	resp := &osb.ProvisionResponse{
		OperationKey: &opKey,
		Async:        true,
	}

	svc.doAsync(iID, id, app.Name, getApplicationServiceID(req), namespace, service.EventProvider, service.DisplayName)
	return resp, nil
}

func getApplicationServiceID(req *osb.ProvisionRequest) internal.ApplicationServiceID {
	return internal.ApplicationServiceID(req.ServiceID)
}

func (svc *ProvisionService) doAsync(iID internal.InstanceID, opID internal.OperationID, appName internal.ApplicationName, appID internal.ApplicationServiceID, ns internal.Namespace, eventProvider bool, displayName string) {
	go svc.do(iID, opID, appName, appID, ns, eventProvider, displayName)
}

func (svc *ProvisionService) do(iID internal.InstanceID, opID internal.OperationID, appName internal.ApplicationName, appID internal.ApplicationServiceID, ns internal.Namespace, eventProvider bool, displayName string) {
	if svc.asyncHook != nil {
		defer svc.asyncHook()
	}
	instanceState := internal.InstanceStateSucceeded
	opState := internal.OperationStateSucceeded
	opDesc := internal.OperationDescriptionProvisioningSucceeded

	canProvisionOutput, err := svc.accessChecker.CanProvision(iID, appID, ns, svc.maxWaitTime)
	svc.log.Infof("Access checker: canProvisionInstance(appName=[%s], appID=[%s], ns=[%s]) returned: canProvisionOutput=[%+v], error=[%v]", appName, appID, ns, canProvisionOutput, err)
	if err != nil {
		instanceState = internal.InstanceStateFailed
		opState = internal.OperationStateFailed
		opDesc = fmt.Sprintf("provisioning failed on error: %s", err)
		svc.updateStates(iID, opID, instanceState, opState, opDesc)
		return
	}

	if !canProvisionOutput.Allowed {
		instanceState = internal.InstanceStateFailed
		opState = internal.OperationStateFailed
		opDesc = fmt.Sprintf("Forbidden provisioning instance [%s] for application [name: %s, id: %s] in namespace: [%s]. Reason: [%s]", iID, appName, appID, ns, canProvisionOutput.Reason)
		svc.updateStates(iID, opID, instanceState, opState, opDesc)
		return
	}

	if !eventProvider {
		svc.updateStates(iID, opID, instanceState, opState, opDesc)
		return
	}

	// create Kyma EventActivation
	if err := svc.createEaOnSuccessProvision(appName, appID, ns, displayName, iID); err != nil {
		instanceState = internal.InstanceStateFailed
		opState = internal.OperationStateFailed
		opDesc = fmt.Sprintf("provisioning failed while creating EventActivation on error: %s", err)
		svc.updateStates(iID, opID, instanceState, opState, opDesc)
		return
	}

	// persist Knative Subscription
	if err := svc.persistKnativeSubscription(appName, ns); err != nil {
		svc.log.Printf("Error persisting Knative Subscription: %v", err)
		instanceState = internal.InstanceStateFailed
		opState = internal.OperationStateFailed
		opDesc = fmt.Sprintf("provisioning failed while persisting Knative Subscription for application: %s namespace: %s on error: %s", appName, ns, err)
		svc.updateStates(iID, opID, instanceState, opState, opDesc)
		return
	}

	// enable the namespace default Knative Broker
	if err := svc.enableDefaultKnativeBroker(ns); err != nil {
		instanceState = internal.InstanceStateFailed
		opState = internal.OperationStateFailed
		opDesc = fmt.Sprintf("provisioning failed while enabling default Knative Broker for namespace: %s"+
			" on error: %s", ns, err)
		svc.updateStates(iID, opID, instanceState, opState, opDesc)
		return
	}

	// Create istio policy
	if err := svc.createIstioPolicy(ns); err != nil {
		svc.log.Errorf("Error creating istio policy: %v", err)
		instanceState = internal.InstanceStateFailed
		opState = internal.OperationStateFailed
		opDesc = fmt.Sprintf("provisioning failed while creating an istio policy for application: %s"+
			" namespace: %s on error: %s", appName, ns, err)
		svc.updateStates(iID, opID, instanceState, opState, opDesc)
		return
	}

	svc.updateStates(iID, opID, instanceState, opState, opDesc)
}

func (svc *ProvisionService) updateStates(iID internal.InstanceID, opID internal.OperationID,
	instanceState internal.InstanceState, opState internal.OperationState, opDesc string) {

	if err := svc.instanceStateUpdater.UpdateState(iID, instanceState); err != nil {
		svc.log.Errorf("Cannot update state of the stored instance [%s]: [%v]", iID, err)
	}

	if err := svc.operationUpdater.UpdateStateDesc(iID, opID, opState, &opDesc); err != nil {
		svc.log.Errorf("Cannot update state for ServiceInstance [%s]: [%v]", iID, err)
	}
}

func (svc *ProvisionService) createEaOnSuccessProvision(appName internal.ApplicationName,
	appID internal.ApplicationServiceID, ns internal.Namespace, displayName string, iID internal.InstanceID) error {

	// instance ID is the serviceInstance.Spec.ExternalID
	si, err := svc.serviceInstanceGetter.GetByNamespaceAndExternalID(string(ns), string(iID))
	if err != nil {
		return errors.Wrapf(err, "while getting service instance with external id: %q in namespace: %q", iID, ns)
	}
	ea := &v1alpha1.EventActivation{
		TypeMeta: v1.TypeMeta{
			Kind:       "EventActivation",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      string(appID),
			Namespace: string(ns),
			OwnerReferences: []v1.OwnerReference{
				{
					APIVersion: serviceCatalogAPIVersion,
					Kind:       "ServiceInstance",
					Name:       si.Name,
					UID:        si.UID,
				},
			},
		},
		Spec: v1alpha1.EventActivationSpec{
			DisplayName: displayName,
			SourceID:    string(appName),
		},
	}
	_, err = svc.eaClient.EventActivations(string(ns)).Create(ea)
	switch {
	case err == nil:
		svc.log.Infof("Created EventActivation: [%s], in namespace: [%s]", appID, ns)
	case apiErrors.IsAlreadyExists(err):
		// We perform update action to adjust OwnerReference of the EventActivation after the backup restore.
		if err = svc.ensureEaUpdate(string(appID), string(ns), si); err != nil {
			return errors.Wrapf(err, "while ensuring update on EventActivation")
		}
		svc.log.Infof("Updated EventActivation: [%s], in namespace: [%s]", appID, ns)
	default:
		return errors.Wrapf(err, "while creating EventActivation with name: %q in namespace: %q", appID, ns)
	}
	return nil
}

func (svc *ProvisionService) ensureEaUpdate(appID, ns string, si *v1beta1.ServiceInstance) error {
	ea, err := svc.eaClient.EventActivations(ns).Get(appID, v1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "while getting EventActivation with name: %q from namespace: %q", appID, ns)
	}
	ea.OwnerReferences = []v1.OwnerReference{
		{
			APIVersion: serviceCatalogAPIVersion,
			Kind:       "ServiceInstance",
			Name:       si.Name,
			UID:        si.UID,
		},
	}
	_, err = svc.eaClient.EventActivations(ns).Update(ea)
	if err != nil {
		return errors.Wrapf(err, "while updating EventActivation with name: %q in namespace: %q", appID, ns)
	}
	return nil
}

// enableDefaultKnativeBroker enables the Knative Eventing default broker for the given namespace
// by adding the proper label to the namespace.
func (svc *ProvisionService) enableDefaultKnativeBroker(ns internal.Namespace) error {
	// get the namespace
	namespace, err := svc.knClient.GetNamespace(string(ns))
	if err != nil {
		return errors.Wrap(err, "namespace not found")
	}

	// check if the namespace has the injection label
	if val, ok := namespace.Labels[knativeEventingInjectionLabelKey]; ok && val == knativeEventingInjectionLabelValueEnabled {
		svc.log.Printf("the default Knative Eventing Broker is already enabled for the namespace: [%s]", namespace.Name)
		return nil
	}

	// add the injection label to the namespace
	if len(namespace.Labels) == 0 {
		namespace.Labels = make(map[string]string, 1)
	}
	namespace.Labels[knativeEventingInjectionLabelKey] = knativeEventingInjectionLabelValueEnabled

	// update the namespace
	_, err = svc.knClient.UpdateNamespace(namespace)
	if err != nil {
		svc.log.Printf("error enabling the default Knative Eventing Broker for namespace: [%v] [%v]", namespace, err)
	}
	return err
}

// persistKnativeSubscription will get a Knative Subscription given application name and namespace and will
// update and persist it. If there is no Knative Subscription found, a new one will be created.
func (svc *ProvisionService) persistKnativeSubscription(applicationName internal.ApplicationName, ns internal.Namespace) error {
	// construct the default broker URI using the given namespace.
	defaultBrokerURI := knative.GetDefaultBrokerURI(ns)

	// get the Knative channel for the application
	channel, err := svc.channelForApp(applicationName)
	if err != nil {
		return errors.Wrapf(err, "getting the Knative channel for the application [%v]", applicationName)
	}

	// subscription selector labels
	labels := map[string]string{
		brokerNamespaceLabelKey: string(ns),
		applicationNameLabelKey: string(applicationName),
	}

	// get Knative subscription by labels
	subscription, err := svc.knClient.GetSubscriptionByLabels(integrationNamespace, labels)
	switch {
	case apiErrors.IsNotFound(err):
		// subscription not found, create a new one
		newSubscription := knative.Subscription(knSubscriptionNamePrefix, integrationNamespace).Spec(channel, defaultBrokerURI).Labels(labels).Build()
		if _, err := svc.knClient.CreateSubscription(newSubscription); err != nil {
			return errors.Wrapf(err, "creating Subscription %s", newSubscription.Name)
		}
		svc.log.Printf("created Knative Subscription: [%v]", newSubscription.Name)
		return nil
	case err != nil:
		return errors.Wrapf(err, "getting Subscription by labels [%v]", labels)
	}

	// update Knative Subscription
	subscription = knative.FromSubscription(subscription).Spec(channel, defaultBrokerURI).Labels(labels).Build()
	subscription, err = svc.knClient.UpdateSubscription(subscription)
	if err != nil {
		return errors.Wrapf(err, "updating existing Knative Subscription with labels [%v] for channel: [%v]", labels, channel.Name)
	}
	svc.log.Printf("Updated Knative Subscription: [%v]", subscription.Name)
	return nil
}

func (svc *ProvisionService) channelForApp(applicationName internal.ApplicationName) (*messagingv1alpha1.Channel, error) {
	labels := map[string]string{
		applicationNameLabelKey: string(applicationName),
	}
	return svc.knClient.GetChannelByLabels(integrationNamespace, labels)
}

// Create a new policy
func (svc *ProvisionService) createIstioPolicy(ns internal.Namespace) error {
	policyName := fmt.Sprintf("%s%s", ns, policyNameSuffix)

	policyLabels := make(map[string]string)
	policyLabels["eventing.knative.dev/broker"] = "default"

	svc.log.Infof("Creating Policy %s in namespace: %s", policyName, string(ns))

	brokerTargetSelector := &istioauthenticationalpha1.TargetSelector{
		Name: brokerTargetSelectorName,
	}
	filterTargetSelector := &istioauthenticationalpha1.TargetSelector{
		Name: filterTargetSelectorName,
	}
	mtls := &istioauthenticationalpha1.MutualTls{
		Mode: istioMtlsPermissiveMode,
	}
	peerAuthenticationMethod := &istioauthenticationalpha1.PeerAuthenticationMethod{
		Params: &istioauthenticationalpha1.PeerAuthenticationMethod_Mtls{
			Mtls: mtls,
		},
	}

	policy := &istiov1alpha1.Policy{
		ObjectMeta: v1.ObjectMeta{
			Name:      policyName,
			Namespace: string(ns),
			Labels:    policyLabels,
		},
		Spec: istioauthenticationalpha1.Policy{
			Targets: []*istioauthenticationalpha1.TargetSelector{brokerTargetSelector, filterTargetSelector},
			Peers:   []*istioauthenticationalpha1.PeerAuthenticationMethod{peerAuthenticationMethod},
		},
	}

	_, err := svc.istioClient.AuthenticationV1alpha1().Policies(string(ns)).Create(policy)
	if err != nil {
		if apiErrors.IsAlreadyExists(err) {
			if _, err := svc.istioClient.AuthenticationV1alpha1().Policies(string(ns)).Update(policy); err != nil {
				return errors.Wrapf(err, "while updating istio policy with name: %q in namespace: %q", policyName, ns)
			}
			return nil
		}
		return errors.Wrapf(err, "while creating istio policy with name: %q in namespace: %q", policyName, ns)
	}
	return nil
}

func getNamespaceFromContext(contextProfile map[string]interface{}) (internal.Namespace, error) {
	return internal.Namespace(contextProfile["namespace"].(string)), nil
}

func getSvcByID(services []internal.Service, id internal.ApplicationServiceID) (internal.Service, error) {
	for _, svc := range services {
		if svc.ID == id {
			return svc, nil
		}
	}
	return internal.Service{}, errors.Errorf("cannot find service with ID [%s]", id)
}

func strPtr(str string) *string {
	return &str
}
