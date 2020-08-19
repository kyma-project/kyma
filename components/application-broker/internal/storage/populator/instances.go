package populator

import (
	"encoding/json"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/internal/broker"
	"github.com/kyma-project/kyma/components/application-broker/internal/nsbroker"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	readyMode                           = "ready"
	provisionPollingLastOperationMode   = "provisionPollingLastOperation"
	deprovisionCallFailedMode           = "deprovisionCallFailed"
	deprovisionPollingLastOperationMode = "deprovisionPollingLastOperation"
)

// Instances provides method for populating Instance Storage
type Instances struct {
	inserter    instanceInserter
	converter   instanceConverter
	opInserter  operationInserter
	scClientSet clientset.Interface
	broker      brokerProcesses
	idSelector  applicationServiceIDSelector
	log         logrus.FieldLogger
}

// NewInstances returns Instances object
func NewInstances(
	scClientSet clientset.Interface,
	inserter instanceInserter,
	converter instanceConverter,
	operationInserter operationInserter,
	broker brokerProcesses,
	idSelector applicationServiceIDSelector,
	log logrus.FieldLogger) *Instances {
	return &Instances{
		scClientSet: scClientSet,
		inserter:    inserter,
		converter:   converter,
		opInserter:  operationInserter,
		broker:      broker,
		idSelector:  idSelector,
		log:         log,
	}
}

// Do populates Instance Storage
func (p *Instances) Do() error {
	p.log.Info("Instance storage population...")

	serviceClasses, err := p.scClientSet.ServicecatalogV1beta1().ServiceClasses(v1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "while listing service classes")
	}

	abClassNames := make(map[string]bool)
	for _, sc := range serviceClasses.Items {
		if sc.Spec.ServiceBrokerName == nsbroker.NamespacedBrokerName {
			abClassNames[sc.Name] = true
		}
	}

	serviceInstances, err := p.scClientSet.ServicecatalogV1beta1().ServiceInstances(v1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "while listing service instances")
	}

	for _, si := range serviceInstances.Items {
		if si.Spec.ServiceClassRef != nil {
			if _, ex := abClassNames[si.Spec.ServiceClassRef.Name]; ex {
				p.log.Infof("process ServiceInstance (%s)", si.Spec.ExternalID)
				if err := p.restoreInstanceData(&si); err != nil {
					return errors.Wrap(err, "while saving service instance data")
				}
			}
		}
	}

	p.log.Info("Instance storage populated")
	return nil
}

func (p *Instances) restoreInstanceData(si *v1beta1.ServiceInstance) error {
	var addInstance bool

	switch p.specifyRestoreMode(si) {
	case readyMode:
		p.log.Info("ServiceInstance is in ready state")
		_, err := p.addOperation(si, internal.OperationTypeCreate, internal.OperationStateSucceeded, "")
		if err != nil {
			return errors.Wrap(err, "while add operation for ready instance")
		}
		addInstance = true
	case provisionPollingLastOperationMode:
		p.log.Info("ServiceInstance is in failed provision state with PollingLastOperation error")
		var params map[string]interface{}
		if si.Spec.Parameters != nil {
			err := json.Unmarshal(si.Spec.Parameters.Raw, &params)
			if err != nil {
				return errors.Wrap(err, "while unmarshaling instance parameters")
			}
		}
		p.log.Info("restore provisioning process")
		err := p.broker.ProvisionProcess(broker.RestoreProvisionRequest{
			Parameters:           params,
			InstanceID:           internal.InstanceID(si.Spec.ExternalID),
			OperationID:          internal.OperationID(*si.Status.LastOperation),
			Namespace:            internal.Namespace(si.Namespace),
			ApplicationServiceID: p.idSelector.SelectApplicationServiceID(si.Spec.ServiceClassRef.Name, si.Spec.ServicePlanRef.Name),
		})
		if err != nil {
			return errors.Wrap(err, "while triggering provisioning process")
		}

		_, err = p.addOperation(si, internal.OperationTypeCreate, internal.OperationStateInProgress, internal.OperationID(*si.Status.LastOperation))
		if err != nil {
			return errors.Wrap(err, "while add operation for failed instance (PollingLastOperation error)")
		}
		addInstance = true
	case deprovisionPollingLastOperationMode, deprovisionCallFailedMode:
		p.log.Info("ServiceInstance is in failed deprovision state")
		_, err := p.addOperation(si, internal.OperationTypeCreate, internal.OperationStateSucceeded, "")
		if err != nil {
			return errors.Wrap(err, "while add operation for failed instance (DeprovisionCallFailed)")
		}
		opID, err := p.addOperation(si, internal.OperationTypeRemove, internal.OperationStateInProgress, "")
		if err != nil {
			return errors.Wrap(err, "while add operation for failed instance (DeprovisionCallFailed)")
		}
		p.broker.DeprovisionProcess(broker.DeprovisionProcessRequest{
			Instance:             p.converter.MapServiceInstance(si),
			OperationID:          opID,
			ApplicationServiceID: p.idSelector.SelectApplicationServiceID(si.Spec.ServiceClassRef.Name, si.Spec.ServicePlanRef.Name),
		})
	default:
		p.log.Infof("ServiceInstance will not be populated, unsupported mode (%s - last state: %s)",
			si.Status.CurrentOperation,
			si.Status.LastConditionState)
	}

	if !addInstance {
		return nil
	}

	if err := p.inserter.Insert(p.converter.MapServiceInstance(si)); err != nil {
		return errors.Wrap(err, "while inserting service instance")
	}

	return nil
}

func (p *Instances) specifyRestoreMode(instance *v1beta1.ServiceInstance) string {
	for _, cond := range instance.Status.Conditions {
		if cond.Type == v1beta1.ServiceInstanceConditionReady && cond.Status == v1beta1.ConditionTrue {
			return readyMode
		}
	}

	// case when AppBroker restarts before SC receives response with operation ID
	if instance.Status.LastConditionState == "DeprovisionCallFailed" {
		return deprovisionCallFailedMode
	}

	// if AppBroker restarts after SC receives a response with operation ID,
	// field 'status.LastOperation' should contain mentioned operation ID
	if instance.Status.LastOperation == nil {
		return ""
	}

	// case when AppBroker restarts after SC receives response with operation ID
	if instance.Status.LastConditionState == "ErrorPollingLastOperation" {
		switch instance.Status.CurrentOperation {
		case v1beta1.ServiceInstanceOperationProvision:
			return provisionPollingLastOperationMode
		case v1beta1.ServiceInstanceOperationDeprovision:
			return deprovisionPollingLastOperationMode
		}
	}

	return ""
}

func (p *Instances) addOperation(si *v1beta1.ServiceInstance, operationType internal.OperationType, state internal.OperationState, operationID internal.OperationID) (internal.OperationID, error) {
	if operationID == "" {
		newOperationID, err := p.broker.NewOperationID()
		if err != nil {
			return "", errors.Wrap(err, "while generating new instance operation ID")
		}
		operationID = newOperationID
	}

	err := p.opInserter.Insert(&internal.InstanceOperation{
		InstanceID:  internal.InstanceID(si.Spec.ExternalID),
		OperationID: operationID,
		Type:        operationType,
		State:       state,
	})
	if err != nil {
		return operationID, errors.Wrap(err, "while inserting instance operation")
	}

	return operationID, nil
}
