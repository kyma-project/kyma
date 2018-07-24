package controller_test

import (
	"fmt"
	"testing"

	"github.com/kyma-project/kyma/components/binding-usage-controller/internal/controller"
	sbuTypes "github.com/kyma-project/kyma/components/binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	k8sTesting "k8s.io/client-go/testing"
)

func TestBindingUsageSpecStorageGetSuccess(t *testing.T) {
	// given
	var (
		fixSBUSpec          = fixUsageSpec()
		fixMarshaledSBUSpec = mustMarshal(fixSBUSpec)

		fixUsageName          = "test-usage"
		fixUsageNamespace     = "test-usage-ns"
		fixUsageSpecCfgMapKey = fmt.Sprintf("%s.%s.spec.usedBy", fixUsageNamespace, fixUsageName)
		cfgData               = map[string]string{fixUsageSpecCfgMapKey: fixMarshaledSBUSpec}
		fixCfgMap             = fixConfigMap(cfgData)

		k8sCli       = fake.NewSimpleClientset(fixCfgMap)
		cfgMapClient = k8sCli.CoreV1().ConfigMaps(fixCfgMap.Namespace)
	)

	specStorage := controller.NewBindingUsageSpecStorage(
		cfgMapClient,
		fixCfgMap.Name)

	// when
	storedSpec, found, err := specStorage.Get(fixUsageNamespace, fixUsageName)

	// then
	require.NoError(t, err)

	assert.True(t, found)
	assert.Equal(t, fixSBUSpec, storedSpec)
}

func TestBindingUsageSpecStorageGetFailure(t *testing.T) {
	t.Run("given usage was not found in config map", func(t *testing.T) {
		// given
		var (
			fixUsageName      = "test-usage"
			fixUsageNamespace = "test-usage-ns"
			emptyCfgData      = map[string]string{}
			fixCfgMap         = fixConfigMap(emptyCfgData)

			k8sCli       = fake.NewSimpleClientset(fixCfgMap)
			cfgMapClient = k8sCli.CoreV1().ConfigMaps(fixCfgMap.Namespace)
		)

		specStorage := controller.NewBindingUsageSpecStorage(
			cfgMapClient,
			fixCfgMap.Name)

		// when
		storedSpec, found, err := specStorage.Get(fixUsageNamespace, fixUsageName)

		// then
		require.NoError(t, err)

		assert.False(t, found)
		assert.Nil(t, storedSpec)
	})

	t.Run("config map does not exists", func(t *testing.T) {
		// given
		var (
			fixUsageName      = "test-usage"
			fixUsageNamespace = "test-usage-ns"

			k8sCli       = fake.NewSimpleClientset()
			cfgMapClient = k8sCli.CoreV1().ConfigMaps("system")
		)

		specStorage := controller.NewBindingUsageSpecStorage(
			cfgMapClient,
			"not-existing-cm")

		// when
		storedSpec, found, err := specStorage.Get(fixUsageNamespace, fixUsageName)

		// then
		assertErrorContainsStatement(t, err, `configmaps "not-existing-cm" not found`)

		assert.False(t, found)
		assert.Nil(t, storedSpec)
	})

	t.Run("malformed data in config map", func(t *testing.T) {
		// given
		var (
			fixMarshaledSBUSpec = `wrong config data for given usage`

			fixUsageName          = "test-usage"
			fixUsageNamespace     = "test-usage-ns"
			fixUsageSpecCfgMapKey = fmt.Sprintf("%s.%s.spec.usedBy", fixUsageNamespace, fixUsageName)
			cfgData               = map[string]string{fixUsageSpecCfgMapKey: fixMarshaledSBUSpec}
			fixCfgMap             = fixConfigMap(cfgData)

			k8sCli       = fake.NewSimpleClientset(fixCfgMap)
			cfgMapClient = k8sCli.CoreV1().ConfigMaps(fixCfgMap.Namespace)
		)

		specStorage := controller.NewBindingUsageSpecStorage(
			cfgMapClient,
			fixCfgMap.Name)

		// when
		storedSpec, found, err := specStorage.Get(fixUsageNamespace, fixUsageName)

		// then
		assertErrorContainsStatement(t, err, "while unmarshalling spec")

		assert.False(t, found)
		assert.Nil(t, storedSpec)
	})
}

