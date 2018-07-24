package memory

import (
	"fmt"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"

	"k8s.io/helm/pkg/proto/hapi/chart"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
)

type chartNameVersion string

// NewChart creates new storage for Charts
func NewChart() *Chart {
	return &Chart{
		storage: make(map[chartNameVersion]*chart.Chart),
	}
}

// Chart entity
type Chart struct {
	threadSafeStorage
	storage map[chartNameVersion]*chart.Chart
}

// Upsert persists Chart in memory.
//
// If chart already exists in storage than full replace is performed.
//
// True is returned if chart already existed in storage and was replaced.
func (s *Chart) Upsert(c *chart.Chart) (replaced bool, err error) {
	defer unlock(s.lockW())

	if c == nil {
		return replaced, errors.New("entity may not be nil")
	}

	nvk, err := s.keyFromChart(c)
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
func (s *Chart) Get(name internal.ChartName, ver semver.Version) (*chart.Chart, error) {
	defer unlock(s.lockR())

	nkv, err := s.key(name, ver)
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
func (s *Chart) Remove(name internal.ChartName, ver semver.Version) error {
	defer unlock(s.lockW())

	nkv, err := s.key(name, ver)
	if err != nil {
		return err
	}

	if _, found := s.storage[nkv]; !found {
		return notFoundError{}
	}

	delete(s.storage, nkv)

	return nil
}

func (s *Chart) keyFromChart(c *chart.Chart) (k chartNameVersion, err error) {
	if c == nil {
		return k, errors.New("entity may not be nil")
	}

	if c.Metadata == nil {
		return k, errors.New("entity metadata may not be nil")
	}

	if c.Metadata.Name == "" || c.Metadata.Version == "" {
		return k, errors.New("both name and version must be set")
	}

	return chartNameVersion(fmt.Sprintf("%s|%s", c.Metadata.Name, c.Metadata.Version)), nil
}

func (*Chart) key(name internal.ChartName, ver semver.Version) (k chartNameVersion, err error) {
	if name == "" || ver.Original() == "" {
		return k, errors.New("both name and version must be set")
	}

	return chartNameVersion(fmt.Sprintf("%s|%s", name, ver.String())), nil
}
