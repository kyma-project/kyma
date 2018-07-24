package etcd

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/namespace"
	"github.com/pkg/errors"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
)

// NewBundle creates new storage for Bundles
func NewBundle(cli clientv3.KV) (*Bundle, error) {

	prefixParts := append(entityNamespacePrefixParts(), string(entityNamespaceBundle))
	kv := namespace.NewKV(cli, strings.Join(prefixParts, entityNamespaceSeparator))

	d := &Bundle{
		generic: generic{
			kv: kv,
		},
	}

	return d, nil
}

// Bundle implements etcd storage for Bundle entities.
type Bundle struct {
	generic
}

// Upsert persists object in storage.
//
// If bundle already exists in storage than full replace is performed.
//
// True is returned if object already existed in storage and was replaced.
func (s *Bundle) Upsert(b *internal.Bundle) (bool, error) {
	nv, err := s.nameVersionFromBundle(b)
	if err != nil {
		return false, err
	}

	dso, err := s.encodeDMToDSO(b)
	if err != nil {
		return false, err
	}

	// TODO: switch to transaction wrapping writes to both spaces
	idSpaceResp, err := s.kv.Put(context.TODO(), s.idKey(b.ID), dso, clientv3.WithPrevKV())
	if err != nil {
		return false, errors.Wrap(err, "while calling database in ID space")
	}

	// Bundle is immutable so for simplicity we are duplicating write into Name/Version namespace
	if _, err := s.kv.Put(context.TODO(), s.nameVersionKey(nv), dso, clientv3.WithPrevKV()); err != nil {
		return false, errors.Wrap(err, "while calling database in NameVersion space")
	}

	if idSpaceResp.PrevKv != nil {
		return true, nil
	}

	return false, nil
}

// Get returns object from storage.
func (s *Bundle) Get(name internal.BundleName, ver semver.Version) (*internal.Bundle, error) {
	nv, err := s.nameVersion(name, ver)
	if err != nil {
		return nil, err
	}

	resp, err := s.kv.Get(context.TODO(), s.nameVersionKey(nv))
	if err != nil {
		return nil, errors.Wrap(err, "while calling database")
	}

	return s.handleGetResp(resp)
}

// GetByID returns object by primary ID from storage.
func (s *Bundle) GetByID(id internal.BundleID) (*internal.Bundle, error) {
	resp, err := s.kv.Get(context.TODO(), s.idKey(id))
	if err != nil {
		return nil, errors.Wrap(err, "while calling database")
	}

	return s.handleGetResp(resp)
}

func (s *Bundle) handleGetResp(resp *clientv3.GetResponse) (*internal.Bundle, error) {
	switch resp.Count {
	case 1:
	case 0:
		return nil, notFoundError{}
	default:
		return nil, errors.New("more than one element matching requested id, should never happen")
	}

	return s.decodeDSOToDM(resp.Kvs[0].Value)
}

func (s *Bundle) encodeDMToDSO(dm *internal.Bundle) (string, error) {
	buf := bytes.Buffer{}
	dso, err := newBundleDSO(dm)
	if err != nil {
		return "", errors.Wrap(err, "while encoding Bundle to DSO")
	}
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(dso); err != nil {
		return "", errors.Wrap(err, "while encoding entity")
	}
	return buf.String(), nil
}

func (*Bundle) decodeDSOToDM(dsoEnc []byte) (*internal.Bundle, error) {
	dec := gob.NewDecoder(bytes.NewReader(dsoEnc))
	var dso bundleDSO
	if err := dec.Decode(&dso); err != nil {
		return nil, errors.Wrap(err, "while decoding DSO")
	}

	b, err := dso.NewModel()
	if err != nil {
		return nil, errors.Wrap(err, "while creating model from DSO")
	}
	return b, nil
}

