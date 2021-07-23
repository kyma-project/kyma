package main

import (
	"context"

	"github.com/sirupsen/logrus"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

type migrator struct {
	ctx                  context.Context
	secretRepository     SecretRepository
	includeSourceKeyFunc IncludeKeyFunc
}

func NewMigrator(secretRepository SecretRepository, includeSourceKeyFunc IncludeKeyFunc) migrator {
	return migrator{
		ctx:                  context.Background(),
		secretRepository:     secretRepository,
		includeSourceKeyFunc: includeSourceKeyFunc,
	}
}

type IncludeKeyFunc func(string) bool

func (m migrator) Do(source, target types.NamespacedName) error {
	logrus.Info("Checking if secret needs to be migrated.")
	if source.Name == "" {
		logrus.Infof("Skipping secret migration. Source secret name is empty.")
		return nil
	}

	logrus.Infof("Migrating secret. Source: %s , target=%s.", source.String(), target.String())

	sourceData, sourceExists, err := m.getSecret(source)
	if err != nil {
		logrus.Errorf("Failed to read source secret: %v", err)
		return err
	}

	if !sourceExists {
		logrus.Infof("Skipping secret migration. Source secret %s doesn't exist in %s namespace.", source.Name, source.Namespace)
		return nil
	}

	_, targetExists, err := m.getSecret(target)
	if err != nil {
		logrus.Errorf("Failed to read target secret: %v", err)
		return err
	}

	if !targetExists {
		err = m.createSecret(target, filterOut(sourceData, m.includeSourceKeyFunc))
		if err != nil {
			logrus.Errorf("Failed to create target secret: %v", err)
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

func filterOut(data map[string][]byte, includeKeyFunc IncludeKeyFunc) map[string][]byte {
	newData := make(map[string][]byte)

	for k, v := range data {
		if includeKeyFunc(k) {
			newData[k] = v
		}
	}

	return newData
}
