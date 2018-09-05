package populator

import (
	"context"

	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	scv1beta "github.com/kubernetes-incubator/service-catalog/pkg/client/informers_generated/externalversions/servicecatalog/v1beta1"
	listersv1beta "github.com/kubernetes-incubator/service-catalog/pkg/client/listers_generated/servicecatalog/v1beta1"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// InstancesFromClusterBroker provide method for populating Instance storage
type InstancesFromClusterBroker struct {
	inserter    instanceInserter
	converter   instanceConverter
	scClientSet clientset.Interface
	brokerName  string
}

// NewInstancesFromClusterBroker is a constructor of InstancesFromClusterBroker populator
func NewInstancesFromClusterBroker(scClientSet clientset.Interface, inserter instanceInserter, converter instanceConverter, brokerName string) *InstancesFromClusterBroker {
	return &InstancesFromClusterBroker{
		scClientSet: scClientSet,
		inserter:    inserter,
		converter:   converter,
		brokerName:  brokerName,
	}
}

// Do perform instances population
func (p *InstancesFromClusterBroker) Do(ctx context.Context) error {
	siInformer := scv1beta.NewServiceInstanceInformer(p.scClientSet, v1.NamespaceAll, informerResyncPeriod, nil)
	scInformer := scv1beta.NewClusterServiceClassInformer(p.scClientSet, informerResyncPeriod, nil)

	go siInformer.Run(ctx.Done())
	go scInformer.Run(ctx.Done())

	if !cache.WaitForCacheSync(ctx.Done(), siInformer.HasSynced) {
		return errors.New("cannot synchronize service instance cache")
	}

	if !cache.WaitForCacheSync(ctx.Done(), scInformer.HasSynced) {
		return errors.New("cannot synchronize service class cache")
	}

	scLister := listersv1beta.NewClusterServiceClassLister(scInformer.GetIndexer())
	serviceClasses, err := scLister.List(labels.Everything())
	if err != nil {
		return errors.Wrap(err, "while listing service classes")
	}

	rebClassNames := make(map[string]struct{})
	for _, sc := range serviceClasses {
		if sc.Spec.ClusterServiceBrokerName == p.brokerName {
			rebClassNames[sc.Name] = struct{}{}
		}
	}

	siLister := listersv1beta.NewServiceInstanceLister(siInformer.GetIndexer())
	serviceInstances, err := siLister.List(labels.Everything())
	if err != nil {
		return errors.Wrap(err, "while listing service instances")
	}

	for _, si := range serviceInstances {
		if si.Spec.ClusterServiceClassRef != nil {
			if _, ex := rebClassNames[si.Spec.ClusterServiceClassRef.Name]; ex {
				if err := p.inserter.Insert(p.converter.MapServiceInstance(si)); err != nil {
					return errors.Wrap(err, "while inserting service instance")
				}
			}
		}
	}
	return nil
}

/*
Objects taking part, example:

- apiVersion: servicecatalog.k8s.io/v1beta1
  kind: ClusterServiceClass
  metadata:
    name: 48ab05bf-9aa4-4cb7-8999-0d3587265ac3
  spec:
    clusterServiceBrokerName: core-remote-environment-broker


---
- apiVersion: servicecatalog.k8s.io/v1beta1
  kind: ServiceInstance
  metadata:
    name: reb-instance-1
    namespace: default
  spec:
    clusterServiceClassExternalName: orders
    clusterServiceClassRef:
      name: 48ab05bf-9aa4-4cb7-8999-0d3587265ac3
    clusterServicePlanExternalName: default
    clusterServicePlanRef:
      name: 48ab05bf-9aa4-4cb7-8999-0d3587265ac3-plan
    externalID: b180ef2f-1215-4439-a24c-850caf78d74b

---
PUT /v2/service_instances/b180ef2f-1215-4439-a24c-850caf78d74b/

*/
