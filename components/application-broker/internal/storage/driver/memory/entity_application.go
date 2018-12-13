package memory

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/kyma-project/kyma/components/application-broker/internal"
)

const applicationKeyPattern = "app:%s"

type applicationName string

// NewApplication creates new storage for Applications
func NewApplication() *Application {
	return &Application{
		storage: make(map[applicationName]*internal.Application),
	}
}

// Application entity
type Application struct {
	threadSafeStorage
	storage map[applicationName]*internal.Application
}

// Upsert persists Application in memory.
//
// If Application already exists in storage than full replace is performed.
//
// True is returned if Application already existed in storage and was replaced.
func (s *Application) Upsert(app *internal.Application) (bool, error) {
	defer unlock(s.lockW())

	if app == nil {
		return false, errors.New("entity may not be nil")
	}

	nk, err := s.keyFromRE(app)
	if err != nil {
		return false, err
	}

	_, existedPreviously := s.storage[nk]

	s.storage[nk] = app

	return existedPreviously, nil
}

// Get returns from memory Application with given name
func (s *Application) Get(name internal.ApplicationName) (*internal.Application, error) {
	defer unlock(s.lockR())

	nk, err := s.key(name)
	if err != nil {
		return nil, err
	}

	app, found := s.storage[nk]
	if !found {
		return nil, notFoundError{}
	}

	return app, nil
}

// FindAll returns from memory all Application
func (s *Application) FindAll() ([]*internal.Application, error) {
	defer unlock(s.lockR())

	dmList := make([]*internal.Application, 0, len(s.storage))
	for _, item := range s.storage {
		dmList = append(dmList, item)
	}

	return dmList, nil
}

// FindOneByServiceID returns Application which contains Service with given ID
func (s *Application) FindOneByServiceID(id internal.ApplicationServiceID) (*internal.Application, error) {
	all, err := s.FindAll()
	if err != nil {
		return nil, errors.Wrap(err, "while reading all applications")
	}
	for _, app := range all {
		for _, srv := range app.Services {
			if id == srv.ID {
				return app, nil
			}
		}
	}
	return nil, nil
}

// Remove removes from memory Application with given name
func (s *Application) Remove(name internal.ApplicationName) error {
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

func (s *Application) keyFromRE(app *internal.Application) (applicationName, error) {
	if app == nil {
		return "", errors.New("entity may not be nil")
	}

	return s.key(app.Name)
}

func (*Application) key(name internal.ApplicationName) (applicationName, error) {
	if name == "" {
		return "", errors.New("name must be set")
	}

	return applicationName(fmt.Sprintf(applicationKeyPattern, name)), nil
}