// FindAll returns all objects from storage.
func (s *Bundle) FindAll() ([]*internal.Bundle, error) {
	out := []*internal.Bundle{}

	resp, err := s.kv.Get(context.TODO(), entityNamespaceBundleMappingID, clientv3.WithPrefix())
	if err != nil {
		return nil, errors.Wrap(err, "while calling database")
	}

	for _, kv := range resp.Kvs {
		b, err := s.decodeDSOToDM(kv.Value)
		if err != nil {
			return nil, errors.Wrap(err, "while decoding returned entities")
		}
		out = append(out, b)
	}

	return out, nil
}

// Remove removes object from storage.
func (s *Bundle) Remove(name internal.BundleName, ver semver.Version) error {
	nv, err := s.nameVersion(name, ver)
	if err != nil {
		return errors.Wrap(err, "while getting nameVersion from deleted entity")
	}

	resp, err := s.kv.Delete(context.TODO(), s.nameVersionKey(nv), clientv3.WithPrevKV())
	if err != nil {
		return errors.Wrap(err, "while calling database on NV namespace")
	}

	b, err := s.handleDeleteResp(resp)
	if err != nil {
		return err
	}

	if _, err := s.kv.Delete(context.TODO(), s.idKey(b.ID)); err != nil {
		return errors.Wrap(err, "while calling database on ID namespace")
	}

	return nil
}

// RemoveByID is removing object by primary ID from storage.
func (s *Bundle) RemoveByID(id internal.BundleID) error {
	resp, err := s.kv.Delete(context.TODO(), s.idKey(id), clientv3.WithPrevKV())
	if err != nil {
		return errors.Wrap(err, "while calling database on ID namespace")
	}

	b, err := s.handleDeleteResp(resp)
	if err != nil {
		return err
	}

	nv, err := s.nameVersionFromBundle(b)
	if err != nil {
		return errors.Wrap(err, "while getting nameVersion from deleted entity")
	}

	if _, err := s.kv.Delete(context.TODO(), s.nameVersionKey(nv)); err != nil {
		return errors.Wrap(err, "while calling database on NV namespace")
	}

	return nil
}

func (s *Bundle) handleDeleteResp(resp *clientv3.DeleteResponse) (*internal.Bundle, error) {
	switch resp.Deleted {
	case 1:
	case 0:
		return nil, notFoundError{}
	default:
		return nil, errors.New("more than one element matching requested id, should never happen")
	}

	kv := resp.PrevKvs[0]

	b, err := s.decodeDSOToDM(kv.Value)
	if err != nil {
		return nil, err
	}

	return b, nil
}

type bundleNameVersion string

func (s *Bundle) nameVersionFromBundle(b *internal.Bundle) (k bundleNameVersion, err error) {
	if b == nil {
		return k, errors.New("entity may not be nil")
	}

	if b.Name == "" || b.Version.Original() == "" {
		return k, errors.New("both name and version must be set")
	}

	return s.nameVersion(b.Name, b.Version)
}

func (*Bundle) nameVersion(name internal.BundleName, ver semver.Version) (k bundleNameVersion, err error) {
	if name == "" || ver.Original() == "" {
		return k, errors.New("both name and version must be set")
	}

	return bundleNameVersion(fmt.Sprintf("%s|%s", name, ver.String())), nil
}

func (*Bundle) idKey(id internal.BundleID) string {
	return strings.Join([]string{entityNamespaceBundleMappingID, string(id)}, entityNamespaceSeparator)
}

func (*Bundle) nameVersionKey(nv bundleNameVersion) string {
	return strings.Join([]string{entityNamespaceBundleMappingNV, string(nv)}, entityNamespaceSeparator)
}

func newBundleDSO(in *internal.Bundle) (*bundleDSO, error) {
	dsoPlans := map[internal.BundlePlanID]bundlePlanDSO{}
	for k, v := range in.Plans {
		var err error
		if dsoPlans[k], err = newBundlePlanDSO(v); err != nil {
			return nil, errors.Wrap(err, "while converting Bundle to DSO")
		}
	}
	return &bundleDSO{
		ID:          in.ID,
		Name:        in.Name,
		Version:     in.Version.String(),
		Description: in.Description,
		Plans:       dsoPlans,
		Metadata:    in.Metadata,
		Tags:        in.Tags,
		Bindable:    in.Bindable,
	}, nil
}

