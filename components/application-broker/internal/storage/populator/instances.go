package populator

import (
	"context"

	"github.com/kyma-project/kyma/components/application-broker/internal/nsbroker"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"

	"github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset"
	scv1beta "github.com/kubernetes-sigs/service-catalog/pkg/client/informers_generated/externalversions/servicecatalog/v1beta1"
	listersv1beta "github.com/kubernetes-sigs/service-catalog/pkg/client/listers_generated/servicecatalog/v1beta1"

	v1 "k8s.io/api/core/v1"
)

// Instances provides method for populating Instance Storage
type Instances struct {
	inserter    instanceInserter
	converter   instanceConverter
	scClientSet clientset.Interface
}

// NewInstances returns Instances object
func NewInstances(scClientSet clientset.Interface, inserter instanceInserter, converter instanceConverter) *Instances {
	return &Instances{
		scClientSet: scClientSet,
		inserter:    inserter,
		converter:   converter,
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
				if err := p.inserter.Insert(p.converter.MapServiceInstance(si)); err != nil {
					return errors.Wrap(err, "while inserting service instance")
				}
			}
		}
	}
	return nil
}
