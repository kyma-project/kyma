package etcd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/namespace"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"k8s.io/helm/pkg/proto/hapi/chart"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
)

// NewChart creates new storage for Charts
func NewChart(cli clientv3.KV) (*Chart, error) {

	prefixParts := append(entityNamespacePrefixParts(), string(entityNamespaceChart))
	kv := namespace.NewKV(cli, strings.Join(prefixParts, entityNamespaceSeparator))

	d := &Chart{
		generic: generic{
			kv: kv,
		},
	}

	return d, nil
}

// Chart provides storage operations on Chart entity
type Chart struct {
	generic
}

// Upsert persists Chart in memory.
//
// If chart already exists in storage than full replace is performed.
//
// Replace is set to true if chart already existed in storage and was replaced.
func (s *Chart) Upsert(c *chart.Chart) (replaced bool, err error) {
	nv, err := s.nameVersionFromChart(c)
	if err != nil {
		return false, err
	}

	data, err := proto.Marshal(c)
	if err != nil {
		return false, errors.Wrap(err, "while encoding DSO")
	}

	resp, err := s.kv.Put(context.TODO(), s.key(nv), string(data), clientv3.WithPrevKV())
	if err != nil {
		return false, errors.Wrap(err, "while calling database")
	}

	if resp.PrevKv != nil {
		return true, nil
	}

	return false, nil
}

// Get returns chart with given name and version from storage
func (s *Chart) Get(name internal.ChartName, ver semver.Version) (*chart.Chart, error) {
	nv, err := s.nameVersion(name, ver)
	if err != nil {
		return nil, err
	}

	resp, err := s.kv.Get(context.TODO(), s.key(nv))
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

	var c chart.Chart
	if err := proto.Unmarshal(resp.Kvs[0].Value, &c); err != nil {
		return nil, errors.Wrap(err, "while decoding DSO")
	}

	return &c, nil
}

// Remove is removing chart with given name and version from storage
func (s *Chart) Remove(name internal.ChartName, ver semver.Version) error {
	nv, err := s.nameVersion(name, ver)
	if err != nil {
		return errors.Wrap(err, "while getting nameVersion from deleted entity")
	}

	resp, err := s.kv.Delete(context.TODO(), s.key(nv))
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

type chartNameVersion string

func (s *Chart) nameVersionFromChart(c *chart.Chart) (k chartNameVersion, err error) {
	if c == nil {
		return k, errors.New("entity may not be nil")
	}

	if c.Metadata == nil {
		return k, errors.New("entity metadata may not be nil")
	}

	if c.Metadata.Name == "" || c.Metadata.Version == "" {
		return k, errors.New("both name and version must be set")
	}

	ver, err := semver.NewVersion(c.Metadata.Version)
	if err != nil {
		return k, errors.Wrap(err, "while parsing version")
	}

	return s.nameVersion(internal.ChartName(c.Metadata.Name), *ver)
}

func (*Chart) nameVersion(name internal.ChartName, ver semver.Version) (k chartNameVersion, err error) {
	if name == "" || ver.Original() == "" {
		return k, errors.New("both name and version must be set")
	}

	return chartNameVersion(fmt.Sprintf("%s|%s", name, ver.String())), nil
}

func (*Chart) key(nv chartNameVersion) string {
	return string(nv)
}
