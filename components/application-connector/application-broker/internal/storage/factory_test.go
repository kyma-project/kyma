package storage_test

import (
	"github.com/kyma-project/kyma/components/application-broker/internal/storage"
	"testing"

	"github.com/kyma-project/kyma/components/application-broker/internal/storage/driver/memory"
	"github.com/kyma-project/kyma/components/application-broker/internal/storage/testdata"
	"github.com/stretchr/testify/assert"
)

func TestNewFactory(t *testing.T) {
	for s, tc := range map[string]struct {
		cfgGen         func() storage.ConfigList
		expApplication interface{}
	}{
		"MemorySingleAll":        {testdata.GoldenConfigMemorySingleAll, &memory.Application{}},
		"MemorySingleSeparate":   {testdata.GoldenConfigMemorySingleSeparate, &memory.Application{}},
		"MemoryMultipleSeparate": {testdata.GoldenConfigMemoryMultipleSeparate, &memory.Application{}},
	} {
		t.Run(s, func(t *testing.T) {
			// GIVEN:
			cfg := tc.cfgGen()

			// WHEN:
			got, err := storage.NewFactory(&cfg)

			// THEN:
			assert.NoError(t, err)

			assert.IsType(t, tc.expApplication, got.Application())
		})
	}
}
