package memory

import (
	"fmt"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
)

type bundleNameVersion string

// NewBundle creates new storage for Bundles
func NewBundle() *Bundle {
	return &Bundle{
		nameVerToID: make(map[bundleNameVersion]internal.BundleID),
		storage:     make(map[internal.BundleID]*internal.Bundle),
	}
}

// Bundle implements in-memory storage for Bundle entities.
type Bundle struct {
	threadSafeStorage
	nameVerToID map[bundleNameVersion]internal.BundleID
	storage     map[internal.BundleID]*internal.Bundle
}

// Upsert persists object in storage.
//
// If bundle already exists in storage than full replace is performed.
//
// True is returned if object already existed in storage and was replaced.
func (s *Bundle) Upsert(b *internal.Bundle) (replaced bool, err error) {
	defer unlock(s.lockW())

	if b == nil {
		return replaced, errors.New("entity may not be nil")
	}

	nvk, err := s.keyFromBundle(b)
	if err != nil {
		return replaced, err
	}
	replaced = true

	if _, found := s.nameVerToID[nvk]; !found {
		s.nameVerToID[nvk] = b.ID
		replaced = false
	}

	s.storage[b.ID] = b
	return replaced, nil
}

// Get returns object from storage.
func (s *Bundle) Get(name internal.BundleName, ver semver.Version) (*internal.Bundle, error) {
	defer unlock(s.lockR())

	id, err := s.id(name, ver)
	if err != nil {
		return nil, err
	}

	b, found := s.storage[id]
	// storage consistency failed - serious internal error
	if !found {
		// attempt to self-heal storage by removal of mapping
		if nvk, err := s.key(name, ver); err == nil {
			delete(s.nameVerToID, nvk)
		}
		return nil, notFoundError{}
	}

	return b, nil
}

// GetByID returns object by primary ID from storage.
func (s *Bundle) GetByID(id internal.BundleID) (*internal.Bundle, error) {
	defer unlock(s.lockR())

	b, found := s.storage[id]
	// storage consistency failed - serious internal error
	if !found {
		return nil, notFoundError{}
	}
	return b, nil
}

// FindAll returns all objects from storage.
func (s *Bundle) FindAll() ([]*internal.Bundle, error) {
	defer unlock(s.lockR())

	out := []*internal.Bundle{}

	for id := range s.storage {
		out = append(out, s.storage[id])
	}

	return out, nil
}

// Remove removes object from storage.
func (s *Bundle) Remove(name internal.BundleName, ver semver.Version) error {
	defer unlock(s.lockW())

	id, err := s.id(name, ver)
	if err != nil {
		return err
	}

	return s.removeByID(id)
}

// RemoveByID is removing object by primary ID from storage.
func (s *Bundle) RemoveByID(id internal.BundleID) error {
	defer unlock(s.lockW())

	return s.removeByID(id)
}

func (s *Bundle) removeByID(id internal.BundleID) error {
	if _, found := s.storage[id]; !found {
		return notFoundError{}
	}

	delete(s.storage, id)

	return nil
}

func (s *Bundle) keyFromBundle(b *internal.Bundle) (k bundleNameVersion, err error) {
	if b == nil {
		return k, errors.New("entity may not be nil")
	}

	if b.Name == "" || b.Version.Original() == "" {
		return k, errors.New("both name and version must be set")
	}

	return s.key(b.Name, b.Version)
}

func (*Bundle) key(name internal.BundleName, ver semver.Version) (k bundleNameVersion, err error) {
	if name == "" || ver.Original() == "" {
		return k, errors.New("both name and version must be set")
	}

	return bundleNameVersion(fmt.Sprintf("%s|%s", name, ver.String())), nil
}

func (s *Bundle) id(name internal.BundleName, ver semver.Version) (id internal.BundleID, err error) {
	nvk, err := s.key(name, ver)
	if err != nil {
		return id, err
	}

	id, found := s.nameVerToID[nvk]
	if !found {
		return id, notFoundError{}
	}

	return id, nil
}
