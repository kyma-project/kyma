package memory

import (
	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/pkg/errors"
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

func (s *Instance) get(iID internal.InstanceID) (*internal.Instance, error) {
	if iID.IsZero() {
		return nil, errors.New("instance id must be set")
	}

	if _, found := s.storage[iID]; !found {
		return nil, notFoundError{}
	}

	i, found := s.storage[iID]
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

// FindOne returns from storage first object which passes the match.
func (s *Instance) FindOne(m func(i *internal.Instance) bool) (*internal.Instance, error) {
	defer unlock(s.lockW())

	for iID, i := range s.storage {
		if m(i) {
			return s.storage[iID], nil
		}
	}

	return nil, nil
}

// FindOne returns from storage first object which passes the match.
func (s *Instance) FindAll(m func(i *internal.Instance) bool) ([]*internal.Instance, error) {
	defer unlock(s.lockW())

	var matches []*internal.Instance
	for iID, i := range s.storage {
		if m(i) {
			matches = append(matches, s.storage[iID])
		}
	}

	return matches, nil
}

// UpdateState modifies state on object in storage.
func (s *Instance) UpdateState(iID internal.InstanceID, state internal.InstanceState) error {
	defer unlock(s.lockW())

	i, err := s.get(iID)
	if err != nil {
		return err
	}

	i.State = state

	return nil
}
