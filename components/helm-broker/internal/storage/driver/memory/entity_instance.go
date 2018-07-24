package memory

import (
	"github.com/pkg/errors"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
)

// NewInstance creates new Instances storage
func NewInstance() *Instance {
	return &Instance{
		storage: make(map[internal.InstanceID]*internal.Instance),
	}
}

// Instance implements in-memory storage for Instance entities.
type Instance struct {
	threadSafeStorage
	storage map[internal.InstanceID]*internal.Instance
}

// Insert inserts object to storage.
func (s *Instance) Insert(i *internal.Instance) error {
	defer unlock(s.lockW())

	if i == nil {
		return errors.New("entity may not be nil")
	}

	if i.ID.IsZero() {
		return errors.New("instance id must be set")
	}

	if _, found := s.storage[i.ID]; found {
		return alreadyExistsError{}
	}

	s.storage[i.ID] = i

	return nil
}

// Get returns object from storage.
func (s *Instance) Get(id internal.InstanceID) (*internal.Instance, error) {
	defer unlock(s.lockR())

	i, found := s.storage[id]
	if !found {
		return nil, notFoundError{}
	}

	return i, nil
}

// Remove removing object from storage.
func (s *Instance) Remove(id internal.InstanceID) error {
	defer unlock(s.lockW())

	_, found := s.storage[id]
	if !found {
		return notFoundError{}
	}

	delete(s.storage, id)

	return nil
}
