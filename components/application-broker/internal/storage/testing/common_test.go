package testing

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/application-broker/internal/storage"
	"github.com/kyma-project/kyma/components/application-broker/internal/storage/driver/memory"
)

var allDrivers = map[storage.DriverType]func() storage.ConfigList{
	storage.DriverMemory: func() storage.ConfigList {
		return storage.ConfigList{storage.Config{
			Driver:  storage.DriverMemory,
			Provide: storage.ProviderConfigMap{storage.EntityAll: storage.ProviderConfig{}},
			Memory: memory.Config{
				// Ignored for now
				MaxKeys: 666,
			},
		}}
	},
}

func tRunDrivers(t *testing.T, tName string, f func(*testing.T, storage.Factory)) bool {
	result := true
	for dt, clGen := range allDrivers {
		cl := clGen()

		fT := func(t *testing.T) {
			sf, err := storage.NewFactory(&cl)
			require.NoError(t, err)

			f(t, sf)
		}
		result = t.Run(fmt.Sprintf("%s/%s", dt, tName), fT) && result
	}

	return result
}
