package memory

import (
	"fmt"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
)

type bundleKey string

// NewBundle creates new storage for Bundles
func NewBundle() *Bundle {
	return &Bundle{
		ketToID: make(map[bundleKey]internal.BundleID),
		storage: make(map[internal.Namespace]map[internal.BundleID]*internal.Bundle),
	}
}

// Bundle implements in-memory storage for Bundle entities.
type Bundle struct {
	threadSafeStorage
	ketToID map[bundleKey]internal.BundleID
	storage map[internal.Namespace]map[internal.BundleID]*internal.Bundle
}

// Upsert persists object in storage.
//
// If bundle already exists in storage than full replace is performed.
//
// True is returned if object already existed in storage and was replaced.
func (s *Bundle) Upsert(namespace internal.Namespace, b *internal.Bundle) (replaced bool, err error) {
	defer unlock(s.lockW())

	if b == nil {
		return replaced, errors.New("entity may not be nil")
	}

	nvk, err := s.keyFromBundle(namespace, b)
	if err != nil {
		return replaced, err
	}
	replaced = true

	if _, found := s.ketToID[nvk]; !found {
		s.ketToID[nvk] = b.ID
		replaced = false
	}

	if _, exists := s.storage[namespace]; !exists {
		s.storage[namespace] = make(map[internal.BundleID]*internal.Bundle)
	}
	s.storage[namespace][b.ID] = b
	return replaced, nil
}

// Get returns object from storage.
func (s *Bundle) Get(namespace internal.Namespace, name internal.BundleName, ver semver.Version) (*internal.Bundle, error) {
	defer unlock(s.lockR())

	id, err := s.id(namespace, name, ver)
	if err != nil {
		return nil, err
	}

	nsStorage, found := s.storage[namespace]
	if !found {
		return nil, notFoundError{}
	}
	b, found := nsStorage[id]
	// storage consistency failed - serious internal error
	if !found {
		// attempt to self-heal storage by removal of mapping
		if nvk, err := s.key(namespace, name, ver); err == nil {
			delete(s.ketToID, nvk)
		}
		return nil, notFoundError{}
	}

	return b, nil
}

// GetByID returns object by primary ID from storage.
func (s *Bundle) GetByID(namespace internal.Namespace, id internal.BundleID) (*internal.Bundle, error) {
	defer unlock(s.lockR())

	nsStorage, found := s.storage[namespace]
	if !found {
		return nil, notFoundError{}
	}
	b, found := nsStorage[id]
	// storage consistency failed - serious internal error
	if !found {
		return nil, notFoundError{}
	}
	return b, nil
}

// FindAll returns all objects from storage for a given Namespace.
func (s *Bundle) FindAll(namespace internal.Namespace) ([]*internal.Bundle, error) {
	defer unlock(s.lockR())

	out := []*internal.Bundle{}
	nsStorage, found := s.storage[namespace]
	if !found {
		return out, nil
	}

	for _, b := range nsStorage {
		out = append(out, b)
	}

	return out, nil
}

// Remove removes object from storage.
func (s *Bundle) Remove(namespace internal.Namespace, name internal.BundleName, ver semver.Version) error {
	defer unlock(s.lockW())

	id, err := s.id(namespace, name, ver)
	if err != nil {
		return err
	}

	return s.removeByID(namespace, id)
}

// RemoveByID is removing object by primary ID from storage.
func (s *Bundle) RemoveByID(namespace internal.Namespace, id internal.BundleID) error {
	defer unlock(s.lockW())

	return s.removeByID(namespace, id)
}

// RemoveAll removes all bundles from storage.
func (s *Bundle) RemoveAll(namespace internal.Namespace) error {
	bundles, err := s.FindAll(namespace)
	if err != nil {
		return errors.Wrap(err, "while getting bundles")
	}
	for _, bundle := range bundles {
		if err := s.RemoveByID(namespace, bundle.ID); err != nil {
			return errors.Wrapf(err, "while removing bundle with ID: %v", bundle.ID)
		}
	}
	return nil
}

func (s *Bundle) removeByID(namespace internal.Namespace, id internal.BundleID) error {
	nsStorage, found := s.storage[namespace]
	if !found {
		return notFoundError{}
	}
	if _, found := nsStorage[id]; !found {
		return notFoundError{}
	}

	delete(nsStorage, id)
	if len(nsStorage) == 0 {
		delete(s.storage, namespace)
	}

	return nil
}

func (s *Bundle) keyFromBundle(namespace internal.Namespace, b *internal.Bundle) (k bundleKey, err error) {
	if b == nil {
		return k, errors.New("entity may not be nil")
	}

	if b.Name == "" || b.Version.Original() == "" {
		return k, errors.New("both name and version must be set")
	}

	return s.key(namespace, b.Name, b.Version)
}

func (*Bundle) key(namespace internal.Namespace, name internal.BundleName, ver semver.Version) (k bundleKey, err error) {
	if name == "" || ver.Original() == "" {
		return k, errors.New("both name and version must be set")
	}

	return bundleKey(fmt.Sprintf("%s|%s|%s", namespace, name, ver.String())), nil
}

func (s *Bundle) id(namespace internal.Namespace, name internal.BundleName, ver semver.Version) (id internal.BundleID, err error) {
	nvk, err := s.key(namespace, name, ver)
	if err != nil {
		return id, err
	}

	id, found := s.ketToID[nvk]
	if !found {
		return id, notFoundError{}
	}

	return id, nil
}
