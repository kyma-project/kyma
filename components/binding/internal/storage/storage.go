package storage

import (
	"fmt"
	"strings"
	"sync"

	"github.com/kyma-project/kyma/components/binding/pkg/apis/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type KindManager interface {
	AddLabel(*v1alpha1.Binding) error
	RemoveLabel(*v1alpha1.Binding) error
	LabelExist(*v1alpha1.Binding) (bool, error)
	RemoveOldAddNewLabel(*v1alpha1.Binding) error
}

// Kind represents Kubernetes Kind name
type Kind string

type ResourceData struct {
	Schema      schema.GroupVersionResource
	LabelFields []string
}

type KindStorage struct {
	registered map[Kind]*ResourceData
	mu         sync.RWMutex
}

func NewKindStorage() *KindStorage {
	return &KindStorage{
		registered: make(map[Kind]*ResourceData),
	}
}

func newResourceData(gvr schema.GroupVersionResource, labelsPath string) *ResourceData {
	return &ResourceData{
		Schema:      gvr,
		LabelFields: strings.Split(labelsPath, "."),
	}
}

// Register adds KindManager for given Kind to KindStorage
func (s *KindStorage) Register(tk v1alpha1.TargetKind) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.registered[Kind(tk.Name)] = newResourceData(schema.GroupVersionResource{
		Group:    tk.Spec.Resource.Group,
		Version:  tk.Spec.Resource.Version,
		Resource: strings.ToLower(tk.Spec.Resource.Kind + "s"),
	}, tk.Spec.LabelsPath)

	return nil
}

// Unregister removes given Kind from KindStorage
func (s *KindStorage) Unregister(tk v1alpha1.TargetKind) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.registered, Kind(tk.Name))
	return nil
}

// Get returns KindManager for given Kind
func (s *KindStorage) Get(kind Kind) (*ResourceData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	concreteResourceData, exists := s.registered[kind]
	if !exists {
		return &ResourceData{}, fmt.Errorf("TargetKind %s was not found", kind)
	}
	return concreteResourceData, nil
}
