package etcd

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"strings"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/namespace"
	"github.com/pkg/errors"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
)

// NewInstanceBindData returns new instance of BindData storage.
func NewInstanceBindData(cli clientv3.KV) (*InstanceBindData, error) {
	prefixParts := append(entityNamespacePrefixParts(), string(entityNamespaceInstanceBindData))
	kv := namespace.NewKV(cli, strings.Join(prefixParts, entityNamespaceSeparator))

	d := &InstanceBindData{
		generic: generic{
			kv: kv,
		},
	}

	return d, nil
}

// InstanceBindData implements etcd based storage for BindData.
type InstanceBindData struct {
	generic
}

// Insert inserts object into storage.
func (s *InstanceBindData) Insert(ibd *internal.InstanceBindData) error {
	if ibd == nil {
		return errors.New("entity may not be nil")
	}

	if ibd.InstanceID.IsZero() {
		return errors.New("instance id must be set")
	}

	opKey := s.key(ibd.InstanceID)

	respGet, err := s.kv.Get(context.TODO(), opKey)
	if err != nil {
		return errors.Wrap(err, "while calling database on get")
	}
	if respGet.Count > 0 {
		return alreadyExistsError{}
	}

	dso, err := s.encodeDMToDSO(ibd)
	if err != nil {
		return err
	}

	if _, err := s.kv.Put(context.TODO(), opKey, dso); err != nil {
		return errors.Wrap(err, "while calling database on put")
	}

	return nil
}

// Get returns object from storage.
func (s *InstanceBindData) Get(iID internal.InstanceID) (*internal.InstanceBindData, error) {
	if iID.IsZero() {
		return nil, errors.New("both instance and operation id must be set")
	}

	resp, err := s.kv.Get(context.TODO(), s.key(iID))
	if err != nil {
		return nil, errors.Wrap(err, "while calling database")
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

// Remove removes object from storage.
func (s *InstanceBindData) Remove(iID internal.InstanceID) error {
	resp, err := s.kv.Delete(context.TODO(), s.key(iID))
	if err != nil {
		return errors.Wrap(err, "while calling database")
	}

	switch resp.Deleted {
	case 1:
	case 0:
		return notFoundError{}
	default:
		return fmt.Errorf("was removed more than one element matching requested id %q, should never happen", iID)
	}

	return nil
}

func (s *InstanceBindData) encodeDMToDSO(dm *internal.InstanceBindData) (string, error) {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(dm); err != nil {
		return "", errors.Wrap(err, "while encoding entity")
	}

	return buf.String(), nil
}

func (s *InstanceBindData) decodeDSOToDM(dsoEnc []byte) (*internal.InstanceBindData, error) {
	dec := gob.NewDecoder(bytes.NewReader(dsoEnc))
	var ibd internal.InstanceBindData
	if err := dec.Decode(&ibd); err != nil {
		return nil, errors.Wrap(err, "while decoding DSO")
	}

	return &ibd, nil
}

// key returns key for the specific operation in a instance space
func (s *InstanceBindData) key(iID internal.InstanceID) string {
	return string(iID)
}
