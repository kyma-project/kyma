package memory

import (
	"sort"
	"time"

	"github.com/kyma-project/kyma/components/application-connector/application-broker/internal"
	"github.com/pkg/errors"

	pTime "github.com/kyma-project/kyma/components/application-connector/application-broker/platform/time"
)

// NewInstanceOperation returns new instance of InstanceOperation storage.
func NewInstanceOperation() *InstanceOperation {
	return &InstanceOperation{
		storage: make(map[internal.InstanceID]map[internal.OperationID]*internal.InstanceOperation),
	}
}

// InstanceOperation implements in-memory storage InstanceOperation.
type InstanceOperation struct {
	threadSafeStorage
	storage     map[internal.InstanceID]map[internal.OperationID]*internal.InstanceOperation
	nowProvider pTime.NowProvider
}

// instanceOperations implements sort.Interface and allows you to sort the slice of *internal.InstanceOperation
// by CreatedAt property in a descending order.
type instanceOperations []*internal.InstanceOperation

func (b instanceOperations) Len() int           { return len(b) }
func (b instanceOperations) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b instanceOperations) Less(i, j int) bool { return b[i].CreatedAt.After(b[j].CreatedAt) }

// WithTimeProvider allows for passing custom time provider.
// Used mostly in testing.
func (s *InstanceOperation) WithTimeProvider(nowProvider func() time.Time) {
	s.nowProvider = nowProvider
}

// Insert inserts object into storage.
func (s *InstanceOperation) Insert(io *internal.InstanceOperation) error {
	defer unlock(s.lockW())

	if io == nil {
		return errors.New("entity may not be nil")
	}

	if io.InstanceID.IsZero() || io.OperationID.IsZero() {
		return errors.New("both instance and operation id must be set")
	}

	if _, found := s.storage[io.InstanceID]; !found {
		s.storage[io.InstanceID] = make(map[internal.OperationID]*internal.InstanceOperation)
	}

	if _, found := s.storage[io.InstanceID][io.OperationID]; found {
		return alreadyExistsError{}
	}

	for oID := range s.storage[io.InstanceID] {
		if s.storage[io.InstanceID][oID].State == internal.OperationStateInProgress {
			return activeOperationInProgressError{}
		}
	}

	io.CreatedAt = s.nowProvider.Now()

	s.storage[io.InstanceID][io.OperationID] = io

	return nil
}

// Get returns object from storage.
func (s *InstanceOperation) Get(iID internal.InstanceID, opID internal.OperationID) (*internal.InstanceOperation, error) {
	defer unlock(s.lockR())

	return s.get(iID, opID)
}

func (s *InstanceOperation) get(iID internal.InstanceID, opID internal.OperationID) (*internal.InstanceOperation, error) {
	if iID.IsZero() || opID.IsZero() {
		return nil, errors.New("both instance and operation id must be set")
	}

	if _, found := s.storage[iID]; !found {
		return nil, notFoundError{}
	}

	io, found := s.storage[iID][opID]
	if !found {
		return nil, notFoundError{}
	}

	return io, nil
}

// GetAll returns all objects from storage.
func (s *InstanceOperation) GetAll(iID internal.InstanceID) ([]*internal.InstanceOperation, error) {
	defer unlock(s.lockR())

	out := instanceOperations{}

	opsForInstance, found := s.storage[iID]
	if !found {
		return nil, notFoundError{}
	}

	for i := range opsForInstance {
		out = append(out, opsForInstance[i])
	}
	sort.Sort(instanceOperations(out))
	return out, nil
}

// GetLast returns last inserted object from storage.
func (s *InstanceOperation) GetLast(iID internal.InstanceID) (*internal.InstanceOperation, error) {
	defer unlock(s.lockR())

	ops, err := s.GetAll(iID)
	if err != nil {
		return nil, err
	}

	if len(ops) == 0 {
		return nil, notFoundError{}
	}
	io := ops[0]

	return io, nil
}

// UpdateState modifies state on object in storage.
func (s *InstanceOperation) UpdateState(iID internal.InstanceID, opID internal.OperationID, state internal.OperationState) error {
	defer unlock(s.lockW())

	op, err := s.get(iID, opID)
	if err != nil {
		return err
	}

	op.State = state
	op.StateDescription = nil

	//s.logStateChange(iID, opID, state, nil)
	return nil
}

// UpdateStateDesc updates both state and description for single operation.
// If desc is nil than description will be removed.
func (s *InstanceOperation) UpdateStateDesc(iID internal.InstanceID, opID internal.OperationID, state internal.OperationState, desc *string) error {
	defer unlock(s.lockW())

	op, err := s.get(iID, opID)
	if err != nil {
		return err
	}

	op.State = state
	op.StateDescription = desc

	//s.logStateChange(iID, opID, state, desc)
	return nil
}

// Remove removes object from storage.
func (s *InstanceOperation) Remove(iID internal.InstanceID, opID internal.OperationID) error {
	defer unlock(s.lockW())

	if _, err := s.get(iID, opID); err != nil {
		return err
	}

	delete(s.storage[iID], opID)
	if len(s.storage[iID]) == 0 {
		delete(s.storage, iID)
	}

	return nil
}
