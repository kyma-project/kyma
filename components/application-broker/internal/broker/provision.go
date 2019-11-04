package broker

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/komkom/go-jsonhash"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/pkg/errors"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/sirupsen/logrus"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/internal/access"
	"github.com/kyma-project/kyma/components/application-broker/internal/knative"
	"github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	v1client "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
)

const (
	integrationNamespace     = "kyma-integration"
	serviceCatalogAPIVersion = "servicecatalog.k8s.io/v1beta1"

	// label used for enabling Knative eventing default broker for a given namespace
	knativeEventingInjectionLabelKey          = "knative-eventing-injection"
	knativeEventingInjectionLabelValueEnabled = "enabled"

	// labels used for selecting Knative Channels and Subscriptions
	applicationNameLabelKey = "applicationName"
	brokerNamespaceLabelKey = "brokerNamespace"

	// knSubscriptionNamePrefix is the prefix used for the generated Knative Subscription name
	knSubscriptionNamePrefix = "brokersub"
)

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

	mu sync.Mutex

	maxWaitTime time.Duration
	log         logrus.FieldLogger
	asyncHook   func()
}

// NewProvisioner creates provisioner
func NewProvisioner(instanceInserter instanceInserter,
	instanceGetter instanceGetter,
	instanceStateGetter instanceStateGetter,
	operationInserter operationInserter,
	operationUpdater operationUpdater,
	accessChecker access.ProvisionChecker,
	appSvcFinder appSvcFinder,
	serviceInstanceGetter serviceInstanceGetter,
	eaClient v1client.ApplicationconnectorV1alpha1Interface,
	knClient knative.Client,
	iStateUpdater instanceStateUpdater,
	operationIDProvider func() (internal.OperationID, error),
	log logrus.FieldLogger) *ProvisionService {

	return &ProvisionService{
		instanceInserter:     instanceInserter,
		instanceGetter:       instanceGetter,
		instanceStateGetter:  instanceStateGetter,
		instanceStateUpdater: iStateUpdater,
		operationInserter:    operationInserter,
		operationUpdater:     operationUpdater,
		operationIDProvider:  operationIDProvider,
		accessChecker:        accessChecker,
		appSvcFinder:         appSvcFinder,

		eaClient: eaClient,
		knClient: knClient,

		serviceInstanceGetter: serviceInstanceGetter,
		maxWaitTime:           time.Minute,
		log:                   log.WithField("service", "provisioner"),
	}
}