func TestBindingUsageSpecStorageDeleteSuccess(t *testing.T) {
	// given
	fixSBUSpec := fixUsageSpec()
	fixMarshaledSBUSpec := mustMarshal(fixSBUSpec)

	fixUsageName := "test-usage"
	fixUsageNamespace := "test-usage-ns"
	fixUsageSpecCfgMapKey := fmt.Sprintf("%s.%s.spec.usedBy", fixUsageNamespace, fixUsageName)

	fixCfgData := map[string]string{
		fixUsageSpecCfgMapKey: fixMarshaledSBUSpec,
		"another-entry":       "another-entry-value",
	}
	expCfgData := map[string]string{
		"another-entry": "another-entry-value",
	}

	fixCfgMap := fixConfigMap(fixCfgData)
	expCfgMap := fixConfigMap(expCfgData)

	k8sCli := fake.NewSimpleClientset(fixCfgMap)
	cfgMapClient := k8sCli.CoreV1().ConfigMaps(fixCfgMap.Namespace)

	specStorage := controller.NewBindingUsageSpecStorage(
		cfgMapClient,
		fixCfgMap.Name)

	// when
	err := specStorage.Delete(fixUsageNamespace, fixUsageName)

	// then
	require.NoError(t, err)

	performedActions := k8sCli.Actions()
	require.Len(t, performedActions, 2)

	checkAction(t, getConfigMapAction(fixCfgMap), performedActions[0])
	checkAction(t, updateConfigMapAction(expCfgMap), performedActions[1])
}

func TestBindingUsageSpecStorageDeleteFailure(t *testing.T) {
	t.Run("config map does not exists", func(t *testing.T) {
		// given
		var (
			fixUsageName      = "test-usage"
			fixUsageNamespace = "test-usage-ns"

			k8sCli       = fake.NewSimpleClientset()
			cfgMapClient = k8sCli.CoreV1().ConfigMaps("system")
		)

		specStorage := controller.NewBindingUsageSpecStorage(
			cfgMapClient,
			"not-existing-cm")

		// when
		err := specStorage.Delete(fixUsageNamespace, fixUsageName)

		// then
		assertErrorContainsStatement(t, err, `configmaps "not-existing-cm" not found`)
	})

	t.Run("entry deletion failed because error occurred on ConfigMap update", func(t *testing.T) {
		// given
		fixSBUSpec := fixUsageSpec()
		fixMarshaledSBUSpec := mustMarshal(fixSBUSpec)

		fixUsageName := "test-usage"
		fixUsageNamespace := "test-usage-ns"
		fixUsageSpecCfgMapKey := fmt.Sprintf("%s.%s.spec.usedBy", fixUsageNamespace, fixUsageName)

		fixCfgData := map[string]string{
			fixUsageSpecCfgMapKey: fixMarshaledSBUSpec,
			"another-entry":       "another-entry-value",
		}

		fixCfgMap := fixConfigMap(fixCfgData)

		k8sCli := fake.NewSimpleClientset(fixCfgMap)
		k8sCli.PrependReactor(failOnUpdateConfigMap())

		cfgMapClient := k8sCli.CoreV1().ConfigMaps(fixCfgMap.Namespace)

		specStorage := controller.NewBindingUsageSpecStorage(
			cfgMapClient,
			fixCfgMap.Name)

		// when
		err := specStorage.Delete(fixUsageNamespace, fixUsageName)

		// then
		assertErrorContainsStatement(t, err, "while updating config map")
	})
}

