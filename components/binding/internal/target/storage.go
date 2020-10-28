package target

import (
	"errors"
	"fmt"
	"sync"

	"github.com/kyma-project/kyma/components/binding/pkg/apis/v1alpha1"
)


type KindStorage struct {
	registered map[string] v1alpha1.TargetKind
	mu sync.RWMutex
}

func NewKindStorage() *KindStorage {
	return &KindStorage{
		registered: make(map[string]v1alpha1.TargetKind),
	}
}

func (s *KindStorage) Register(tk v1alpha1.TargetKind) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.registered[tk.Name] = tk

	return nil
}

func (s *KindStorage) Unregister(tk v1alpha1.TargetKind) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.registered, tk.Name)
	return nil
}

func (s *KindStorage) Get(targetKindName string) (v1alpha1.TargetKind, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	targetKind, exists := s.registered[targetKindName]
	if !exists {
		return v1alpha1.TargetKind{}, errors.New(fmt.Sprintf("TargetKind %q was not found", targetKindName))
	}

	return targetKind, nil
}
