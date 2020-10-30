package storage

import (
	"errors"
	"fmt"
	"sync"

	"github.com/kyma-project/kyma/components/binding/pkg/apis/v1alpha1"
)

type KindManager interface {
	AddLabel(*v1alpha1.Binding) error
	AddLabelToResource(b *v1alpha1.Binding) error
	RemoveLabel(*v1alpha1.Binding) error
	LabelExist(*v1alpha1.Binding) (bool, error)
	RemoveOldAddNewLabel(*v1alpha1.Binding) error
}

// Kind represents Kubernetes Kind name
type Kind string

type KindStorage struct {
	registered map[Kind] KindManager
	mu sync.RWMutex
}

func NewKindStorage() *KindStorage {
	return &KindStorage{
		registered: make(map[Kind]KindManager),
	}
}

// Register adds KindManager for given Kind to KindStorage
func (s *KindStorage) Register(kind Kind, manager KindManager) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.registered[kind] = manager

	return nil
}

// Unregister removes given Kind from KindStorage
func (s *KindStorage) Unregister(kind Kind) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.registered, kind)
	return nil
}

// Get returns KindManager for given Kind
func (s *KindStorage) Get(kind Kind) (KindManager, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	concreteKindManager, exists := s.registered[kind]
	if !exists {
		return nil, errors.New(fmt.Sprintf("TargetKind %s was not found", kind))
	}

	return concreteKindManager, nil
}
