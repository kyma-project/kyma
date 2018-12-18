package broker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/internal/access"
	"github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	v1client "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	"github.com/pkg/errors"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/sirupsen/logrus"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

const serviceCatalogAPIVersion = "servicecatalog.k8s.io/v1beta1"

// NewProvisioner creates provisioner
func NewProvisioner(instanceInserter instanceInserter, instanceStateGetter instanceStateGetter, operationInserter operationInserter, operationUpdater operationUpdater, accessChecker access.ProvisionChecker, reSvcFinder reSvcFinder, serviceInstanceGetter serviceInstanceGetter, reClient v1client.ApplicationconnectorV1alpha1Interface, iStateUpdater instanceStateUpdater,
	operationIDProvider func() (internal.OperationID, error), log logrus.FieldLogger) *ProvisionService {
	return &ProvisionService{
		instanceInserter:      instanceInserter,
		instanceStateGetter:   instanceStateGetter,
		instanceStateUpdater:  iStateUpdater,
		operationInserter:     operationInserter,
		operationUpdater:      operationUpdater,
		operationIDProvider:   operationIDProvider,
		accessChecker:         accessChecker,
		reSvcFinder:           reSvcFinder,
		reClient:              reClient,
		serviceInstanceGetter: serviceInstanceGetter,
		maxWaitTime:           time.Minute,
		log:                   log.WithField("service", "provisioner"),
	}
}

// ProvisionService performs provisioning action
type ProvisionService struct {
	instanceInserter      instanceInserter
	instanceStateUpdater  instanceStateUpdater
	operationInserter     operationInserter
	operationUpdater      operationUpdater
	instanceStateGetter   instanceStateGetter
	operationIDProvider   func() (internal.OperationID, error)
	reSvcFinder           reSvcFinder
	reClient              v1client.ApplicationconnectorV1alpha1Interface
	accessChecker         access.ProvisionChecker
	serviceInstanceGetter serviceInstanceGetter

	mu sync.Mutex

	maxWaitTime time.Duration
	log         logrus.FieldLogger
	asyncHook   func()
}

// Provision action
func (svc *ProvisionService) Provision(ctx context.Context, osbCtx osbContext, req *osb.ProvisionRequest) (*osb.ProvisionResponse, error) {
	if !req.AcceptsIncomplete {
		return nil, errors.New("asynchronous operation mode required")
	}

	svc.mu.Lock()
	defer svc.mu.Unlock()

	iID := internal.InstanceID(req.InstanceID)

	switch state, err := svc.instanceStateGetter.IsProvisioned(iID); true {
	case err != nil:
		return nil, errors.Wrap(err, "while checking if instance is already provisioned")
	case state:
		return &osb.ProvisionResponse{Async: false}, nil
	}

	switch opIDInProgress, inProgress, err := svc.instanceStateGetter.IsProvisioningInProgress(iID); true {
	case err != nil:
		return nil, errors.Wrap(err, "while checking if instance is being provisioned")
	case inProgress:
		opKeyInProgress := osb.OperationKey(opIDInProgress)
		return &osb.ProvisionResponse{Async: true, OperationKey: &opKeyInProgress}, nil
	}

	id, err := svc.operationIDProvider()
	if err != nil {
		return nil, errors.Wrap(err, "while generating ID for operation")
	}
	opID := internal.OperationID(id)

	paramHash := "TODO"
	op := internal.InstanceOperation{
		InstanceID:  iID,
		OperationID: opID,
		Type:        internal.OperationTypeCreate,
		State:       internal.OperationStateInProgress,
		ParamsHash:  paramHash,
	}

	if err := svc.operationInserter.Insert(&op); err != nil {
		return nil, errors.Wrap(err, "while inserting instance operation to storage")
	}

	svcID := internal.ServiceID(req.ServiceID)
	svcPlanID := internal.ServicePlanID(req.PlanID)

	re, err := svc.reSvcFinder.FindOneByServiceID(internal.RemoteServiceID(req.ServiceID))
	if err != nil {
		return nil, errors.Wrapf(err, "while getting remote environment with id: %s to storage", req.ServiceID)
	}

	namespace, err := getNamespaceFromContext(req.Context)
	if err != nil {
		return nil, errors.Wrap(err, "while getting namespace from context")
	}

	service, err := getSvcByID(re.Services, internal.RemoteServiceID(req.ServiceID))
	if err != nil {
		return nil, errors.Wrapf(err, "while getting service [%s] from RemoteEnvironment [%s]", req.ServiceID, re.Name)
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
		return nil, errors.Wrap(err, "while inserting instance to storage")
	}

	opKey := osb.OperationKey(op.OperationID)
	resp := &osb.ProvisionResponse{
		OperationKey: &opKey,
		Async:        true,
	}

	svc.doAsync(iID, opID, re.Name, getRemoteServiceID(req), namespace, service.EventProvider, service.DisplayName)
	return resp, nil
}