// Provision action
func (svc *ProvisionService) Provision(ctx context.Context, osbCtx osbContext, req *osb.ProvisionRequest) (*osb.ProvisionResponse, *osb.HTTPStatusCodeError) {
	if !req.AcceptsIncomplete {
		return nil, &osb.HTTPStatusCodeError{StatusCode: http.StatusBadRequest, ErrorMessage: strPtr("asynchronous operation mode required")}
	}

	svc.mu.Lock()
	defer svc.mu.Unlock()

	iID := internal.InstanceID(req.InstanceID)
	paramHash := jsonhash.HashS(req.Parameters)

	switch state, err := svc.instanceStateGetter.IsProvisioned(iID); true {
	case err != nil:
		return nil, &osb.HTTPStatusCodeError{StatusCode: http.StatusBadRequest, ErrorMessage: strPtr(fmt.Sprintf("while checking if instance is already provisioned: %v", err))}
	case state:
		if err := svc.compareProvisioningParameters(iID, paramHash); err != nil {
			return nil, &osb.HTTPStatusCodeError{StatusCode: http.StatusConflict, ErrorMessage: strPtr(fmt.Sprintf("while comparing provisioning parameters %v: %v", req.Parameters, err))}
		}
		return &osb.ProvisionResponse{Async: false}, nil
	}

	switch opIDInProgress, inProgress, err := svc.instanceStateGetter.IsProvisioningInProgress(iID); true {
	case err != nil:
		return nil, &osb.HTTPStatusCodeError{StatusCode: http.StatusBadRequest, ErrorMessage: strPtr(fmt.Sprintf("while checking if instance is being provisioned: %v", err))}
	case inProgress:
		if err := svc.compareProvisioningParameters(iID, paramHash); err != nil {
			return nil, &osb.HTTPStatusCodeError{StatusCode: http.StatusConflict, ErrorMessage: strPtr(fmt.Sprintf("while comparing provisioning parameters %v: %v", req.Parameters, err))}
		}
		opKeyInProgress := osb.OperationKey(opIDInProgress)
		return &osb.ProvisionResponse{Async: true, OperationKey: &opKeyInProgress}, nil
	}

	id, err := svc.operationIDProvider()
	if err != nil {
		return nil, &osb.HTTPStatusCodeError{StatusCode: http.StatusBadRequest, ErrorMessage: strPtr(fmt.Sprintf("while generating ID for operation: %v", err))}
	}
	opID := internal.OperationID(id)

	op := internal.InstanceOperation{
		InstanceID:  iID,
		OperationID: opID,
		Type:        internal.OperationTypeCreate,
		State:       internal.OperationStateInProgress,
		ParamsHash:  paramHash,
	}

	if err := svc.operationInserter.Insert(&op); err != nil {
		return nil, &osb.HTTPStatusCodeError{StatusCode: http.StatusBadRequest, ErrorMessage: strPtr(fmt.Sprintf("while inserting instance operation to storage: %v", err))}
	}

	svcID := internal.ServiceID(req.ServiceID)
	svcPlanID := internal.ServicePlanID(req.PlanID)

	app, err := svc.appSvcFinder.FindOneByServiceID(internal.ApplicationServiceID(req.ServiceID))
	if err != nil {
		return nil, &osb.HTTPStatusCodeError{StatusCode: http.StatusBadRequest, ErrorMessage: strPtr(fmt.Sprintf("while getting application with id: %s to storage: %v", req.ServiceID, err))}
	}

	namespace, err := getNamespaceFromContext(req.Context)
	if err != nil {
		return nil, &osb.HTTPStatusCodeError{StatusCode: http.StatusBadRequest, ErrorMessage: strPtr(fmt.Sprintf("while getting namespace from context %v", err))}
	}

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
		ParamsHash:    paramHash,
	}

	if err = svc.instanceInserter.Insert(&i); err != nil {
		return nil, &osb.HTTPStatusCodeError{StatusCode: http.StatusBadRequest, ErrorMessage: strPtr(fmt.Sprintf("while inserting instance to storage: %v", err))}
	}

	opKey := osb.OperationKey(op.OperationID)
	resp := &osb.ProvisionResponse{
		OperationKey: &opKey,
		Async:        true,
	}

	svc.doAsync(iID, opID, app.Name, getApplicationServiceID(req), namespace, service.EventProvider, service.DisplayName)
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
	canProvisionOutput, err := svc.accessChecker.CanProvision(iID, appID, ns, svc.maxWaitTime)
	svc.log.Infof("Access checker: canProvisionInstance(appName=[%s], appID=[%s], ns=[%s]) returned: canProvisionOutput=[%+v], error=[%v]", appName, appID, ns, canProvisionOutput, err)

	var instanceState internal.InstanceState
	var opState internal.OperationState
	var opDesc string

	if err != nil {
		instanceState = internal.InstanceStateFailed
		opState = internal.OperationStateFailed
		opDesc = fmt.Sprintf("provisioning failed on error: %s", err)
	} else if !canProvisionOutput.Allowed {
		instanceState = internal.InstanceStateFailed
		opState = internal.OperationStateFailed
		opDesc = fmt.Sprintf("Forbidden provisioning instance [%s] for application [name: %s, id: %s] in namespace: [%s]. Reason: [%s]", iID, appName, appID, ns, canProvisionOutput.Reason)
	} else {
		instanceState = internal.InstanceStateSucceeded
		opState = internal.OperationStateSucceeded
		opDesc = "provisioning succeeded"
		if eventProvider {
			// prepare selectors
			namespace := string(ns)
			applicationID := string(appID)
			applicationName := string(appName)

			// enable the Eventing flow
			if err = svc.enableDefaultKnativeEventingBrokerOnSuccessProvision(namespace); err != nil {
				instanceState = internal.InstanceStateFailed
				opState = internal.OperationStateFailed
				opDesc = fmt.Sprintf("provisioning failed while enabling the default Knative Eventing Broker for namespace: %s on error: %s", namespace, err)
			} else if err = svc.persistKnativeSubscriptionOnSuccessProvision(applicationName, namespace); err != nil {
				svc.log.Printf("Error persisting Knative Subscription: %v", err)
				instanceState = internal.InstanceStateFailed
				opState = internal.OperationStateFailed
				opDesc = fmt.Sprintf("provisioning failed while persisting the Knative Subscription for application: %s and namespace: %s on error: %s", applicationName, namespace, err)
			} else if err = svc.createEaOnSuccessProvision(applicationName, applicationID, namespace, displayName, iID); err != nil {
				instanceState = internal.InstanceStateFailed
				opState = internal.OperationStateFailed
				opDesc = fmt.Sprintf("provisioning failed while creating EventActivation on error: %s", err)
			}
		}
	}

	if err := svc.instanceStateUpdater.UpdateState(iID, instanceState); err != nil {
		svc.log.Errorf("Cannot update state of the stored instance [%s]: [%v]\n", iID, err)
	}

	if err := svc.operationUpdater.UpdateStateDesc(iID, opID, opState, &opDesc); err != nil {
		svc.log.Errorf("Cannot update state for ServiceInstance [%s]: [%v]\n", iID, err)
		return
	}
}

