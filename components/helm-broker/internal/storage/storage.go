package storage

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"github.com/kyma-project/kyma/components/helm-broker/internal/storage/driver/etcd"
	"github.com/kyma-project/kyma/components/helm-broker/internal/storage/driver/memory"
)

// Factory provides access to concrete storage.
// Multiple calls should to specific storage return the same storage instance.
type Factory interface {
	Bundle() Bundle
	Chart() Chart
	Instance() Instance
	InstanceOperation() InstanceOperation
	InstanceBindData() InstanceBindData
}

// DriverType defines type of data storage
type DriverType string

const (
	// DriverEtcd is a driver for key-value store - Etcd
	DriverEtcd DriverType = "etcd"
	// DriverMemory is a driver to local in-memory store
	DriverMemory DriverType = "memory"
)

// EntityName defines name of the entity in database
type EntityName string

const (
	// EntityAll represents name of all entities
	EntityAll EntityName = "all"
	// EntityChart represents name of chart entities
	EntityChart EntityName = "chart"
	// EntityBundle represents name of bundle entities
	EntityBundle EntityName = "bundle"
	// EntityInstance represents name of services instances entities
	EntityInstance EntityName = "instance"
	// EntityInstanceOperation represents name of instances operations entities
	EntityInstanceOperation EntityName = "instanceOperation"
	// EntityInstanceBindData represents name of bind data entities
	EntityInstanceBindData EntityName = "entityInstanceBindData"
)

// ProviderConfig provides configuration to the database provider
type ProviderConfig struct{}

// ProviderConfigMap contains map of provided configurations for given entities
type ProviderConfigMap map[EntityName]ProviderConfig

// Config contains database configuration.
type Config struct {
	Driver  DriverType        `json:"driver" valid:"required"`
	Provide ProviderConfigMap `json:"provide" valid:"required"`
	Etcd    etcd.Config       `json:"etcd"`
	Memory  memory.Config     `json:"memory"`
}

// ConfigList is a list of configurations
type ConfigList []Config

// ConfigParse is parsing yaml file to the ConfigList
func ConfigParse(inByte []byte) (*ConfigList, error) {
	var cl ConfigList

	if err := yaml.Unmarshal(inByte, &cl); err != nil {
		return nil, errors.Wrap(err, "while unmarshalling yaml")
	}

	return &cl, nil
}

// NewConfigListAllMemory returns configured configList with the memory driver for all entities.
func NewConfigListAllMemory() *ConfigList {
	return &ConfigList{{Driver: DriverMemory, Provide: ProviderConfigMap{EntityAll: ProviderConfig{}}}}
}

// NewFactory is a factory for entities based on given ConfigList
// TODO: add error handling
func NewFactory(cl *ConfigList) (Factory, error) {
	fact := concreteFactory{}

	for _, cfg := range *cl {

		var (
			bundleFact            func() (Bundle, error)
			chartFact             func() (Chart, error)
			instanceFact          func() (Instance, error)
			instanceOperationFact func() (InstanceOperation, error)
			instanceBindDataFact  func() (InstanceBindData, error)
		)

		switch cfg.Driver {
		case DriverMemory:
			bundleFact = func() (Bundle, error) {
				return memory.NewBundle(), nil
			}
			chartFact = func() (Chart, error) {
				return memory.NewChart(), nil
			}
			instanceFact = func() (Instance, error) {
				return memory.NewInstance(), nil
			}
			instanceOperationFact = func() (InstanceOperation, error) {
				return memory.NewInstanceOperation(), nil
			}
			instanceBindDataFact = func() (InstanceBindData, error) {
				return memory.NewInstanceBindData(), nil
			}
		case DriverEtcd:
			var cli etcd.Client
			if cfg.Etcd.ForceClient != nil {
				cli = cfg.Etcd.ForceClient
			} else {
				cli, _ = etcd.NewClient(cfg.Etcd)
			}

			bundleFact = func() (Bundle, error) {
				return etcd.NewBundle(cli)
			}
			chartFact = func() (Chart, error) {
				return etcd.NewChart(cli)
			}
			instanceFact = func() (Instance, error) {
				return etcd.NewInstance(cli)
			}
			instanceOperationFact = func() (InstanceOperation, error) {
				return etcd.NewInstanceOperation(cli)
			}
			instanceBindDataFact = func() (InstanceBindData, error) {
				return etcd.NewInstanceBindData(cli)
			}
		default:
			return nil, errors.New("unknown driver type")
		}

		for em := range cfg.Provide {
			switch em {
			case EntityChart:
				fact.chart, _ = chartFact()
			case EntityBundle:
				fact.bundle, _ = bundleFact()
			case EntityInstance:
				fact.instance, _ = instanceFact()
			case EntityInstanceOperation:
				fact.instanceOperation, _ = instanceOperationFact()
			case EntityInstanceBindData:
				fact.instanceBindData, _ = instanceBindDataFact()
			case EntityAll:
				fact.chart, _ = chartFact()
				fact.bundle, _ = bundleFact()
				fact.instance, _ = instanceFact()
				fact.instanceOperation, _ = instanceOperationFact()
				fact.instanceBindData, _ = instanceBindDataFact()
			default:
			}
		}
	}

	return &fact, nil
}

type concreteFactory struct {
	bundle            Bundle
	chart             Chart
	instance          Instance
	instanceOperation InstanceOperation
	instanceBindData  InstanceBindData
}

func (f *concreteFactory) Bundle() Bundle {
	return f.bundle
}
func (f *concreteFactory) Chart() Chart {
	return f.chart
}
func (f *concreteFactory) Instance() Instance {
	return f.instance
}
func (f *concreteFactory) InstanceOperation() InstanceOperation {
	return f.instanceOperation
}
func (f *concreteFactory) InstanceBindData() InstanceBindData {
	return f.instanceBindData
}
