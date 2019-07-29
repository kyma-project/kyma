package memory

import (
	"fmt"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
)

type addonKey string

// NewAddon creates new storage for Addons
func NewAddon() *Addon {
	return &Addon{
		ketToID: make(map[addonKey]internal.AddonID),
		storage: make(map[internal.Namespace]map[internal.AddonID]*internal.Addon),
	}
}

// Addon implements in-memory storage for Addon entities.
type Addon struct {
	threadSafeStorage
	ketToID map[addonKey]internal.AddonID
	storage map[internal.Namespace]map[internal.AddonID]*internal.Addon
}

// Upsert persists object in storage.
//
// If addon already exists in storage than full replace is performed.
//
// True is returned if object already existed in storage and was replaced.
func (s *Addon) Upsert(namespace internal.Namespace, addon *internal.Addon) (replaced bool, err error) {
	defer unlock(s.lockW())

	if addon == nil {
		return replaced, errors.New("entity may not be nil")
	}

	nvk, err := s.keyFromAddon(namespace, addon)
	if err != nil {
		return replaced, err
	}
	replaced = true

	if _, found := s.ketToID[nvk]; !found {
		s.ketToID[nvk] = addon.ID
		replaced = false
	}

	if _, exists := s.storage[namespace]; !exists {
		s.storage[namespace] = make(map[internal.AddonID]*internal.Addon)
	}
	s.storage[namespace][addon.ID] = addon
	return replaced, nil
}

// Get returns object from storage.
func (s *Addon) Get(namespace internal.Namespace, name internal.AddonName, ver semver.Version) (*internal.Addon, error) {
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
func (s *Addon) GetByID(namespace internal.Namespace, id internal.AddonID) (*internal.Addon, error) {
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
func (s *Addon) FindAll(namespace internal.Namespace) ([]*internal.Addon, error) {
	defer unlock(s.lockR())

	out := []*internal.Addon{}
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
func (s *Addon) Remove(namespace internal.Namespace, name internal.AddonName, ver semver.Version) error {
	defer unlock(s.lockW())

	id, err := s.id(namespace, name, ver)
	if err != nil {
		return err
	}

	return s.removeByID(namespace, id)
}

// RemoveByID is removing object by primary ID from storage.
func (s *Addon) RemoveByID(namespace internal.Namespace, id internal.AddonID) error {
	defer unlock(s.lockW())

	return s.removeByID(namespace, id)
}

// RemoveAll removes all addons from storage.
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

func (s *Addon) removeByID(namespace internal.Namespace, id internal.AddonID) error {
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

func (s *Addon) keyFromAddon(namespace internal.Namespace, b *internal.Addon) (k addonKey, err error) {
	if b == nil {
		return k, errors.New("entity may not be nil")
	}

	if b.Name == "" || b.Version.Original() == "" {
		return k, errors.New("both name and version must be set")
	}

	return s.key(namespace, b.Name, b.Version)
}

func (*Addon) key(namespace internal.Namespace, name internal.AddonName, ver semver.Version) (k addonKey, err error) {
	if name == "" || ver.Original() == "" {
		return k, errors.New("both name and version must be set")
	}

	return addonKey(fmt.Sprintf("%s|%s|%s", namespace, name, ver.String())), nil
}

func (s *Addon) id(namespace internal.Namespace, name internal.AddonName, ver semver.Version) (id internal.AddonID, err error) {
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