func getRemoteServiceID(req *osb.ProvisionRequest) internal.RemoteServiceID {
	return internal.RemoteServiceID(req.ServiceID)
}

func (svc *ProvisionService) doAsync(iID internal.InstanceID, opID internal.OperationID, reName internal.RemoteEnvironmentName, reID internal.RemoteServiceID, ns internal.Namespace, eventProvider bool, displayName string) {
	go svc.do(iID, opID, reName, reID, ns, eventProvider, displayName)
}

func (svc *ProvisionService) do(iID internal.InstanceID, opID internal.OperationID, reName internal.RemoteEnvironmentName, reID internal.RemoteServiceID, ns internal.Namespace, eventProvider bool, displayName string) {
	if svc.asyncHook != nil {
		defer svc.asyncHook()
	}
	canProvisionOutput, err := svc.accessChecker.CanProvision(iID, reID, ns, svc.maxWaitTime)
	svc.log.Infof("Access checker: canProvisionInstance(reName=[%s], reID=[%s], ns=[%s]) returned: canProvisionOutput=[%+v], error=[%v]", reName, reID, ns, canProvisionOutput, err)

	var instanceState internal.InstanceState
	var opState internal.OperationState
	var opDesc string

	if err != nil {
		instanceState = internal.InstanceStateFailed
		opState = internal.OperationStateFailed
		opDesc = fmt.Sprintf("provisioning failed on error: %s", err.Error())
	} else if !canProvisionOutput.Allowed {
		instanceState = internal.InstanceStateFailed
		opState = internal.OperationStateFailed
		opDesc = fmt.Sprintf("Forbidden provisioning instance [%s] for remote environment [name: %s, id: %s] in namespace: [%s]. Reason: [%s]", iID, reName, reID, ns, canProvisionOutput.Reason)
	} else {
		instanceState = internal.InstanceStateSucceeded
		opState = internal.OperationStateSucceeded
		opDesc = "provisioning succeeded"
		if eventProvider {
			err := svc.createEaOnSuccessProvision(string(reName), string(reID), string(ns), displayName, iID)
			if err != nil {
				instanceState = internal.InstanceStateFailed
				opState = internal.OperationStateFailed
				opDesc = fmt.Sprintf("provisioning failed while creating EventActivation on error: %s", err.Error())
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

func (svc *ProvisionService) createEaOnSuccessProvision(reName, reID, ns string, displayName string, iID internal.InstanceID) error {
	// instance ID is the serviceInstance.Spec.ExternalID
	si, err := svc.serviceInstanceGetter.GetByNamespaceAndExternalID(ns, string(iID))
	if err != nil {
		return errors.Wrapf(err, "while getting service instance with external id: %q in namespace: %q", iID, ns)
	}
	ea := &v1alpha1.EventActivation{
		ObjectMeta: v1.ObjectMeta{
			Name:      reID,
			Namespace: ns,
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
			SourceID:    reName,
		},
	}
	_, err = svc.reClient.EventActivations(ns).Create(ea)
	switch {
	case err == nil:
		svc.log.Infof("Created EventActivation: [%s], in namespace: [%s]", reID, ns)
	case apiErrors.IsAlreadyExists(err):
		// We perform update action to adjust OwnerReference of the EventActivation after the backup restore.
		_, err := svc.reClient.EventActivations(ns).Update(ea)
		if err != nil {
			return errors.Wrapf(err, "while updating EventActivation with name: %q in namespace: %q", reID, ns)
		}
		svc.log.Infof("Updated EventActivation: [%s], in namespace: [%s]", reID, ns)
	default:
		return errors.Wrapf(err, "while creating EventActivation with name: %q in namespace: %q", reID, ns)
	}
	return nil
}

func getNamespaceFromContext(contextProfile map[string]interface{}) (internal.Namespace, error) {
	return internal.Namespace(contextProfile["namespace"].(string)), nil
}

func getSvcByID(services []internal.Service, id internal.RemoteServiceID) (internal.Service, error) {
	for _, svc := range services {
		if svc.ID == id {
			return svc, nil
		}
	}
	return internal.Service{}, errors.Errorf("cannot find service with ID [%s]", id)
}