func TestBindingUsageSpecStorageUpsertSuccess(t *testing.T) {
	tests := map[string]struct {
		givenCfgData map[string]string
		expCfgData   map[string]string
	}{
		"update existing entry": {
			givenCfgData: map[string]string{
				"another-entry":         "another-entry-value",
				fixUsageSpecCfgMapKey(): `old-spec`,
			},
			expCfgData: map[string]string{
				"another-entry":         "another-entry-value",
				fixUsageSpecCfgMapKey(): mustMarshal(fixUsageSpec()),
			},
		},
		"add new entry": {
			givenCfgData: map[string]string{
				"another-entry": "another-entry-value",
			},
			expCfgData: map[string]string{
				"another-entry":         "another-entry-value",
				fixUsageSpecCfgMapKey(): mustMarshal(fixUsageSpec()),
			},
		},
	}

	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			fixCfgMap := fixConfigMap(tc.givenCfgData)
			expCfgMap := fixConfigMap(tc.expCfgData)

			k8sCli := fake.NewSimpleClientset(fixCfgMap)
			cfgMapClient := k8sCli.CoreV1().ConfigMaps(fixCfgMap.Namespace)

			specStorage := controller.NewBindingUsageSpecStorage(
				cfgMapClient,
				fixCfgMap.Name)

			// when
			err := specStorage.Upsert(fixServiceBindingUsage(), fixUsageSpec().Applied)

			// then
			require.NoError(t, err)

			performedActions := k8sCli.Actions()
			require.Len(t, performedActions, 2)

			checkAction(t, getConfigMapAction(fixCfgMap), performedActions[0])
			checkAction(t, updateConfigMapAction(expCfgMap), performedActions[1])
		})
	}
}

func TestBindingUsageSpecStorageUpsertFailure(t *testing.T) {
	t.Run("config map does not exists", func(t *testing.T) {
		// given
		k8sCli := fake.NewSimpleClientset()
		cfgMapClient := k8sCli.CoreV1().ConfigMaps("system")

		specStorage := controller.NewBindingUsageSpecStorage(
			cfgMapClient,
			"not-existing-cm")

		// when
		err := specStorage.Upsert(fixServiceBindingUsage(), false)

		// then
		assertErrorContainsStatement(t, err, `configmaps "not-existing-cm" not found`)
	})

	t.Run("entry insertion failed because error occurred on ConfigMap update", func(t *testing.T) {
		// given
		fixCfgData := map[string]string{
			"another-entry": "another-entry-value",
		}

		fixCfgMap := fixConfigMap(fixCfgData)

		k8sCli := fake.NewSimpleClientset(fixCfgMap)
		k8sCli.PrependReactor(failOnUpdateConfigMap())
		cfgMapClient := k8sCli.CoreV1().ConfigMaps(fixCfgMap.Namespace)

		specStorage := controller.NewBindingUsageSpecStorage(
			cfgMapClient,
			fixCfgMap.Name)

		// when
		err := specStorage.Upsert(fixServiceBindingUsage(), false)

		// then
		assertErrorContainsStatement(t, err, "while updating config map")
	})
}

func failOnUpdateConfigMap() (string, string, k8sTesting.ReactionFunc) {
	return "update", "configmaps", failingReactor
}

func fixUsageSpecCfgMapKey() string {
	return fmt.Sprintf("%s.%s.spec.usedBy", fixServiceBindingUsage().Namespace, fixServiceBindingUsage().Name)
}

func fixUsageSpec() *controller.UsageSpec {
	return &controller.UsageSpec{
		UsedBy:  fixServiceBindingUsage().Spec.UsedBy,
		Applied: true,
	}
}

func fixServiceBindingUsage() *sbuTypes.ServiceBindingUsage {
	return &sbuTypes.ServiceBindingUsage{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      "fix-sbu",
			Namespace: "fix-sub-ns",
		},
		Spec: sbuTypes.ServiceBindingUsageSpec{
			UsedBy: sbuTypes.LocalReferenceByKindAndName{
				Kind: "Deployment",
				Name: "used-by-name",
			},
			ServiceBindingRef: sbuTypes.LocalReferenceByName{
				Name: "binding-ref-name",
			},
		},
	}
}
