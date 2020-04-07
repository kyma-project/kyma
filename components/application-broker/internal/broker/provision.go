package broker

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

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
	integrationNamespace = "kyma-integration"

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
	accessChecker access.ProvisionChecker, appSvcFinder appSvcFinder,
	eaClient v1client.ApplicationconnectorV1alpha1Interface, knClient knative.Client,
	istioClient istioversionedclient.Interface, iStateUpdater instanceStateUpdater,
	operationIDProvider func() (internal.OperationID, error), log logrus.FieldLogger, selector appSvcIDSelector,
	apiPkgCredsCreator apiPackageCredentialsCreator, validateReq func(req *osb.ProvisionRequest) *osb.HTTPStatusCodeError) *ProvisionService {
	return &ProvisionService{
		instanceInserter:         instanceInserter,
		instanceGetter:           instanceGetter,
		instanceStateGetter:      instanceStateGetter,
		instanceStateUpdater:     iStateUpdater,
		operationInserter:        operationInserter,
		operationUpdater:         operationUpdater,
		operationIDProvider:      operationIDProvider,
		accessChecker:            accessChecker,
		appSvcFinder:             appSvcFinder,
		eaClient:                 eaClient,
		knClient:                 knClient,
		istioClient:              istioClient,
		maxWaitTime:              time.Minute,
		appSvcIDSelector:         selector,
		apiPkgCredCreator:        apiPkgCredsCreator,
		validateProvisionRequest: validateReq,
		log:                      log.WithField("service", "provisioner"),
	}
}

// ProvisionService performs provisioning action
type ProvisionService struct {
	instanceInserter         instanceInserter
	instanceGetter           instanceGetter
	instanceStateUpdater     instanceStateUpdater
	operationInserter        operationInserter
	operationUpdater         operationUpdater
	instanceStateGetter      instanceStateGetter
	operationIDProvider      func() (internal.OperationID, error)
	appSvcFinder             appSvcFinder
	eaClient                 v1client.ApplicationconnectorV1alpha1Interface
	accessChecker            access.ProvisionChecker
	knClient                 knative.Client
	istioClient              istioversionedclient.Interface
	appSvcIDSelector         appSvcIDSelector
	apiPkgCredCreator        apiPackageCredentialsCreator
	validateProvisionRequest func(req *osb.ProvisionRequest) *osb.HTTPStatusCodeError

	mu sync.Mutex

	maxWaitTime time.Duration
	log         logrus.FieldLogger
	asyncHook   func()
}

