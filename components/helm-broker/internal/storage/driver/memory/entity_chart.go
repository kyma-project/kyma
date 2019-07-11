package memory

import (
	"fmt"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"

	"k8s.io/helm/pkg/proto/hapi/chart"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
)

type chartKey string

// NewChart creates new storage for Charts
func NewChart() *Chart {
	return &Chart{
		storage: make(map[chartKey]*chart.Chart),
	}
}

// Chart entity
type Chart struct {
	threadSafeStorage
	storage map[chartKey]*chart.Chart
}

// Upsert persists Chart in memory.
//
// If chart already exists in storage than full replace is performed.
//
// True is returned if chart already existed in storage and was replaced.
func (s *Chart) Upsert(namespace internal.Namespace, c *chart.Chart) (replaced bool, err error) {
	defer unlock(s.lockW())

	if c == nil {
		return replaced, errors.New("entity may not be nil")
	}

	nvk, err := s.keyFromChart(namespace, c)
	if err != nil {
		return replaced, err
	}
	replaced = true

	if _, found := s.storage[nvk]; !found {
		replaced = false
	}
	s.storage[nvk] = c

	return replaced, nil
}

// Get returns from memory Chart with given name and version
func (s *Chart) Get(namespace internal.Namespace, name internal.ChartName, ver semver.Version) (*chart.Chart, error) {
	defer unlock(s.lockR())

	nkv, err := s.key(namespace, name, ver)
	if err != nil {
		return nil, err
	}

	c, found := s.storage[nkv]
	if !found {
		return nil, notFoundError{}
	}

	return c, nil
}

// Remove removes from memory Chart with given name and version
func (s *Chart) Remove(namespace internal.Namespace, name internal.ChartName, ver semver.Version) error {
	defer unlock(s.lockW())

	nkv, err := s.key(namespace, name, ver)
	if err != nil {
		return err
	}

	if _, found := s.storage[nkv]; !found {
		return notFoundError{}
	}

	delete(s.storage, nkv)

	return nil
}

func (s *Chart) keyFromChart(namespace internal.Namespace, c *chart.Chart) (k chartKey, err error) {
	if c == nil {
		return k, errors.New("entity may not be nil")
	}

	if c.Metadata == nil {
		return k, errors.New("entity metadata may not be nil")
	}

	if c.Metadata.Name == "" || c.Metadata.Version == "" {
		return k, errors.New("both name and version must be set")
	}

	return s.createKey(namespace, c.Metadata.Name, c.Metadata.Version)
}

func (s *Chart) key(namespace internal.Namespace, name internal.ChartName, ver semver.Version) (k chartKey, err error) {
	if name == "" || ver.Original() == "" {
		return k, errors.New("both name and version must be set")
	}

	return s.createKey(namespace, string(name), ver.Original())
}

func (*Chart) createKey(namespace internal.Namespace, name string, ver string) (k chartKey, err error) {
	return chartKey(fmt.Sprintf("%s|%s|%s", namespace, name, ver)), nil
}
