package etcd

import (
	"bytes"
	"context"
	"encoding/gob"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/namespace"
	"github.com/pkg/errors"

	"github.com/kyma-project/kyma/components/helm-broker/platform/ptr"
	yTime "github.com/kyma-project/kyma/components/helm-broker/platform/time"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
)

// NewInstanceOperation returns new instance of InstanceOperation storage.
func NewInstanceOperation(cli clientv3.KV) (*InstanceOperation, error) {
	prefixParts := append(entityNamespacePrefixParts(), string(entityNamespaceInstanceOperation))
	kv := namespace.NewKV(cli, strings.Join(prefixParts, entityNamespaceSeparator))

	d := &InstanceOperation{
		generic: generic{
			kv: kv,
		},
	}

	return d, nil
}

// InstanceOperation implements etcd based storage InstanceOperation.
type InstanceOperation struct {
	generic
	nowProvider yTime.NowProvider
}

// WithTimeProvider allows for passing custom time provider.
// Used mostly in testing.
func (s *InstanceOperation) WithTimeProvider(nowProvider func() time.Time) *InstanceOperation {
	s.nowProvider = nowProvider
	return s
}

// Insert inserts object into storage.
func (s *InstanceOperation) Insert(io *internal.InstanceOperation) error {
	if io == nil {
		return errors.New("entity may not be nil")
	}

	if io.InstanceID.IsZero() || io.OperationID.IsZero() {
		return errors.New("both instance and operation id must be set")
	}

	opKey := s.key(io.InstanceID, io.OperationID)

	respGet, err := s.kv.Get(context.TODO(), opKey)
	if err != nil {
		return errors.Wrap(err, "while calling database on get")
	}
	if respGet.Count > 0 {
		return alreadyExistsError{}
	}

	opInProgress, err := s.isOperationInProgress(io.InstanceID)
	if err != nil {
		return errors.Wrap(err, "while checking if there are operations in progress")
	}
	if *opInProgress {
		return activeOperationInProgressError{}
	}

	io.CreatedAt = s.nowProvider.Now()

	dso, err := s.encodeDMToDSO(io)
	if err != nil {
		return err
	}

	if _, err := s.kv.Put(context.TODO(), opKey, dso); err != nil {
		return errors.Wrap(err, "while calling database on put")
	}

	return nil
}

func (s *InstanceOperation) isOperationInProgress(iID internal.InstanceID) (*bool, error) {
	resp, err := s.kv.Get(context.TODO(), s.instanceKeyPrefix(iID), clientv3.WithPrefix())
	if err != nil {
		return nil, s.handleGetError(err)
	}

	for _, kv := range resp.Kvs {
		io, err := s.decodeDSOToDM(kv.Value)
		if err != nil {
			return nil, errors.Wrap(err, "while decoding returned entities")
		}

		if io.State == internal.OperationStateInProgress {
			return ptr.Bool(true), nil
		}
	}

	return ptr.Bool(false), nil
}

// Get returns object from storage.
func (s *InstanceOperation) Get(iID internal.InstanceID, opID internal.OperationID) (*internal.InstanceOperation, error) {
	return s.get(iID, opID)
}

func (s *InstanceOperation) get(iID internal.InstanceID, opID internal.OperationID) (*internal.InstanceOperation, error) {
	if iID.IsZero() || opID.IsZero() {
		return nil, errors.New("both instance and operation id must be set")
	}

	resp, err := s.kv.Get(context.TODO(), s.key(iID, opID))
	if err != nil {
		return nil, s.handleGetError(err)
	}

	switch resp.Count {
	case 1:
	case 0:
		return nil, notFoundError{}
	default:
		return nil, errors.New("more than one element matching requested id, should never happen")
	}

	return s.decodeDSOToDM(resp.Kvs[0].Value)
}

// GetAll returns all objects from storage.
func (s *InstanceOperation) GetAll(iID internal.InstanceID) ([]*internal.InstanceOperation, error) {
	out := []*internal.InstanceOperation{}

	resp, err := s.kv.Get(context.TODO(), s.instanceKeyPrefix(iID), clientv3.WithPrefix())
	if err != nil {
		return nil, s.handleGetError(err)
	}

	if resp.Count == 0 {
		return nil, notFoundError{}
	}

	for _, kv := range resp.Kvs {
		io, err := s.decodeDSOToDM(kv.Value)
		if err != nil {
			return nil, errors.Wrap(err, "while decoding returned entities")
		}
		out = append(out, io)
	}

	return out, nil
}

func (*InstanceOperation) handleGetError(errIn error) error {
	return errors.Wrap(errIn, "while calling database")
}

func (s *InstanceOperation) encodeDMToDSO(dm *internal.InstanceOperation) (string, error) {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(dm); err != nil {
		return "", errors.Wrap(err, "while encoding entity")
	}

	return buf.String(), nil
}

func (s *InstanceOperation) decodeDSOToDM(dsoEnc []byte) (*internal.InstanceOperation, error) {
	dec := gob.NewDecoder(bytes.NewReader(dsoEnc))
	var io internal.InstanceOperation
	if err := dec.Decode(&io); err != nil {
		return nil, errors.Wrap(err, "while decoding DSO")
	}

	return &io, nil
}

// UpdateState modifies state on object in storage.
func (s *InstanceOperation) UpdateState(iID internal.InstanceID, opID internal.OperationID, state internal.OperationState) error {
	return s.updateStateDesc(iID, opID, state, nil)
}

// UpdateStateDesc modifies state and description on object in storage.
// If desc is nil than description will be removed.
func (s *InstanceOperation) UpdateStateDesc(iID internal.InstanceID, opID internal.OperationID, state internal.OperationState, desc *string) error {
	return s.updateStateDesc(iID, opID, state, desc)
}

func (s *InstanceOperation) updateStateDesc(iID internal.InstanceID, opID internal.OperationID, state internal.OperationState, desc *string) error {
	io, err := s.get(iID, opID)
	if err != nil {
		return err
	}

	io.State = state
	io.StateDescription = desc

	dso, err := s.encodeDMToDSO(io)
	if err != nil {
		return err
	}

	if _, err := s.kv.Put(context.TODO(), s.key(iID, opID), dso); err != nil {
		return errors.Wrap(err, "while calling database on put")
	}

	return nil
}

// Remove removes object from storage.
func (s *InstanceOperation) Remove(iID internal.InstanceID, opID internal.OperationID) error {
	resp, err := s.kv.Delete(context.TODO(), s.key(iID, opID))
	if err != nil {
		return errors.Wrap(err, "while calling database")
	}

	switch resp.Deleted {
	case 1:
	case 0:
		return notFoundError{}
	default:
		return errors.New("more than one element matching requested id, should never happen")
	}

	return nil
}

// key returns key for the specific operation in a instance space
func (s *InstanceOperation) key(iID internal.InstanceID, opID internal.OperationID) string {
	return s.instanceKeyPrefix(iID) + string(opID)
}

// instanceKeyPrefix returns prefix for all operation keys in single instance namespace
// Trailing separator is appended.
func (*InstanceOperation) instanceKeyPrefix(id internal.InstanceID) string {
	return string(id) + entityNamespaceSeparator
}
