package memory

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/kyma-project/kyma/components/application-broker/internal"
)

const remoteEnvironmentKeyPattern = "re:%s"

type remoteEnvironmentName string

// NewRemoteEnvironment creates new storage for RemoteEnvironments
func NewRemoteEnvironment() *RemoteEnvironment {
	return &RemoteEnvironment{
		storage: make(map[remoteEnvironmentName]*internal.RemoteEnvironment),
	}
}

// RemoteEnvironment entity
type RemoteEnvironment struct {
	threadSafeStorage
	storage map[remoteEnvironmentName]*internal.RemoteEnvironment
}

// Upsert persists RemoteEnvironment in memory.
//
// If RemoteEnvironment already exists in storage than full replace is performed.
//
// True is returned if RemoteEnvironment already existed in storage and was replaced.
func (s *RemoteEnvironment) Upsert(re *internal.RemoteEnvironment) (bool, error) {
	defer unlock(s.lockW())

	if re == nil {
		return false, errors.New("entity may not be nil")
	}

	nk, err := s.keyFromRE(re)
	if err != nil {
		return false, err
	}

	_, existedPreviously := s.storage[nk]

	s.storage[nk] = re

	return existedPreviously, nil
}

// Get returns from memory RemoteEnvironment with given name
func (s *RemoteEnvironment) Get(name internal.RemoteEnvironmentName) (*internal.RemoteEnvironment, error) {
	defer unlock(s.lockR())

	nk, err := s.key(name)
	if err != nil {
		return nil, err
	}

	re, found := s.storage[nk]
	if !found {
		return nil, notFoundError{}
	}

	return re, nil
}

// FindAll returns from memory all RemoteEnvironment
func (s *RemoteEnvironment) FindAll() ([]*internal.RemoteEnvironment, error) {
	defer unlock(s.lockR())

	dmList := make([]*internal.RemoteEnvironment, 0, len(s.storage))
	for _, item := range s.storage {
		dmList = append(dmList, item)
	}

	return dmList, nil
}

// FindOneByServiceID returns RemoteEnvironment which contains Service with given ID
func (s *RemoteEnvironment) FindOneByServiceID(id internal.RemoteServiceID) (*internal.RemoteEnvironment, error) {
	all, err := s.FindAll()
	if err != nil {
		return nil, errors.Wrap(err, "while reading all remote environments")
	}
	for _, re := range all {
		for _, srv := range re.Services {
			if id == srv.ID {
				return re, nil
			}
		}
	}
	return nil, nil
}

// Remove removes from memory RemoteEnvironment with given name
func (s *RemoteEnvironment) Remove(name internal.RemoteEnvironmentName) error {
	defer unlock(s.lockW())

	nk, err := s.key(name)
	if err != nil {
		return err
	}

	if _, found := s.storage[nk]; !found {
		return notFoundError{}
	}

	delete(s.storage, nk)

	return nil
}

func (s *RemoteEnvironment) keyFromRE(re *internal.RemoteEnvironment) (remoteEnvironmentName, error) {
	if re == nil {
		return "", errors.New("entity may not be nil")
	}

	return s.key(re.Name)
}

func (*RemoteEnvironment) key(name internal.RemoteEnvironmentName) (remoteEnvironmentName, error) {
	if name == "" {
		return "", errors.New("name must be set")
	}

	return remoteEnvironmentName(fmt.Sprintf(remoteEnvironmentKeyPattern, name)), nil
}
