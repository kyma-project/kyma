package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MigrationService is responsible for migration from ConfigMap to Cl
type MigrationService struct {
	cli       client.Client
	namespace string
}

// NewMigrationService creates a new instance of MigrationService
func NewMigrationService(cli client.Client, namespace string) *MigrationService {
	return &MigrationService{
		cli:       cli,
		namespace: namespace,
	}
}

// Migrate performs ConfigMap with URLs to ClusterAddonConfiguration migration. The method tries to migrate all config maps. It does not stop if any errors occurs.
func (s *MigrationService) Migrate() error {
	configMaps := &v1.ConfigMapList{}

	err := s.cli.List(context.TODO(), &client.ListOptions{
		Namespace: s.namespace,
		LabelSelector: labels.SelectorFromValidatedSet(
			labels.Set(map[string]string{
				"helm-broker-repo": "true",
			}))}, configMaps)
	if err != nil {
		return err
	}

	for _, cm := range configMaps.Items {
		logrus.Infof("Migrating ConfigMap %s/%s", cm.Namespace, cm.Name)
		err := s.migrateConfigMap(&cm)
		if err != nil {
			logrus.Errorf("could not migrate ConfigMap %s/%s, error: %s", cm.Namespace, cm.Name, err.Error())
		}
	}

	return nil
}

func (s *MigrationService) migrateConfigMap(cm *v1.ConfigMap) error {
	urlsAsString, exists := cm.Data["URLs"]
	if !exists {
		return fmt.Errorf("could not find URLs fired in the 'Data'")
	}
	cac := &v1alpha1.ClusterAddonsConfiguration{}
	cac.Spec.Repositories = []v1alpha1.SpecRepository{}
	cac.Name = cm.Name

	for _, url := range strings.Split(urlsAsString, "\n") {
		if strings.HasSuffix(url, "/") {
			url = url + "index.yaml"
		}
		cac.Spec.Repositories = append(cac.Spec.Repositories, v1alpha1.SpecRepository{
			URL: url,
		})
	}

	err := s.cli.Create(context.TODO(), cac)
	if err != nil {
		return err
	}
	return s.cli.Delete(context.TODO(), cm)
}
