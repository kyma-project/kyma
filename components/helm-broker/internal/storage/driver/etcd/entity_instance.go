package etcd

import (
	"bytes"
	"context"
	"encoding/gob"
	"strings"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/namespace"
	"github.com/pkg/errors"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
)

// NewInstance creates new Instances storage
func NewInstance(cli clientv3.KV) (*Instance, error) {

	prefixParts := append(entityNamespacePrefixParts(), string(entityNamespaceInstance))
	kv := namespace.NewKV(cli, strings.Join(prefixParts, entityNamespaceSeparator))

	d := &Instance{
		generic: generic{
			kv: kv,
		},
	}

	return d, nil
}

// Instance implements etcd based storage for Instance entities.
type Instance struct {
	generic
}

// Insert inserts object to storage.
func (s *Instance) Insert(i *internal.Instance) error {
	if i == nil {
		return errors.New("entity may not be nil")
	}

	if i.ID.IsZero() {
		return errors.New("instance id must be set")
	}

	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(i); err != nil {
		return errors.Wrap(err, "while encoding entity")
	}

	respGet, err := s.kv.Get(context.TODO(), s.key(i.ID))
	if err != nil {
		return errors.Wrap(err, "while calling database on get")
	}
	if respGet.Count > 0 {
		return alreadyExistsError{}
	}

	if _, err := s.kv.Put(context.TODO(), s.key(i.ID), buf.String()); err != nil {
		return errors.Wrap(err, "while calling database on put")
	}

	return nil
}

// Get returns object from storage.
func (s *Instance) Get(id internal.InstanceID) (*internal.Instance, error) {
	resp, err := s.kv.Get(context.TODO(), s.key(id))
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

	dec := gob.NewDecoder(bytes.NewReader(resp.Kvs[0].Value))
	var i internal.Instance
	if err := dec.Decode(&i); err != nil {
		return nil, errors.Wrap(err, "while decoding DSO")
	}

	return &i, nil
}

// Remove removing object from storage.
func (s *Instance) Remove(id internal.InstanceID) error {
	resp, err := s.kv.Delete(context.TODO(), s.key(id))
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

func (*Instance) key(id internal.InstanceID) string {
	return string(id)
}
