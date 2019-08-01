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

// NewAddon creates new storage for Addons
func NewAddon(cli clientv3.KV) (*Addon, error) {

	prefixParts := append(entityNamespacePrefixParts(), string(entityNamespaceAddon))
	kv := namespace.NewKV(cli, strings.Join(prefixParts, entityNamespaceSeparator))

	d := &Addon{
		generic: generic{
			kv: kv,
		},
	}

	return d, nil
}

// Addon implements etcd storage for Addon entities.
type Addon struct {
	generic
}

// Upsert persists object in storage.
//
// If addon already exists in storage than full replace is performed.
//
// True is returned if object already existed in storage and was replaced.
func (s *Addon) Upsert(namespace internal.Namespace, b *internal.Addon) (bool, error) {
	nv, err := s.nameVersionFromAddon(b)
	if err != nil {
		return false, err
	}

	dso, err := s.encodeDMToDSO(b)
	if err != nil {
		return false, err
	}

	// TODO: switch to transaction wrapping writes to both spaces
	idSpaceResp, err := s.kv.Put(context.TODO(), s.idKey(namespace, b.ID), dso, clientv3.WithPrevKV())
	if err != nil {
		return false, errors.Wrap(err, "while calling database in ID space")
	}

	// Addon is immutable so for simplicity we are duplicating write into Name/Version namespace
	if _, err := s.kv.Put(context.TODO(), s.nameVersionKey(namespace, nv), dso, clientv3.WithPrevKV()); err != nil {
		return false, errors.Wrap(err, "while calling database in NameVersion space")
	}

	if idSpaceResp.PrevKv != nil {
		return true, nil
	}

	return false, nil
}

// Get returns object from storage.
func (s *Addon) Get(namespace internal.Namespace, name internal.AddonName, ver semver.Version) (*internal.Addon, error) {
	nv, err := s.nameVersion(name, ver)
	if err != nil {
		return nil, err
	}

	resp, err := s.kv.Get(context.TODO(), s.nameVersionKey(namespace, nv))
	if err != nil {
		return nil, errors.Wrap(err, "while calling database")
	}

	return s.handleGetResp(resp)
}

// GetByID returns object by primary ID from storage.
func (s *Addon) GetByID(namespace internal.Namespace, id internal.AddonID) (*internal.Addon, error) {
	resp, err := s.kv.Get(context.TODO(), s.idKey(namespace, id))
	if err != nil {
		return nil, errors.Wrap(err, "while calling database")
	}

	return s.handleGetResp(resp)
}

func (s *Addon) handleGetResp(resp *clientv3.GetResponse) (*internal.Addon, error) {
	switch resp.Count {
	case 1:
	case 0:
		return nil, notFoundError{}
	default:
		return nil, errors.New("more than one element matching requested id, should never happen")
	}

	return s.decodeDSOToDM(resp.Kvs[0].Value)
}

func (s *Addon) encodeDMToDSO(dm *internal.Addon) (string, error) {
	buf := bytes.Buffer{}
	dso, err := newAddonDSO(dm)
	if err != nil {
		return "", errors.Wrap(err, "while encoding Addon to DSO")
	}
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(dso); err != nil {
		return "", errors.Wrap(err, "while encoding entity")
	}
	return buf.String(), nil
}

