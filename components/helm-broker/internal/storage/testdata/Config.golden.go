package testdata

import "github.com/kyma-project/kyma/components/helm-broker/internal/storage"

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
				storage.EntityBundle:            storage.ProviderConfig{},
				storage.EntityChart:             storage.ProviderConfig{},
				storage.EntityInstance:          storage.ProviderConfig{},
				storage.EntityInstanceOperation: storage.ProviderConfig{},
			},
		},
	}
}

func GoldenConfigMemoryMultipleSeparate() storage.ConfigList {
	return storage.ConfigList{
		{Driver: storage.DriverMemory, Provide: storage.ProviderConfigMap{storage.EntityBundle: storage.ProviderConfig{}}},
		{Driver: storage.DriverMemory, Provide: storage.ProviderConfigMap{storage.EntityChart: storage.ProviderConfig{}}},
		{Driver: storage.DriverMemory, Provide: storage.ProviderConfigMap{storage.EntityInstance: storage.ProviderConfig{}}},
		{Driver: storage.DriverMemory, Provide: storage.ProviderConfigMap{storage.EntityInstanceOperation: storage.ProviderConfig{}}},
	}
}

func GoldenConfigEtcdSingleAll() storage.ConfigList {
	return storage.ConfigList{
		{
			Driver: storage.DriverEtcd,
			Provide: storage.ProviderConfigMap{
				storage.EntityAll: storage.ProviderConfig{},
			},
		},
	}
}

func GoldenConfigEtcdSingleSeparate() storage.ConfigList {
	return storage.ConfigList{
		{
			Driver: storage.DriverEtcd,
			Provide: storage.ProviderConfigMap{
				storage.EntityBundle:            storage.ProviderConfig{},
				storage.EntityChart:             storage.ProviderConfig{},
				storage.EntityInstance:          storage.ProviderConfig{},
				storage.EntityInstanceOperation: storage.ProviderConfig{},
			},
		},
	}
}

func GoldenConfigEtcdMultipleSeparate() storage.ConfigList {
	return storage.ConfigList{
		{Driver: storage.DriverEtcd, Provide: storage.ProviderConfigMap{storage.EntityBundle: storage.ProviderConfig{}}},
		{Driver: storage.DriverEtcd, Provide: storage.ProviderConfigMap{storage.EntityChart: storage.ProviderConfig{}}},
		{Driver: storage.DriverEtcd, Provide: storage.ProviderConfigMap{storage.EntityInstance: storage.ProviderConfig{}}},
		{Driver: storage.DriverEtcd, Provide: storage.ProviderConfigMap{storage.EntityInstanceOperation: storage.ProviderConfig{}}},
	}
}

func GoldenConfigMixEMMESeparate() storage.ConfigList {
	return storage.ConfigList{
		{Driver: storage.DriverEtcd, Provide: storage.ProviderConfigMap{storage.EntityBundle: storage.ProviderConfig{}}},
		{Driver: storage.DriverMemory, Provide: storage.ProviderConfigMap{storage.EntityChart: storage.ProviderConfig{}}},
		{Driver: storage.DriverMemory, Provide: storage.ProviderConfigMap{storage.EntityInstance: storage.ProviderConfig{}}},
		{Driver: storage.DriverEtcd, Provide: storage.ProviderConfigMap{storage.EntityInstanceOperation: storage.ProviderConfig{}}},
	}
}

func GoldenConfigMixMMEEGrouped() storage.ConfigList {
	return storage.ConfigList{
		{Driver: storage.DriverMemory, Provide: storage.ProviderConfigMap{
			storage.EntityBundle: storage.ProviderConfig{},
			storage.EntityChart:  storage.ProviderConfig{},
		}},
		{Driver: storage.DriverEtcd, Provide: storage.ProviderConfigMap{
			storage.EntityInstance:          storage.ProviderConfig{},
			storage.EntityInstanceOperation: storage.ProviderConfig{},
		}},
	}
}
