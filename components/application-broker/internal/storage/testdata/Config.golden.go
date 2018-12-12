package testdata

import "github.com/kyma-project/kyma/components/application-broker/internal/storage"

func GoldenConfigMemorySingleAll() storage.ConfigList {
	return storage.ConfigList{
		{
			Driver: storage.DriverMemory,
			Provide: storage.ProviderConfigMap{
				storage.EntityAll: storage.ProviderConfig{},
			},
		},
	}
}

func GoldenConfigMemorySingleSeparate() storage.ConfigList {
	return storage.ConfigList{
		{
			Driver: storage.DriverMemory,
			Provide: storage.ProviderConfigMap{
				storage.EntityRemoteEnvironment: storage.ProviderConfig{},
			},
		},
	}
}

func GoldenConfigMemoryMultipleSeparate() storage.ConfigList {
	return storage.ConfigList{
		{Driver: storage.DriverMemory, Provide: storage.ProviderConfigMap{storage.EntityRemoteEnvironment: storage.ProviderConfig{}}},
	}
}