// Provision action
func (svc *ProvisionService) Provision(ctx context.Context, osbCtx osbContext, req *osb.ProvisionRequest) (*osb.ProvisionResponse, *osb.HTTPStatusCodeError) {
	if err := svc.validateProvisionRequest(req); err != nil {
		return nil, err
	}

	svc.mu.Lock()
	defer svc.mu.Unlock()

	var (
		iID       = internal.InstanceID(req.InstanceID)
		namespace = internal.Namespace(osbCtx.BrokerNamespace)
	)

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

	opID, err := svc.operationIDProvider()
	if err != nil {
		return nil, &osb.HTTPStatusCodeError{StatusCode: http.StatusInternalServerError, ErrorMessage: strPtr(fmt.Sprintf("while generating ID for operation: %v", err))}
	}

	op := internal.InstanceOperation{
		InstanceID:  iID,
		OperationID: opID,
		Type:        internal.OperationTypeCreate,
		State:       internal.OperationStateInProgress,
	}

	if err := svc.operationInserter.Insert(&op); err != nil {
		return nil, &osb.HTTPStatusCodeError{StatusCode: http.StatusInternalServerError, ErrorMessage: strPtr(fmt.Sprintf("while inserting instance operation to storage: %v", err))}
	}

	appSvcID := svc.appSvcIDSelector.SelectID(req)
	app, err := svc.appSvcFinder.FindOneByServiceID(appSvcID)
	if err != nil {
		return nil, &osb.HTTPStatusCodeError{StatusCode: http.StatusInternalServerError, ErrorMessage: strPtr(fmt.Sprintf("while getting application with id: %s to storage: %v", appSvcID, err))}
	}

	service, err := getSvcByID(app.Services, appSvcID)
	if err != nil {
		return nil, &osb.HTTPStatusCodeError{StatusCode: http.StatusBadRequest, ErrorMessage: strPtr(fmt.Sprintf("while getting service [%s] from Application [%s]: %v", appSvcID, app.Name, err))}
	}

	i := internal.Instance{
		ID:            iID,
		Namespace:     namespace,
		ServiceID:     internal.ServiceID(req.ServiceID),
		ServicePlanID: internal.ServicePlanID(req.PlanID),
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

	go svc.do(req.Parameters, iID, opID, app.Name, app.CompassMetadata.ApplicationID, appSvcID, namespace, service.EventProvider, service.IsBindable(), service.DisplayName)

	return resp, nil
}

func (svc *ProvisionService) do(inputParams map[string]interface{}, iID internal.InstanceID, opID internal.OperationID, appName internal.ApplicationName, appID string, appSvcID internal.ApplicationServiceID, ns internal.Namespace, eventProvider, apiProvider bool, displayName string) {
	if svc.asyncHook != nil {
		defer svc.asyncHook()
	}
	canProvisionOutput, err := svc.accessChecker.CanProvision(iID, appSvcID, ns, svc.maxWaitTime)
	svc.log.Infof("Access checker: canProvisionInstance(appName=[%s], appSvcID=[%s], ns=[%s]) returned: canProvisionOutput=[%+v], error=[%v]", appName, appSvcID, ns, canProvisionOutput, err)
	if err != nil {
		opDesc := fmt.Sprintf("provisioning failed on error: %s", err)
		svc.updateStateFailed(iID, opID, opDesc)
		return
	}

	if !canProvisionOutput.Allowed {
		opDesc := fmt.Sprintf("Forbidden provisioning instance [%s] for application [name: %s, id: %s] in namespace: [%s]. Reason: [%s]", iID, appName, appSvcID, ns, canProvisionOutput.Reason)
		svc.updateStateFailed(iID, opID, opDesc)
		return
	}

	if apiProvider {
		svc.log.Infof("Ensuring that APIPackage credentials are available [appID: %q, appSvcID: %q, instanceID: %q, inputParams: %v]", appID, appSvcID, iID, inputParams)
		if err := svc.apiPkgCredCreator.EnsureAPIPackageCredentials(context.Background(), appID, string(appSvcID), string(iID), inputParams); err != nil {
			opDesc := fmt.Sprintf("provisioning failed while ensuring API Package credentials: %s", err)
			svc.updateStateFailed(iID, opID, opDesc)
			return
		}
		svc.log.Infof("Created APIPackage credentials successfully [appID: %q, appSvcID: %q, instanceID: %q, inputParams: %v]", appID, appSvcID, iID, inputParams)
	}

	if !eventProvider {
		svc.updateStateSuccess(iID, opID)
		return
	}

	// create Kyma EventActivation
	if err := svc.createEaOnSuccessProvision(appName, appSvcID, ns, displayName); err != nil {
		opDesc := fmt.Sprintf("provisioning failed while creating EventActivation on error: %s", err)
		svc.updateStateFailed(iID, opID, opDesc)
		return
	}

	// persist Knative Subscription
	if err := svc.persistKnativeSubscription(appName, ns); err != nil {
		svc.log.Printf("Error persisting Knative Subscription: %v", err)
		opDesc := fmt.Sprintf("provisioning failed while persisting Knative Subscription for application: %s namespace: %s on error: %s", appName, ns, err)
		svc.updateStateFailed(iID, opID, opDesc)
		return
	}

	// enable the namespace default Knative Broker
	if err := svc.enableDefaultKnativeBroker(ns); err != nil {
		opDesc := fmt.Sprintf("provisioning failed while enabling default Knative Broker for namespace: %s"+
			" on error: %s", ns, err)
		svc.updateStateFailed(iID, opID, opDesc)
		return
	}

	// Create istio policy
	if err := svc.createIstioPolicy(ns); err != nil {
		svc.log.Errorf("Error creating istio policy: %v", err)
		opDesc := fmt.Sprintf("provisioning failed while creating an istio policy for application: %s"+
			" namespace: %s on error: %s", appName, ns, err)
		svc.updateStateFailed(iID, opID, opDesc)
		return
	}

	svc.updateStateSuccess(iID, opID)
}

func (svc *ProvisionService) updateStateSuccess(iID internal.InstanceID, opID internal.OperationID) {
	svc.updateStates(iID, opID, internal.InstanceStateSucceeded, internal.OperationStateSucceeded, internal.OperationDescriptionProvisioningSucceeded)
}

func (svc *ProvisionService) updateStateFailed(iID internal.InstanceID, opID internal.OperationID, opDesc string) {
	svc.updateStates(iID, opID, internal.InstanceStateFailed, internal.OperationStateFailed, opDesc)
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
	appID internal.ApplicationServiceID, ns internal.Namespace, displayName string) error {

	ea := &v1alpha1.EventActivation{
		TypeMeta: v1.TypeMeta{
			Kind:       "EventActivation",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      string(appID),
			Namespace: string(ns),
		},
		Spec: v1alpha1.EventActivationSpec{
			DisplayName: displayName,
			SourceID:    string(appName),
		},
	}
	_, err := svc.eaClient.EventActivations(string(ns)).Create(ea)
	switch {
	case err == nil:
		svc.log.Infof("Created EventActivation: [%s], in namespace: [%s]", appID, ns)
	case apiErrors.IsAlreadyExists(err):
		if err = svc.ensureEaUpdate(string(appID), string(ns), ea); err != nil {
			return errors.Wrapf(err, "while ensuring update on EventActivation")
		}
		svc.log.Infof("Updated EventActivation: [%s], in namespace: [%s]", appID, ns)
	default:
		return errors.Wrapf(err, "while creating EventActivation with name: %q in namespace: %q", appID, ns)
	}
	return nil
}

func (svc *ProvisionService) ensureEaUpdate(appID, ns string, newEA *v1alpha1.EventActivation) error {
	oldEA, err := svc.eaClient.EventActivations(ns).Get(appID, v1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "while getting EventActivation with name: %q from namespace: %q", appID, ns)
	}
	toUpdate := oldEA.DeepCopy()
	oldEA.Spec = newEA.Spec
	_, err = svc.eaClient.EventActivations(ns).Update(toUpdate)
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
		createdSub, err := svc.knClient.CreateSubscription(newSubscription)
		if err != nil {
			return errors.Wrapf(err, "while creating Knative Subscription")
		}
		svc.log.Printf("Created Knative Subscription: [%v]", createdSub.Name)
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

	objInfo := fmt.Sprintf("Istio Policy: [%s], in namespace: [%s]", policy.Name, policy.Namespace)
	_, err := svc.istioClient.AuthenticationV1alpha1().Policies(string(ns)).Create(policy)
	switch {
	case err == nil:
		svc.log.Infof("Created %s", objInfo)
	case apiErrors.IsAlreadyExists(err):
		if err = svc.ensureIstioPolicy(policy); err != nil {
			return errors.Wrapf(err, "while ensuring update on %s", objInfo)
		}
		svc.log.Infof("Updated %s", objInfo)
	default:
		return errors.Wrapf(err, "while creating %s", objInfo)
	}
	return nil
}

