package broker

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	osb "github.com/kubernetes-sigs/go-open-service-broker-client/v2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	securityv1beta1 "istio.io/client-go/pkg/apis/security/v1beta1"
	securityclientv1beta1 "istio.io/client-go/pkg/clientset/versioned/typed/security/v1beta1"

	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/internal/access"
	"github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	v1client "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
)

type RestoreProvisionRequest struct {
	Parameters           map[string]interface{}
	InstanceID           internal.InstanceID
	OperationID          internal.OperationID
	Namespace            internal.Namespace
	ApplicationServiceID internal.ApplicationServiceID
}

// NewProvisioner creates provisioner
func NewProvisioner(instanceInserter instanceInserter, instanceStateGetter instanceStateGetter,
	operationInserter operationInserter, operationUpdater operationUpdater,
	accessChecker access.ProvisionChecker, appSvcFinder appSvcFinder,
	eaClient v1client.ApplicationconnectorV1alpha1Interface,
	istioClient securityclientv1beta1.SecurityV1beta1Interface, iStateUpdater instanceStateUpdater,
	operationIDProvider func() (internal.OperationID, error), log logrus.FieldLogger, selector appSvcIDSelector,
	apiPkgCredsCreator apiPackageCredentialsCreator,
	validateReq func(req *osb.ProvisionRequest) *osb.HTTPStatusCodeError, newEventingFlow bool) *ProvisionService {
	return &ProvisionService{
		instanceInserter:         instanceInserter,
		instanceStateGetter:      instanceStateGetter,
		instanceStateUpdater:     iStateUpdater,
		operationInserter:        operationInserter,
		operationUpdater:         operationUpdater,
		operationIDProvider:      operationIDProvider,
		accessChecker:            accessChecker,
		appSvcFinder:             appSvcFinder,
		eaClient:                 eaClient,
		istioClient:              istioClient,
		maxWaitTime:              time.Minute,
		appSvcIDSelector:         selector,
		apiPkgCredCreator:        apiPkgCredsCreator,
		validateProvisionRequest: validateReq,
		log:                      log.WithField("service", "provisioner"),
		newEventingFlow:          newEventingFlow,
	}
}

// ProvisionService performs provisioning action
type ProvisionService struct {
	instanceInserter         instanceInserter
	instanceStateUpdater     instanceStateUpdater
	operationInserter        operationInserter
	operationUpdater         operationUpdater
	instanceStateGetter      instanceStateGetter
	operationIDProvider      func() (internal.OperationID, error)
	appSvcFinder             appSvcFinder
	eaClient                 v1client.ApplicationconnectorV1alpha1Interface
	accessChecker            access.ProvisionChecker
	istioClient              securityclientv1beta1.PeerAuthenticationsGetter
	appSvcIDSelector         appSvcIDSelector
	apiPkgCredCreator        apiPackageCredentialsCreator
	validateProvisionRequest func(req *osb.ProvisionRequest) *osb.HTTPStatusCodeError

	mu sync.Mutex

	maxWaitTime time.Duration
	log         logrus.FieldLogger
	asyncHook   func()

	newEventingFlow bool
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

// ProvisionReprocess triggers provision process for other than broker (http) calls
func (svc *ProvisionService) ProvisionReprocess(req RestoreProvisionRequest) error {
	var app *internal.Application
	err := wait.PollImmediate(500*time.Millisecond, 10*time.Second, func() (done bool, err error) {
		app, err = svc.appSvcFinder.FindOneByServiceID(req.ApplicationServiceID)
		if err != nil {
			svc.log.Warnf("cannot find application based on service ID %s: %s", req.ApplicationServiceID, err)
			return false, nil
		}
		if app == nil {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return errors.Wrapf(err, "while waiting for application with ID: %s", req.ApplicationServiceID)
	}

	service, err := getSvcByID(app.Services, req.ApplicationServiceID)
	if err != nil {
		return errors.Wrap(err, "while getting service")
	}

	go svc.do(req.Parameters, req.InstanceID, req.OperationID, app.Name, app.CompassMetadata.ApplicationID, req.ApplicationServiceID, req.Namespace, service.EventProvider, service.IsBindable(), service.DisplayName)

	return nil
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

func (svc *ProvisionService) ensurePeerAuthentication(newPeerAuth *securityv1beta1.PeerAuthentication) error {
	oldPeerAuth, err := svc.istioClient.PeerAuthentications(newPeerAuth.Namespace).Get(newPeerAuth.Name, v1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "while getting Istio PeerAuthentication with name: %q in namespace: %q", newPeerAuth.Name, newPeerAuth.Namespace)
	}

	toUpdate := oldPeerAuth.DeepCopy()
	toUpdate.Labels = newPeerAuth.Labels
	toUpdate.Spec = newPeerAuth.Spec

	if _, err := svc.istioClient.PeerAuthentications(toUpdate.Namespace).Update(toUpdate); err != nil {
		return errors.Wrapf(err, "while updating Istio PeerAuthentication with name: %q in namespace: %q", toUpdate.Name, newPeerAuth.Namespace)
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
