package certificates

import (
	"context"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/secrets"

	"github.com/sirupsen/logrus"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

type Migrator struct {
	ctx                  context.Context
	secretRepository     secrets.Repository
	includeSourceKeyFunc IncludeKeyFunc
	doNotOverwrite       bool
}

func NewMigrator(secretRepository secrets.Repository, includeSourceKeyFunc IncludeKeyFunc, doNotOverwrite bool) Migrator {
	return Migrator{
		ctx:                  context.Background(),
		secretRepository:     secretRepository,
		includeSourceKeyFunc: includeSourceKeyFunc,
		doNotOverwrite:       doNotOverwrite,
	}
}

type IncludeKeyFunc func(string) bool

func (m Migrator) DoMoveCredentialSecret(source, target types.NamespacedName) error {
	logrus.Info("Checking if credentials need to be migrated.")

	if source.Name == "" {
		logrus.Infof("Skipping secret migration. Source secret name is empty.")
		return nil
	}

	_, targetExists, err := m.getSecret(target)

	if err != nil {
		logrus.Errorf("Failed to read target secret: %v", err)
		return err
	}

	// TODO: It is just one difference between Do and DoMoveCredentialSecret.
	// Merge and Use doNotOverwrite
	if targetExists {
		logrus.Infof("Skipping secret migration. Target secret %s already exists in %s namespace.", target.Name, target.Namespace)
		return nil
	}

	sourceData, sourceExists, err := m.getSecret(source)
	if err != nil {
		logrus.Errorf("Failed to read source secret: %v", err)
		return err
	}

	if !sourceExists {
		logrus.Infof("Skipping secret migration. Source secret %s doesn't exist in %s namespace.", source.Name, source.Namespace)
		return nil
	}

	logrus.Infof("Moving credential secret. Source: %s , target=%s.", source.String(), target.String())

	//err = m.createSecret(target, filterOut(sourceData, m.includeSourceKeyFunc))
	err = m.createSecret(target, sourceData)
	if err != nil {
		logrus.Errorf("Failed to create target secret: %v", err)
		return err
	}

	return nil
}

func (m Migrator) Do(source, target types.NamespacedName) error {
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

func (m Migrator) getSecret(name types.NamespacedName) (map[string][]byte, bool, error) {
	data, err := m.secretRepository.Get(name)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return map[string][]byte{}, false, nil
		}

		return map[string][]byte{}, false, err
	}

	return data, true, nil
}

func (m Migrator) createSecret(name types.NamespacedName, data map[string][]byte) error {
	return m.secretRepository.UpsertWithReplace(name, data)
}

func (m Migrator) deleteSecret(name types.NamespacedName) error {
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
