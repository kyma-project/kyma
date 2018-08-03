package controller

import (
	"strings"
	"sync"
)

//go:generate mockery -name=KubernetesResourceSupervisor -output=automock -outpkg=automock -case=underscore

// KubernetesResourceSupervisor validates if given Kubernetes resource can be modified by ServiceBindingUsage. If yes
// it can ensure that labels are present or deleted on previous validated resource.
type KubernetesResourceSupervisor interface {
	HasSynced() bool
	EnsureLabelsCreated(resourceNs, resourceName, usageName string, labels map[string]string) error
	EnsureLabelsDeleted(resourceNs, resourceName, usageName string) error
	GetInjectedLabels(resourceNs, resourceName, usageName string) (map[string]string, error)
}

// Kind represents Kubernetes Kind name
type Kind string

// TODO (pluggable SBU cleanup): consider removing normalized method
func (k Kind) normalized() string {
	return strings.ToLower(string(k))
}

const (
	// KindDeployment represents Deployment resource
	KindDeployment Kind = "Deployment"
	// KindKubelessFunction represents Kubeless Function resource
	KindKubelessFunction Kind = "Function"
)

// ResourceSupervisorAggregator aggregates all defined resources supervisors
type ResourceSupervisorAggregator struct {
	registered map[string]KubernetesResourceSupervisor
	mu         sync.RWMutex
}

// NewResourceSupervisorAggregator returns new instance of ResourceSupervisorAggregator
func NewResourceSupervisorAggregator() *ResourceSupervisorAggregator {
	return &ResourceSupervisorAggregator{
		registered: make(map[string]KubernetesResourceSupervisor),
	}
}

// Register adds new resource supervisor
func (f *ResourceSupervisorAggregator) Register(k Kind, supervisor KubernetesResourceSupervisor) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.registered[k.normalized()] = supervisor
	return nil
}

// Unregister removes resource supervisor
func (f *ResourceSupervisorAggregator) Unregister(k Kind) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.registered, k.normalized())
	return nil
}

// HasSynced returns true if all registered supervisors are synced
func (f *ResourceSupervisorAggregator) HasSynced() bool {
	for _, supervisor := range f.registered {
		if !supervisor.HasSynced() {
			return false
		}
	}

	return true
}

// Get returns supervisor for given kind
func (f *ResourceSupervisorAggregator) Get(k Kind) (KubernetesResourceSupervisor, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	concreteSupervisor, exists := f.registered[k.normalized()]
	if !exists {
		return nil, NewNotFoundError("supervisor for kind %s was not found", k)
	}

	return concreteSupervisor, nil
}
