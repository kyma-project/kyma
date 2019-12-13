package broker

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/pkg/errors"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/sirupsen/logrus"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/internal/knative"
)

// DeprovisionService performs deprovision action
type DeprovisionService struct {
	instStorage         instanceStorage
	instanceStateGetter instanceStateGetter
	operationIDProvider func() (internal.OperationID, error)
	operationInserter   operationInserter
	operationUpdater    operationUpdater
	appSvcFinder        appSvcFinder

	knClient knative.Client

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
	log logrus.FieldLogger) *DeprovisionService {

	return &DeprovisionService{
		instStorage:         instStorage,
		instanceStateGetter: instanceStateGetter,
		operationInserter:   operationInserter,
		operationUpdater:    operationUpdater,
		operationIDProvider: opIDProvider,
		appSvcFinder:        appSvcFinder,
		knClient:            knClient,
		log:                 log.WithField("service", "deprovisioner"),
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
	case IsNotFoundError(err):
		return nil, err
	case err != nil:
		return nil, errors.Wrap(err, "checking if instance is already deprovisioned")
	case deprovisioned:
		return &osb.DeprovisionResponse{Async: false}, nil
	}

	opIDInProgress, inProgress, err := svc.instanceStateGetter.IsDeprovisioningInProgress(iID)
	switch {
	case IsNotFoundError(err):
		return nil, err
	case err != nil:
		return nil, errors.Wrap(err, "checking if instance is being deprovisioned")
	case inProgress:
		opKeyInProgress := osb.OperationKey(opIDInProgress)
		return &osb.DeprovisionResponse{Async: true, OperationKey: &opKeyInProgress}, nil
	}

	operationID, err := svc.operationIDProvider()
	if err != nil {
		return nil, errors.Wrap(err, "generating operation ID")
	}

	iS, err := svc.instStorage.Get(iID)
	if err != nil {
		return nil, errors.Wrapf(err, "getting instance %s from storage", iID)
	}

	app, err := svc.appSvcFinder.FindOneByServiceID(internal.ApplicationServiceID(req.ServiceID))
	if err != nil {
		return nil, &osb.HTTPStatusCodeError{
			StatusCode:   http.StatusBadRequest,
			ErrorMessage: strPtr(fmt.Sprintf("getting application with id %s from storage: %v", req.ServiceID, err)),
		}
	}

	op := internal.InstanceOperation{
		InstanceID:  iID,
		OperationID: operationID,
		Type:        internal.OperationTypeRemove,
		State:       internal.OperationStateInProgress,
	}

	if err := svc.operationInserter.Insert(&op); err != nil {
		return nil, errors.Wrap(err, "inserting instance operation to storage")
	}

	err = svc.instStorage.Remove(iID)
	switch {
	case IsNotFoundError(err):
		return nil, err
	case err != nil:
		return nil, errors.Wrap(err, "removing instance from storage")
	}

	opKey := osb.OperationKey(operationID)
	resp := &osb.DeprovisionResponse{
		Async:        true,
		OperationKey: &opKey,
	}

	svc.doAsync(iID, operationID, app.Name, iS.Namespace)
	return resp, nil
}

func (svc *DeprovisionService) doAsync(iID internal.InstanceID,
	opID internal.OperationID, appName internal.ApplicationName, ns internal.Namespace) {

	go svc.do(iID, opID, appName, ns)
}

func (svc *DeprovisionService) do(iID internal.InstanceID,
	opID internal.OperationID, appName internal.ApplicationName, ns internal.Namespace) {

	if svc.asyncHook != nil {
		defer svc.asyncHook()
	}

	err := svc.deprovisionSubscription(appName, ns)
	if err != nil {
		svc.log.Printf("Failed to deprovision Subscription: %s", err)
		svc.setState(iID, opID, internal.OperationStateFailed, "failed to deprovision Subscription")
		return
	}

	svc.setState(iID, opID, internal.OperationStateSucceeded, internal.OperationDescriptionDeprovisioningSucceeded)
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

func subscriptionForApp(cli knative.Client, appName, ns string) (*messagingv1alpha1.Subscription, error) {
	labels := map[string]string{
		brokerNamespaceLabelKey: ns,
		applicationNameLabelKey: appName,
	}
	return cli.GetSubscriptionByLabels(integrationNamespace, labels)
}
