package configurer_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/asset-upload-service/internal/bucket"
	"github.com/kyma-project/kyma/components/asset-upload-service/internal/configurer"
	"github.com/onsi/gomega"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8sTesting "k8s.io/client-go/testing"
)

func TestConfigurer_LoadIfExists(t *testing.T) {
	configMapName := "test-config"
	configMapNamespace := "default"

	t.Run("Success", func(t *testing.T) {
		// Given
		private := "private-test"
		public := "public-test"

		expectedConfig := fixAppConfig(private, public)
		configMap := fixConfigMap(configMapName, configMapNamespace, private, public)

		g := gomega.NewGomegaWithT(t)
		client := fake.NewSimpleClientset(configMap)
		c := configurer.New(client.Core(), configurer.Config{
			Name:      configMapName,
			Namespace: configMapNamespace,
			Enabled:   true,
		})

		// When
		conf, err := c.Load()

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(conf).NotTo(gomega.BeNil())
		g.Expect(*conf).To(gomega.Equal(expectedConfig))
	})

	t.Run("Doesn't exist", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		client := fake.NewSimpleClientset()
		c := configurer.New(client.Core(), configurer.Config{
			Name:      "test",
			Namespace: "default",
			Enabled:   true,
		})

		// When
		conf, err := c.Load()

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(conf).To(gomega.BeNil())
	})

	t.Run("Error", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		client := fake.NewSimpleClientset()
		client.PrependReactor("get", "configmaps", failingReactor)

		c := configurer.New(client.Core(), configurer.Config{
			Name:      "test",
			Namespace: "default",
			Enabled:   true,
		})

		// When
		_, err := c.Load()

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
		g.Expect(err.Error()).To(gomega.ContainSubstring("Test error"))
	})

	t.Run("Disabled", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		c := configurer.New(nil, configurer.Config{
			Name:      configMapName,
			Namespace: configMapNamespace,
			Enabled:   false,
		})

		// When
		conf, err := c.Load()

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(conf).To(gomega.BeNil())
	})
}

func TestConfigurer_Save(t *testing.T) {
	configMapName := "test-config"
	configMapNamespace := "default"

	t.Run("Success", func(t *testing.T) {
		// Given
		private := "private-test"
		public := "public-test"

		config := fixAppConfig(private, public)
		expectedConfigMap := fixConfigMap(configMapName, configMapNamespace, private, public)

		g := gomega.NewGomegaWithT(t)
		client := fake.NewSimpleClientset()
		coreCli := client.Core()
		c := configurer.New(coreCli, configurer.Config{
			Name:      configMapName,
			Namespace: configMapNamespace,
			Enabled:   true,
		})

		// When
		err := c.Save(config)

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())

		cfgMap, err := coreCli.ConfigMaps(configMapNamespace).Get(configMapName, metav1.GetOptions{})
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(cfgMap).To(gomega.Equal(expectedConfigMap))
	})

	t.Run("Error", func(t *testing.T) {
		// Given
		private := "private-test"
		public := "public-test"

		config := fixAppConfig(private, public)

		g := gomega.NewGomegaWithT(t)
		client := fake.NewSimpleClientset()
		client.PrependReactor("create", "configmaps", failingReactor)

		coreCli := client.Core()
		c := configurer.New(coreCli, configurer.Config{
			Name:      configMapName,
			Namespace: configMapNamespace,
			Enabled:   true,
		})

		// When
		err := c.Save(config)

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
	})

	t.Run("Disabled", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		private := "private-test"
		public := "public-test"

		config := fixAppConfig(private, public)

		c := configurer.New(nil, configurer.Config{
			Name:      configMapName,
			Namespace: configMapNamespace,
			Enabled:   false,
		})

		// When
		err := c.Save(config)

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
	})
}

func failingReactor(_ k8sTesting.Action) (handled bool, ret runtime.Object, err error) {
	return true, nil, errors.New("Test error")
}

func fixAppConfig(private, public string) configurer.SharedAppConfig {
	return configurer.SharedAppConfig{
		SystemBuckets: bucket.SystemBucketNames{
			Public:  public,
			Private: private,
		},
	}
}

func fixConfigMap(name, namespace, private, public string) *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: map[string]string{
			"private": private,
			"public":  public,
		},
	}
}
