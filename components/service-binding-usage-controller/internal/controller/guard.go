package controller

import (
	"time"

	sbuClient "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned/typed/servicecatalog/v1alpha1"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

const guardDelay = 10 * time.Minute

type guard struct {
	items          []*guardBag
	delay          time.Duration
	log            logrus.FieldLogger
	kindSupervisor kindsSupervisors
	client         sbuClient.ServicecatalogV1alpha1Interface
}

type guardBag struct {
	sbuKey        string
	lastReconcile time.Time
}

// NewGuard creates a new guard
func NewGuard(
	sbuClient sbuClient.ServicecatalogV1alpha1Interface,
	ks kindsSupervisors,
	sbuDelay time.Duration,
	log logrus.FieldLogger) *guard {
	return &guard{
		log:            log,
		delay:          sbuDelay,
		client:         sbuClient,
		kindSupervisor: ks,
	}
}

// Run launches a new loop for process action
func (g *guard) Run(stopCh <-chan struct{}) {
	ticker := time.NewTicker(guardDelay)

	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			if len(g.items) == 0 {
				continue
			}
			g.Process()
		}
	}
}

// AddBindingUsage adds new ServiceBindingUsage key to guard's items
func (g *guard) AddBindingUsage(key string) {
	for _, bag := range g.items {
		if bag.sbuKey == key {
			return
		}
	}

	g.log.Infof("New ServiceBindingUsage key %q added to guard", key)
	g.items = append(g.items, &guardBag{
		sbuKey:        key,
		lastReconcile: time.Now(),
	})
}

// RemoveBindingUsage removes ServiceBindingUsage key from guard's items
func (g *guard) RemoveBindingUsage(key string) {
	newItems := make([]*guardBag, 0)

	for _, bag := range g.items {
		if bag.sbuKey == key {
			continue
		}
		newItems = append(newItems, bag)
	}

	g.items = newItems
	g.log.Infof("ServiceBindingUsage key %q removed from guard", key)
}

// Process checks each guard's item should be processed
func (g *guard) Process() {
	now := time.Now()

	for _, item := range g.items {
		diff := now.Sub(item.lastReconcile)
		if diff < g.delay {
			continue
		}
		g.log.Infof("Guard processes ServiceBindingUsage %s", item.sbuKey)
		item.lastReconcile = time.Now()

		namespace, name, err := cache.SplitMetaNamespaceKey(item.sbuKey)
		if err != nil {
			g.log.Errorf("Cannot split ServiceBindingUsage key %s inside guard: %v", item.sbuKey, err)
			continue
		}

		sbuItem, err := g.client.ServiceBindingUsages(namespace).Get(name, v1.GetOptions{})
		if err != nil {
			g.log.Errorf("Cannot get ServiceBindingUsage %s/%s inside guard, got error: %v", namespace, name, err)
			continue
		}

		sv, err := g.kindSupervisor.Get(Kind(sbuItem.Spec.UsedBy.Kind))
		if err != nil {
			g.log.Errorf("Guard cannot get UsageSupervisor for %s/%s key, got error: %v", namespace, name, err)
			continue
		}

		_, err = sv.GetInjectedLabels(namespace, sbuItem.Spec.UsedBy.Name, name)
		switch {
		case IsNotFoundError(err):
			g.log.Infof("Guard updates ServiceBindingUsage %s/%s (UsageKind %s: %q not exist)", namespace, name, sbuItem.Spec.UsedBy.Kind, sbuItem.Spec.UsedBy.Name)
			toUpdate := sbuItem.DeepCopy()
			toUpdate.Spec.ReprocessRequest = sbuItem.Spec.ReprocessRequest + 1
			_, err = g.client.ServiceBindingUsages(toUpdate.Namespace).Update(toUpdate)
			if err != nil {
				g.log.Errorf("Guard cannot update ServiceBindingUsage for %s/%s key: %q", namespace, name, err)
			}
		case err == nil:
			continue
		default:
			g.log.Errorf("Guard cannot get response from UsageSupervisor for %s/%s key, got error: %v", namespace, name, err)
		}
	}
}