func (svc *ProvisionService) ensureIstioPolicy(newPolicy *istiov1alpha1.Policy) error {
	oldPolicy, err := svc.istioClient.AuthenticationV1alpha1().Policies(newPolicy.Namespace).Get(newPolicy.Name, v1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "while getting istio policy with name: %q in namespace: %q", newPolicy.Name, newPolicy.Namespace)
	}

	toUpdate := oldPolicy.DeepCopy()
	toUpdate.Labels = newPolicy.Labels
	toUpdate.Spec = newPolicy.Spec

	if _, err := svc.istioClient.AuthenticationV1alpha1().Policies(toUpdate.Namespace).Update(toUpdate); err != nil {
		return errors.Wrapf(err, "while updating istio policy with name: %q in namespace: %q", toUpdate.Name, newPolicy.Namespace)
	}
	return nil
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

func validateProvisionRequestV2(req *osb.ProvisionRequest) *osb.HTTPStatusCodeError {
	if !req.AcceptsIncomplete {
		return &osb.HTTPStatusCodeError{StatusCode: http.StatusBadRequest, ErrorMessage: strPtr("asynchronous operation mode required")}
	}

	return nil
}

// Deprecated, remove in https://github.com/kyma-project/kyma/issues/7415
func validateProvisionRequestV1(req *osb.ProvisionRequest) *osb.HTTPStatusCodeError {
	if len(req.Parameters) > 0 {
		return &osb.HTTPStatusCodeError{StatusCode: http.StatusBadRequest, ErrorMessage: strPtr("application-broker does not support configuration options for provisioning")}
	}
	if !req.AcceptsIncomplete {
		return &osb.HTTPStatusCodeError{StatusCode: http.StatusBadRequest, ErrorMessage: strPtr("asynchronous operation mode required")}
	}

	return nil
}
