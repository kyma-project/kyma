package broker

import (
	"context"
	"sync"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/internal/knative"
	v1client "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"

	"github.com/pkg/errors"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"
)

// DeprovisionService performs deprovision action
type DeprovisionService struct {
	instStorage         instanceStorage
	instanceStateGetter instanceStateGetter
	operationIDProvider func() (internal.OperationID, error)
	operationInserter   operationInserter
	operationUpdater    operationUpdater
	appSvcFinder        appSvcFinder
	appSvcIDSelector    appSvcIDSelector
	eaClient            v1client.ApplicationconnectorV1alpha1Interface
	knClient            knative.Client
	apiPkgCredsRemover  apiPackageCredentialsRemover

	log       logrus.FieldLogger
	mu        sync.Mutex
	asyncHook func()
}

// NewDeprovisioner creates new Deprovisioner
func NewDeprovisioner(
	instStorage instanceStorage,
	instanceStateGetter instanceStateGetter,
	operationInserter operationInserter,
	operationUpdater operationUpdater,
	opIDProvider func() (internal.OperationID, error),
	appSvcFinder appSvcFinder,
	knClient knative.Client,
	eaClient v1client.ApplicationconnectorV1alpha1Interface,
	log logrus.FieldLogger,
	selector appSvcIDSelector,
	apiPkgCredsRemover apiPackageCredentialsRemover) *DeprovisionService {
	return &DeprovisionService{
		instStorage:         instStorage,
		instanceStateGetter: instanceStateGetter,
		operationInserter:   operationInserter,
		operationUpdater:    operationUpdater,
		operationIDProvider: opIDProvider,
		appSvcFinder:        appSvcFinder,
		knClient:            knClient,
		eaClient:            eaClient,
		appSvcIDSelector:    selector,
		apiPkgCredsRemover:  apiPkgCredsRemover,

		log: log.WithField("service", "deprovisioner"),
	}
}

// Deprovision action
func (svc *DeprovisionService) Deprovision(ctx context.Context, osbCtx osbContext, req *osb.DeprovisionRequest) (*osb.DeprovisionResponse, error) {
	if !req.AcceptsIncomplete {
		return nil, errors.New("asynchronous operation mode required")
	}

	svc.mu.Lock()
	defer svc.mu.Unlock()

	iID := internal.InstanceID(req.InstanceID)

	deprovisioned, err := svc.instanceStateGetter.IsDeprovisioned(iID)
	switch {
	case err != nil:
		return nil, errors.Wrap(err, "while checking if instance is already deprovisioned")
	case deprovisioned:
		return &osb.DeprovisionResponse{Async: false}, nil
	}

	opIDInProgress, inProgress, err := svc.instanceStateGetter.IsDeprovisioningInProgress(iID)
	switch {
	case err != nil:
		return nil, errors.Wrap(err, "while checking if instance is being deprovisioned")
	case inProgress:
		opKeyInProgress := osb.OperationKey(opIDInProgress)
		return &osb.DeprovisionResponse{Async: true, OperationKey: &opKeyInProgress}, nil
	}

	instanceToDeprovision, err := svc.instStorage.Get(iID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting instance %s from storage", iID)
	}

	return svc.doAsyncResourceCleanup(instanceToDeprovision, req)
}

func (svc *DeprovisionService) runningInNamespaceByServiceAndPlanID(instance *internal.Instance) func(i *internal.Instance) bool {
	return func(i *internal.Instance) bool {
		if i.ID == instance.ID { // exclude itself
			return false
		}
		if i.State != internal.InstanceStateSucceeded {
			return false
		}
		if i.ServicePlanID != instance.ServicePlanID {
			return false
		}
		if i.ServiceID != instance.ServiceID {
			return false
		}
		if i.Namespace != instance.Namespace {
			return false
		}
		return true
	}
}

// we are the last, do not remove our self and trigger clean-up
func (svc *DeprovisionService) doAsyncResourceCleanup(instance *internal.Instance, req *osb.DeprovisionRequest) (*osb.DeprovisionResponse, error) {
	appSvcID := svc.appSvcIDSelector.SelectID(req)
	app, err := svc.appSvcFinder.FindOneByServiceID(appSvcID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting application with id %s from storage", appSvcID)
	}

	if app == nil {
		return nil, errors.Wrapf(err, "application with id %s was not found in storage", appSvcID)
	}

	operationID, err := svc.operationIDProvider()
	if err != nil {
		return nil, errors.Wrap(err, "while generating ID for operation")
	}

	op := internal.InstanceOperation{
		InstanceID:  instance.ID,
		OperationID: operationID,
		Type:        internal.OperationTypeRemove,
		State:       internal.OperationStateInProgress,
	}
	if err := svc.operationInserter.Insert(&op); err != nil {
		return nil, errors.Wrap(err, "while inserting instance operation to storage")
	}

	if err := svc.instStorage.UpdateState(instance.ID, internal.InstanceStatePendingDeletion); err != nil {
		return nil, errors.Wrapf(err, "while updating state of the stored instance [%s]", instance.ID)
	}

	go svc.do(instance, operationID, appSvcID, app.Name)

	opKey := osb.OperationKey(operationID)
	return &osb.DeprovisionResponse{
		Async:        true,
		OperationKey: &opKey,
	}, nil

}

