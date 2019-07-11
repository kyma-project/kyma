package main

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
)

const namespace = "stage"

func TestMigrationService_MigrateOneURL(t *testing.T) {
	// GIVEN
	sch, err := v1alpha1.SchemeBuilder.Build()
	v1.AddToScheme(sch)
	require.NoError(t, err)

	repoURL := "http://url.to.repo.com/"
	repoConfigMap := fixConfigMapWithRepoURLs("main", []string{repoURL})

	cli := fake.NewFakeClientWithScheme(sch, repoConfigMap)
	svc := NewMigrationService(cli, namespace)

	// WHEN
	err = svc.Migrate()
	require.NoError(t, err)

	// THEN
	expCAC := &v1alpha1.ClusterAddonsConfiguration{}
	err = cli.Get(context.TODO(), types.NamespacedName{Name: "main", Namespace: ""}, expCAC)
	require.NoError(t, err)
	assert.Equal(t, 1, len(expCAC.Spec.Repositories))
	assert.Equal(t, "http://url.to.repo.com/index.yaml", expCAC.Spec.Repositories[0].URL)

	assertNumberOfConfigMaps(t, cli, 0)
}

func TestMigrationService_MigrateTwoURLs(t *testing.T) {
	// GIVEN
	sch, err := v1alpha1.SchemeBuilder.Build()
	v1.AddToScheme(sch)
	require.NoError(t, err)

	repoURL := "http://url.to.repo.com/index.yaml\nhttps://repo.com/prod.yaml"
	repoConfigMap := fixConfigMapWithRepoURLs("main", []string{repoURL})

	cli := fake.NewFakeClientWithScheme(sch, repoConfigMap)
	svc := NewMigrationService(cli, namespace)

	// WHEN
	err = svc.Migrate()
	require.NoError(t, err)

	// THEN
	expCAC := &v1alpha1.ClusterAddonsConfiguration{}
	err = cli.Get(context.TODO(), types.NamespacedName{Name: "main"}, expCAC)
	require.NoError(t, err)
	assert.Equal(t, 2, len(expCAC.Spec.Repositories))
	assert.Equal(t, "http://url.to.repo.com/index.yaml", expCAC.Spec.Repositories[0].URL)
	assert.Equal(t, "https://repo.com/prod.yaml", expCAC.Spec.Repositories[1].URL)
	assertNumberOfConfigMaps(t, cli, 0)
}

func TestMigrationService_MigrateTwoConfigMaps(t *testing.T) {
	// GIVEN
	sch, err := v1alpha1.SchemeBuilder.Build()
	v1.AddToScheme(sch)
	require.NoError(t, err)

	expectedURLs := map[string]string{
		"first":  "http://frist.com/index.yaml",
		"second": "http://second.com/index.yaml",
	}
	firstCM := fixConfigMapWithRepoURLs("first", []string{expectedURLs["first"]})
	secondCM := fixConfigMapWithRepoURLs("second", []string{expectedURLs["second"]})
	otherCM := &v1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "other"}}

	cli := fake.NewFakeClientWithScheme(sch, firstCM, secondCM, otherCM)
	svc := NewMigrationService(cli, namespace)

	// WHEN
	err = svc.Migrate()
	require.NoError(t, err)

	// THEN
	cacList := &v1alpha1.ClusterAddonsConfigurationList{}
	err = cli.List(context.TODO(), &client.ListOptions{}, cacList)
	require.NoError(t, err)
	assert.Equal(t, 2, len(cacList.Items))

	for _, cac := range cacList.Items {
		expected, found := expectedURLs[cac.Name]
		assert.True(t, found)
		assert.Equal(t, expected, cac.Spec.Repositories[0].URL)
	}
	assertNumberOfConfigMaps(t, cli, 1)
}

func fixConfigMapWithRepoURLs(name string, urls []string) *v1.ConfigMap {
	joinedURLs := strings.Join(urls, "\n")
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    map[string]string{"helm-broker-repo": "true"},
		},
		Data: map[string]string{"URLs": joinedURLs},
	}
}

func assertNumberOfConfigMaps(t *testing.T, cli client.Client, expected int) {
	cmList := &v1.ConfigMapList{}
	err := cli.List(context.TODO(), &client.ListOptions{}, cmList)
	require.NoError(t, err)
	assert.Equal(t, expected, len(cmList.Items))
}
