package memory

import (
	"errors"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
)

// NewInstanceBindData returns new instance of BindData storage.
func NewInstanceBindData() *InstanceBindData {
	return &InstanceBindData{
		storage: make(map[internal.InstanceID]*internal.InstanceBindData),
	}
}

// InstanceBindData implements in-memory based storage for BindData.
type InstanceBindData struct {
	threadSafeStorage
	storage map[internal.InstanceID]*internal.InstanceBindData
}

// Insert inserts object into storage.
func (s *InstanceBindData) Insert(ibd *internal.InstanceBindData) error {
	defer unlock(s.lockW())

	if ibd == nil {
		return errors.New("entity may not be nil")
	}

	if ibd.InstanceID.IsZero() {
		return errors.New("instance id must be set")
	}

	if _, found := s.storage[ibd.InstanceID]; found {
		return alreadyExistsError{}
	}

	// TODO switch to deep copy?
	cpy := *ibd
	s.storage[ibd.InstanceID] = &cpy

	return nil
}

// Get returns object from storage.
func (s *InstanceBindData) Get(iID internal.InstanceID) (*internal.InstanceBindData, error) {
	defer unlock(s.lockR())

	i, found := s.storage[iID]
	if !found {
		return nil, notFoundError{}
	}

	return i, nil
}

// Remove removes object from storage.
func (s *InstanceBindData) Remove(iID internal.InstanceID) error {
	defer unlock(s.lockW())

	if _, found := s.storage[iID]; !found {
		return notFoundError{}
	}

	delete(s.storage, iID)

	return nil
}