func (svc *DeprovisionService) do(instance *internal.Instance, opID internal.OperationID, appSvcID internal.ApplicationServiceID, appName internal.ApplicationName) {
	if svc.asyncHook != nil {
		defer svc.asyncHook()
	}

	iID := instance.ID

	if err := svc.cleanupTheWorld(appSvcID, appName, instance); err != nil {
		svc.log.Errorf(errors.Wrap(err, "while clean up created resources").Error())
		svc.setState(iID, opID, internal.OperationStateFailed, "Cannot clean up created resources.")
		return
	}

	err := svc.instStorage.Remove(iID)
	if err != nil && !IsNotFoundError(err) {
		svc.log.Errorf(errors.Wrap(err, "while removing service instance").Error())
		svc.setState(iID, opID, internal.OperationStateFailed, "Failed to remove instance from storage")
		return
	}

	svc.setState(iID, opID, internal.OperationStateSucceeded, internal.OperationDescriptionDeprovisioningSucceeded)
}

func (svc *DeprovisionService) cleanupAlways(instance *internal.Instance) error {
	svc.log.Infof("Executing clean-up process for resources which should be always deleted")
	if err := svc.apiPkgCredsRemover.EnsureAPIPackageCredentialsDeleted(context.Background(), string(instance.ServiceID), string(instance.ServicePlanID), string(instance.ID)); err != nil {
		return errors.Wrap(err, "while removing API Package credentials")
	}
	return nil
}

func (svc *DeprovisionService) cleanupOnlyIfLast(appSvcID internal.ApplicationServiceID, appName internal.ApplicationName, instance *internal.Instance) error {
	otherInstances, err := svc.instStorage.FindAll(svc.runningInNamespaceByServiceAndPlanID(instance))
	if err != nil {
		return errors.Wrap(err, "while checking if instance this was the last instance for the given plan and service ID")
	}

	noOfOtherInstances := len(otherInstances)
	svc.log.Infof("Found %d additional running instances with the same ServiceID %q and PlanID %q", noOfOtherInstances, instance.ServiceID, instance.ServicePlanID)

	if noOfOtherInstances != 0 {
		svc.log.Infof("Skipping deleting resources which are shared between other instances because we are not the last instance for the given plan and service ID")
		return nil
	}

	svc.log.Infof("Executing clean-up process for resources which should be deleted because this is the last instance of the given plan and service ID [%+v]", instance)
	if err := svc.deprovisionSubscription(appName, instance.Namespace); err != nil {
		return errors.Wrap(err, "while removing Knative Subscription")
	}

	if err := svc.deprovisionEventActivation(appSvcID, instance.Namespace); err != nil {
		return errors.Wrap(err, "while removing Event Activation")
	}
	return nil
}

func (svc *DeprovisionService) cleanupTheWorld(appSvcID internal.ApplicationServiceID, appName internal.ApplicationName, instance *internal.Instance) error {
	if err := svc.cleanupAlways(instance); err != nil {
		return err
	}

	if err := svc.cleanupOnlyIfLast(appSvcID, appName, instance); err != nil {
		return err
	}
	return nil
}

func (svc *DeprovisionService) deprovisionSubscription(appName internal.ApplicationName, ns internal.Namespace) error {
	sub, err := subscriptionForApp(svc.knClient, string(appName), string(ns))
	switch {
	case apierrors.IsNotFound(err):
		// Subscription missing, nothing to delete
		return nil
	case err != nil:
		return errors.Wrap(err, "getting existing Subscription")
	}

	err = svc.knClient.DeleteSubscription(sub)
	if err != nil {
		return errors.Wrap(err, "deleting existing Subscription")
	}
	return nil
}

func (svc *DeprovisionService) setState(iID internal.InstanceID,
	opID internal.OperationID, opState internal.OperationState, desc string) {

	err := svc.operationUpdater.UpdateStateDesc(iID, opID, opState, &desc)
	if err != nil {
		svc.log.Errorf("Cannot update state for instance %s: %s", iID, err)
	}
}

func (svc *DeprovisionService) deprovisionEventActivation(id internal.ApplicationServiceID, namespace internal.Namespace) error {
	err := svc.eaClient.EventActivations(string(namespace)).Delete(string(id), &v1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrap(err, "while deleting the Event Activation")
	}

	// TODO: implement waiting until EventActivation is deleted

	return nil
}

func subscriptionForApp(cli knative.Client, appName, ns string) (*messagingv1alpha1.Subscription, error) {
	labels := map[string]string{
		brokerNamespaceLabelKey: ns,
		applicationNameLabelKey: appName,
	}
	return cli.GetSubscriptionByLabels(integrationNamespace, labels)
}