func (*Addon) decodeDSOToDM(dsoEnc []byte) (*internal.Addon, error) {
	dec := gob.NewDecoder(bytes.NewReader(dsoEnc))
	var dso addonDSO
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
func (s *Addon) FindAll(namespace internal.Namespace) ([]*internal.Addon, error) {
	var out []*internal.Addon

	resp, err := s.kv.Get(context.TODO(), s.idPrefixForNamespace(namespace), clientv3.WithPrefix())
	if err != nil {
		return nil, errors.Wrap(err, "while calling database")
	}

	for _, kv := range resp.Kvs {
		a, err := s.decodeDSOToDM(kv.Value)
		if err != nil {
			return nil, errors.Wrap(err, "while decoding returned entities")
		}
		out = append(out, a)
	}

	return out, nil
}

// Remove removes object from storage.
func (s *Addon) Remove(namespace internal.Namespace, name internal.AddonName, ver semver.Version) error {
	nv, err := s.nameVersion(name, ver)
	if err != nil {
		return errors.Wrap(err, "while getting nameVersion from deleted entity")
	}

	resp, err := s.kv.Delete(context.TODO(), s.nameVersionKey(namespace, nv), clientv3.WithPrevKV())
	if err != nil {
		return errors.Wrap(err, "while calling database on NV namespace")
	}

	b, err := s.handleDeleteResp(resp)
	if err != nil {
		return err
	}

	if _, err := s.kv.Delete(context.TODO(), s.idKey(namespace, b.ID)); err != nil {
		return errors.Wrap(err, "while calling database on ID namespace")
	}

	return nil
}

// RemoveByID is removing object by primary ID from storage.
func (s *Addon) RemoveByID(namespace internal.Namespace, id internal.AddonID) error {
	resp, err := s.kv.Delete(context.TODO(), s.idKey(namespace, id), clientv3.WithPrevKV())
	if err != nil {
		return errors.Wrap(err, "while calling database on ID namespace")
	}

	b, err := s.handleDeleteResp(resp)
	if err != nil {
		return err
	}

	nv, err := s.nameVersionFromAddon(b)
	if err != nil {
		return errors.Wrap(err, "while getting nameVersion from deleted entity")
	}

	if _, err := s.kv.Delete(context.TODO(), s.nameVersionKey(namespace, nv)); err != nil {
		return errors.Wrap(err, "while calling database on NV namespace")
	}

	return nil
}

// RemoveAll removes all addons from storage for a given namespace.
func (s *Addon) RemoveAll(namespace internal.Namespace) error {
	addons, err := s.FindAll(namespace)
	if err != nil {
		return errors.Wrap(err, "while getting addons")
	}
	for _, addon := range addons {
		if err := s.RemoveByID(namespace, addon.ID); err != nil {
			return errors.Wrapf(err, "while removing addon with ID: %v", addon.ID)
		}
	}
	return nil
}

func (s *Addon) handleDeleteResp(resp *clientv3.DeleteResponse) (*internal.Addon, error) {
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

type addonNameVersion string

func (s *Addon) nameVersionFromAddon(b *internal.Addon) (k addonNameVersion, err error) {
	if b == nil {
		return k, errors.New("entity may not be nil")
	}

	if b.Name == "" || b.Version.Original() == "" {
		return k, errors.New("both name and version must be set")
	}

	return s.nameVersion(b.Name, b.Version)
}

func (*Addon) nameVersion(name internal.AddonName, ver semver.Version) (k addonNameVersion, err error) {
	if name == "" || ver.Original() == "" {
		return k, errors.New("both name and version must be set")
	}

	return addonNameVersion(fmt.Sprintf("%s|%s", name, ver.String())), nil
}

func (s *Addon) idKey(namespace internal.Namespace, id internal.AddonID) string {
	return strings.Join([]string{s.idPrefixForNamespace(namespace), string(id)}, entityNamespaceSeparator)
}

func (*Addon) idPrefixForNamespace(namespace internal.Namespace) string {
	if namespace == internal.ClusterWide {
		return strings.Join([]string{entityNamespaceAddonMappingID, "cluster"}, entityNamespaceSeparator)
	}
	return strings.Join([]string{entityNamespaceAddonMappingID, "ns", string(namespace)}, entityNamespaceSeparator)
}

func (*Addon) nameVersionKey(namespace internal.Namespace, nv addonNameVersion) string {
	if namespace == internal.ClusterWide {
		return strings.Join([]string{entityNamespaceAddonMappingNV, "cluster", string(nv)}, entityNamespaceSeparator)
	}
	return strings.Join([]string{entityNamespaceAddonMappingNV, "ns", string(namespace), string(nv)}, entityNamespaceSeparator)
}

func newAddonDSO(in *internal.Addon) (*addonDSO, error) {
	dsoPlans := map[internal.AddonPlanID]addonPlanDSO{}
	for k, v := range in.Plans {
		var err error
		if dsoPlans[k], err = newAddonPlanDSO(v); err != nil {
			return nil, errors.Wrap(err, "while converting Addon to DSO")
		}
	}
	return &addonDSO{
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

type addonDSO struct {
	ID                  internal.AddonID
	Name                internal.AddonName
	Version             string
	Description         string
	Plans               map[internal.AddonPlanID]addonPlanDSO
	Metadata            internal.AddonMetadata
	Tags                []internal.AddonTag
	Requires            []string
	Bindable            bool
	BindingsRetrievable bool
	PlanUpdatable       *bool
}

func newAddonPlanDSO(plan internal.AddonPlan) (addonPlanDSO, error) {
	chartValuesDSO, err := newChartValuesDSO(plan.ChartValues)
	if err != nil {
		return addonPlanDSO{}, errors.Wrap(err, "while converting AddonPlan to DSO")
	}

	return addonPlanDSO{
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

type addonPlanDSO struct {
	ID           internal.AddonPlanID
	Name         internal.AddonPlanName
	Description  string
	Schemas      map[internal.PlanSchemaType]internal.PlanSchema
	ChartRef     internal.ChartRef
	ChartValues  chartValuesDSO
	Metadata     internal.AddonPlanMetadata
	BindTemplate internal.AddonPlanBindTemplate
	Bindable     *bool
	Free         *bool
}

func (dso *addonPlanDSO) ToModel() (internal.AddonPlan, error) {
	chValues, err := dso.ChartValues.ToModel()
	if err != nil {
		return internal.AddonPlan{}, errors.Wrap(err, "while converting addonPlanDSO to model")
	}
	return internal.AddonPlan{
		ID:           dso.ID,
		BindTemplate: dso.BindTemplate,
		Metadata:     dso.Metadata,
		Description:  dso.Description,
		Bindable:     dso.Bindable,
		ChartRef:     dso.ChartRef,
		Name:         dso.Name,
		Schemas:      dso.Schemas,
		Free:         dso.Free,
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

func (dto *addonDSO) NewModel() (*internal.Addon, error) {
	// TODO: do deep copy so that we are completely separated from PB entity

	plans := map[internal.AddonPlanID]internal.AddonPlan{}
	for k, v := range dto.Plans {
		var err error
		if plans[k], err = v.ToModel(); err != nil {
			return nil, errors.Wrap(err, "while converting AddonDSO to model")
		}
	}

	out := internal.Addon{
		ID:                  dto.ID,
		Name:                dto.Name,
		Description:         dto.Description,
		Plans:               plans,
		Metadata:            dto.Metadata,
		Tags:                dto.Tags,
		Bindable:            dto.Bindable,
		PlanUpdatable:       dto.PlanUpdatable,
		BindingsRetrievable: dto.BindingsRetrievable,
		Requires:            dto.Requires,
	}

	ver, err := semver.NewVersion(dto.Version)
	if err != nil {
		return nil, errors.Wrap(err, "while decoding version")
	}

	out.Version = *ver

	return &out, nil
}