func (svc *ProvisionService) createEaOnSuccessProvision(appName, appID, ns string, displayName string, iID internal.InstanceID) error {
	// instance ID is the serviceInstance.Spec.ExternalID
	si, err := svc.serviceInstanceGetter.GetByNamespaceAndExternalID(ns, string(iID))
	if err != nil {
		return errors.Wrapf(err, "while getting service instance with external id: %q in namespace: %q", iID, ns)
	}
	ea := &v1alpha1.EventActivation{
		TypeMeta: metav1.TypeMeta{
			Kind:       "EventActivation",
			APIVersion: "applicationconnector.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      appID,
			Namespace: ns,
			OwnerReferences: []metav1.OwnerReference{
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
			SourceID:    appName,
		},
	}
	_, err = svc.eaClient.EventActivations(ns).Create(ea)
	switch {
	case err == nil:
		svc.log.Infof("Created EventActivation: [%s], in namespace: [%s]", appID, ns)
	case apierrors.IsAlreadyExists(err):
		// We perform update action to adjust OwnerReference of the EventActivation after the backup restore.
		if err = svc.ensureEaUpdate(appID, ns, si); err != nil {
			return errors.Wrapf(err, "while ensuring update on EventActivation")
		}
		svc.log.Infof("Updated EventActivation: [%s], in namespace: [%s]", appID, ns)
	default:
		return errors.Wrapf(err, "while creating EventActivation with name: %q in namespace: %q", appID, ns)
	}
	return nil
}

func (svc *ProvisionService) ensureEaUpdate(appID, ns string, si *v1beta1.ServiceInstance) error {
	ea, err := svc.eaClient.EventActivations(ns).Get(appID, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "while getting EventActivation with name: %q from namespace: %q", appID, ns)
	}
	ea.OwnerReferences = []metav1.OwnerReference{
		{
			APIVersion: serviceCatalogAPIVersion,
			Kind:       "ServiceInstance",
			Name:       si.Name,
			UID:        si.UID,
		},
	}
	ea, err = svc.eaClient.EventActivations(ns).Update(ea)
	if err != nil {
		return errors.Wrapf(err, "while updating EventActivation with name: %q in namespace: %q", appID, ns)
	}
	return nil
}

