package storage_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/kyma/components/helm-broker/internal/storage"
	"github.com/kyma-project/kyma/components/helm-broker/internal/storage/driver/etcd"
	"github.com/kyma-project/kyma/components/helm-broker/internal/storage/driver/memory"
	"github.com/kyma-project/kyma/components/helm-broker/internal/storage/testdata"
)

func TestNewFactory(t *testing.T) {
	for s, tc := range map[string]struct {
		cfgGen               func() storage.ConfigList
		expAddon             interface{}
		expChart             interface{}
		expInstance          interface{}
		expInstanceOperation interface{}
	}{
		"MemorySingleAll":        {testdata.GoldenConfigMemorySingleAll, &memory.Addon{}, &memory.Chart{}, &memory.Instance{}, &memory.InstanceOperation{}},
		"MemorySingleSeparate":   {testdata.GoldenConfigMemorySingleSeparate, &memory.Addon{}, &memory.Chart{}, &memory.Instance{}, &memory.InstanceOperation{}},
		"MemoryMultipleSeparate": {testdata.GoldenConfigMemoryMultipleSeparate, &memory.Addon{}, &memory.Chart{}, &memory.Instance{}, &memory.InstanceOperation{}},
		"EtcdSingleAll":          {testdata.GoldenConfigEtcdSingleAll, &etcd.Addon{}, &etcd.Chart{}, &etcd.Instance{}, &etcd.InstanceOperation{}},
		"EtcdSingleSeparate":     {testdata.GoldenConfigEtcdSingleSeparate, &etcd.Addon{}, &etcd.Chart{}, &etcd.Instance{}, &etcd.InstanceOperation{}},
		"EtcdMultipleSeparate":   {testdata.GoldenConfigEtcdMultipleSeparate, &etcd.Addon{}, &etcd.Chart{}, &etcd.Instance{}, &etcd.InstanceOperation{}},
		"MixEMMESeparate":        {testdata.GoldenConfigMixEMMESeparate, &etcd.Addon{}, &memory.Chart{}, &memory.Instance{}, &etcd.InstanceOperation{}},
		"MixMMEEGrouped":         {testdata.GoldenConfigMixMMEEGrouped, &memory.Addon{}, &memory.Chart{}, &etcd.Instance{}, &etcd.InstanceOperation{}},
	} {
		t.Run(s, func(t *testing.T) {
			// GIVEN:
			cfg := tc.cfgGen()

			// WHEN:
			got, err := storage.NewFactory(&cfg)

			// THEN:
			assert.NoError(t, err)

			assert.IsType(t, tc.expAddon, got.Addon())
			assert.IsType(t, tc.expChart, got.Chart())
			assert.IsType(t, tc.expInstance, got.Instance())
			assert.IsType(t, tc.expInstanceOperation, got.InstanceOperation())
		})
	}
}