type bundleDSO struct {
	ID          internal.BundleID
	Name        internal.BundleName
	Version     string
	Description string
	Plans       map[internal.BundlePlanID]bundlePlanDSO
	Metadata    internal.BundleMetadata
	Tags        []internal.BundleTag
	Bindable    bool
}

func newBundlePlanDSO(plan internal.BundlePlan) (bundlePlanDSO, error) {
	chartValuesDSO, err := newChartValuesDSO(plan.ChartValues)
	if err != nil {
		return bundlePlanDSO{}, errors.Wrap(err, "while converting BundlePlan to DSO")
	}

	return bundlePlanDSO{
		Schemas:      plan.Schemas,
		Name:         plan.Name,
		ChartRef:     plan.ChartRef,
		Bindable:     plan.Bindable,
		ChartValues:  chartValuesDSO,
		ID:           plan.ID,
		Description:  plan.Description,
		Metadata:     plan.Metadata,
		BindTemplate: plan.BindTemplate,
	}, nil
}

type bundlePlanDSO struct {
	ID           internal.BundlePlanID
	Name         internal.BundlePlanName
	Description  string
	Schemas      map[internal.PlanSchemaType]internal.PlanSchema
	ChartRef     internal.ChartRef
	ChartValues  chartValuesDSO
	Metadata     internal.BundlePlanMetadata
	Bindable     *bool
	BindTemplate internal.BundlePlanBindTemplate
}

func (dso *bundlePlanDSO) ToModel() (internal.BundlePlan, error) {
	chValues, err := dso.ChartValues.ToModel()
	if err != nil {
		return internal.BundlePlan{}, errors.Wrap(err, "while converting BundlePlanDSO to model")
	}
	return internal.BundlePlan{
		ID:           dso.ID,
		BindTemplate: dso.BindTemplate,
		Metadata:     dso.Metadata,
		Description:  dso.Description,
		Bindable:     dso.Bindable,
		ChartRef:     dso.ChartRef,
		Name:         dso.Name,
		Schemas:      dso.Schemas,
		ChartValues:  chValues,
	}, nil
}

func newChartValuesDSO(values internal.ChartValues) (chartValuesDSO, error) {
	b, err := json.Marshal(values)
	if err != nil {
		return chartValuesDSO{}, errors.Wrap(err, "while converting ChartValues to DSO")
	}
	return chartValuesDSO(b), nil
}

type chartValuesDSO json.RawMessage

func (dso chartValuesDSO) ToModel() (internal.ChartValues, error) {
	out := map[string]interface{}{}
	if err := json.Unmarshal(dso, &out); err != nil {
		return internal.ChartValues{}, errors.Wrap(err, "while converting ChartValuesDSO to model")
	}
	return out, nil
}

func (dto *bundleDSO) NewModel() (*internal.Bundle, error) {
	// TODO: do deep copy so that we are completely separated from PB entity

	plans := map[internal.BundlePlanID]internal.BundlePlan{}
	for k, v := range dto.Plans {
		var err error
		if plans[k], err = v.ToModel(); err != nil {
			return nil, errors.Wrap(err, "while converting BundleDSO to model")
		}
	}

	out := internal.Bundle{
		ID:          dto.ID,
		Name:        dto.Name,
		Description: dto.Description,
		Plans:       plans,
		Metadata:    dto.Metadata,
		Tags:        dto.Tags,
		Bindable:    dto.Bindable,
	}

	ver, err := semver.NewVersion(dto.Version)
	if err != nil {
		return nil, errors.Wrap(err, "while decoding version")
	}

	out.Version = *ver

	return &out, nil
}