func (svc *ProvisionService) compareProvisioningParameters(iID internal.InstanceID, newHash string) error {
	instance, err := svc.instanceGetter.Get(iID)
	switch {
	case err == nil:
	case IsNotFoundError(err):
		return nil
	default:
		return errors.Wrapf(err, "while getting instance %s from storage", iID)
	}

	if instance.ParamsHash != newHash {
		return errors.Errorf("provisioning parameters hash differs - new %s, old %s, for instance %s", newHash, instance.ParamsHash, iID)
	}

	return nil
}

// enableDefaultKnativeEventingBroker enables the Knative Eventing default broker for the given namespace
// by adding the proper label to the namespace.
func (svc *ProvisionService) enableDefaultKnativeEventingBrokerOnSuccessProvision(ns string) error {
	// get the namespace and return error if it does not exist
	namespace, err := svc.knClient.GetNamespace(ns)
	if err != nil {
		svc.log.Printf("error getting namespace: [%s] [%v]", ns, err)
		return err
	}

	// check if the namespace already has the knative-eventing-injection label and it is set to true
	if val, ok := namespace.Labels[knativeEventingInjectionLabelKey]; ok && val == knativeEventingInjectionLabelValueEnabled {
		svc.log.Printf("the default Knative Eventing Broker is already enabled for the namespace: [%s]", namespace.Name)
		return nil
	}

	// add the knative-eventing-injection label to the namespace
	namespace.Labels[knativeEventingInjectionLabelKey] = knativeEventingInjectionLabelValueEnabled

	// update the namespace
	_, err = svc.knClient.UpdateNamespace(namespace)
	if err != nil {
		svc.log.Printf("error enabling the default Knative Eventing Broker for namespace: [%v] [%v]", namespace, err)
	}
	return err
}

// persistKnativeSubscriptionOnSuccessProvision will get a Knative Subscription based on the given application name and namespace
// and will update and persist it. If there is no Knative Subscription for this application name and namespace existed before,
// a new one will be created.
func (svc *ProvisionService) persistKnativeSubscriptionOnSuccessProvision(applicationName, ns string) error {
	// prepare the selector labels
	labels := map[string]string{
		brokerNamespaceLabelKey: ns,
		applicationNameLabelKey: applicationName,
	}

	// construct the default broker URI using the given namespace.
	defaultBrokerURI := knative.GetDefaultBrokerURI(ns)

	// get the Knative channel by labels
	channel, err := svc.knClient.GetChannelByLabels(integrationNamespace, labels)
	if err != nil {
		return errors.Wrapf(err, "getting Channel by labels %v", labels)
	}

	// try to get the knative subscription in case it was created before by a previous provisioning request
	currentSubscription, err := svc.knClient.GetSubscriptionByLabels(integrationNamespace, labels)
	switch {
	// the subscription does not exist before, create a new one
	case apierrors.IsNotFound(err):
		// create the Knative subscription
		desiredSubscription := knative.Subscription(knSubscriptionNamePrefix, integrationNamespace).Spec(channel, defaultBrokerURI).Labels(labels).Build()
		if _, err := svc.knClient.CreateSubscription(desiredSubscription); err != nil {
			return errors.Wrapf(err, "creating Subscription %s", desiredSubscription.Name)
		}
		svc.log.Printf("Created Knative Subscription: [%v]", desiredSubscription.Name)
		return nil

	case err != nil:
		return errors.Wrapf(err, "getting Subscription by labels %v", labels)
	}

	// update the current Knative Subscription
	currentSubscription = knative.FromSubscription(currentSubscription).Spec(channel, defaultBrokerURI).Labels(labels).Build()
	currentSubscription, err = svc.knClient.UpdateSubscription(currentSubscription)
	if err != nil {
		return errors.Wrapf(err, "updating existing Knative Subscription %v", currentSubscription.Name)
	}
	svc.log.Printf("Updated Knative Subscription: [%v]", currentSubscription.Name)
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
