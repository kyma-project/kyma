package populator

import (
	"context"
	"encoding/json"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/internal/broker"
	"github.com/kyma-project/kyma/components/application-broker/internal/nsbroker"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset"
	scv1beta "github.com/kubernetes-sigs/service-catalog/pkg/client/informers_generated/externalversions/servicecatalog/v1beta1"
	listersv1beta "github.com/kubernetes-sigs/service-catalog/pkg/client/listers_generated/servicecatalog/v1beta1"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
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
	log         logrus.FieldLogger
}

// NewInstances returns Instances object
func NewInstances(
	scClientSet clientset.Interface,
	inserter instanceInserter,
	converter instanceConverter,
	operationInserter operationInserter,
	broker brokerProcesses,
	log logrus.FieldLogger) *Instances {
	return &Instances{
		scClientSet: scClientSet,
		inserter:    inserter,
		converter:   converter,
		opInserter:  operationInserter,
		broker:      broker,
		log:         log,
	}
}

// Do populates Instance Storage
func (p *Instances) Do(ctx context.Context) error {
	siInformer := scv1beta.NewServiceInstanceInformer(p.scClientSet, v1.NamespaceAll, informerResyncPeriod, nil)
	scInformer := scv1beta.NewServiceClassInformer(p.scClientSet, v1.NamespaceAll, informerResyncPeriod, nil)

	go siInformer.Run(ctx.Done())
	go scInformer.Run(ctx.Done())

	if !cache.WaitForCacheSync(ctx.Done(), siInformer.HasSynced) {
		return errors.New("cannot synchronize service instance cache")
	}

	if !cache.WaitForCacheSync(ctx.Done(), scInformer.HasSynced) {
		return errors.New("cannot synchronize service class cache")
	}

	scLister := listersv1beta.NewServiceClassLister(scInformer.GetIndexer())
	serviceClasses, err := scLister.List(labels.Everything())
	if err != nil {
		return errors.Wrap(err, "while listing service classes")
	}

	abClassNames := make(map[string]struct{})
	for _, sc := range serviceClasses {
		if sc.Spec.ServiceBrokerName == nsbroker.NamespacedBrokerName {
			abClassNames[sc.Name] = struct{}{}
		}
	}

	siLister := listersv1beta.NewServiceInstanceLister(siInformer.GetIndexer())
	serviceInstances, err := siLister.List(labels.Everything())
	if err != nil {
		return errors.Wrap(err, "while listing service instances")
	}

	for _, si := range serviceInstances {
		if si.Spec.ServiceClassRef != nil {
			if _, ex := abClassNames[si.Spec.ServiceClassRef.Name]; ex {
				p.log.Infof("process ServiceInstance (%s)", si.Spec.ExternalID)
				if err := p.restoreInstanceData(si); err != nil {
					return errors.Wrap(err, "while saving service instance data")
				}
			}
		}
	}
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
			ApplicationServiceID: internal.ApplicationServiceID(si.Spec.ServiceClassRef.Name),
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
			ApplicationServiceID: internal.ApplicationServiceID(si.Spec.ServiceClassRef.Name),
		})
	default:
		p.log.Infof("ServiceInstance will to be populated, unsupported mode (%s - last state: %s)",
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
