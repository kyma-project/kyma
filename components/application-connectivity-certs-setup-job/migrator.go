package main

import (
	"context"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

type migrator struct {
	ctx              context.Context
	secretRepository SecretRepository
}

func NewMigrator(secretRepository SecretRepository) migrator {
	return migrator{
		ctx:              context.Background(),
		secretRepository: secretRepository,
	}
}

func (m migrator) Do(source, target types.NamespacedName) error {
	if source.Name == "" {
		return nil
	}

	sourceData, sourceExists, err := m.getSecret(source)
	if err != nil {
		return err
	}

	if !sourceExists {
		return nil
	}

	_, targetExists, err := m.getSecret(target)
	if err != nil {
		return err
	}

	if !targetExists {
		err = m.createSecret(target, sourceData)
		if err != nil {
			return err
		}
	}

	return m.deleteSecret(source)
}

func (m migrator) getSecret(name types.NamespacedName) (map[string][]byte, bool, error) {
	data, err := m.secretRepository.Get(name)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return map[string][]byte{}, false, nil
		}

		return map[string][]byte{}, false, err
	}

	return data, true, nil
}

func (m migrator) createSecret(name types.NamespacedName, data map[string][]byte) error {
	return m.secretRepository.Upsert(name, data)
}

func (m migrator) deleteSecret(name types.NamespacedName) error {
	return m.secretRepository.Delete(name)
}
