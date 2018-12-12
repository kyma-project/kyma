package storage

import (
	"github.com/kyma-project/kyma/components/application-broker/internal/storage/driver/memory"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// Factory provides access to concrete storage.
// Multiple calls should to specific storage return the same storage instance.
type Factory interface {
	RemoteEnvironment() RemoteEnvironment
	Instance() Instance
	InstanceOperation() InstanceOperation
}

// DriverType defines type of data storage
type DriverType string

const (
	// DriverMemory is a driver to local in-memory store
	DriverMemory DriverType = "memory"
)

// EntityName defines name of the entity in database
type EntityName string

const (
	// EntityAll represents name of all entities
	EntityAll EntityName = "all"
	// EntityRemoteEnvironment represents name of remote environment entities
	EntityRemoteEnvironment EntityName = "remoteenvironment"
	// EntityInstance represents name of services instances entities
	EntityInstance EntityName = "instance"
	// EntityInstanceOperation represents name of instances operations entities
	EntityInstanceOperation EntityName = "instanceOperation"
)

// ProviderConfig provides configuration to the database provider
type ProviderConfig struct{}

// ProviderConfigMap contains map of provided configurations for given entities
type ProviderConfigMap map[EntityName]ProviderConfig

// Config contains database configuration.
type Config struct {
	Driver  DriverType        `json:"driver" valid:"required"`
	Provide ProviderConfigMap `json:"provide" valid:"required"`
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

// NewFactory is a factory for entities based on given ConfigList
func NewFactory(cl *ConfigList) (Factory, error) {
	fact := concreteFactory{}

	for _, cfg := range *cl {

		var (
			remoteEnvironmentFactory func() (RemoteEnvironment, error)
			instanceFactory          func() (Instance, error)
			instanceOperationFactory func() (InstanceOperation, error)
		)

		switch cfg.Driver {
		case DriverMemory:
			remoteEnvironmentFactory = func() (RemoteEnvironment, error) {
				return memory.NewRemoteEnvironment(), nil
			}
			instanceFactory = func() (Instance, error) {
				return memory.NewInstance(), nil
			}
			instanceOperationFactory = func() (InstanceOperation, error) {
				return memory.NewInstanceOperation(), nil
			}
		default:
			return nil, errors.New("unknown driver type")
		}

		for em := range cfg.Provide {
			switch em {
			case EntityRemoteEnvironment:
				fact.re, _ = remoteEnvironmentFactory()
			case EntityInstance:
				fact.instance, _ = instanceFactory()
			case EntityInstanceOperation:
				fact.op, _ = instanceOperationFactory()
			case EntityAll:
				fact.re, _ = remoteEnvironmentFactory()
				fact.instance, _ = instanceFactory()
				fact.op, _ = instanceOperationFactory()
			default:
			}
		}
	}

	return &fact, nil
}

type concreteFactory struct {
	re       RemoteEnvironment
	instance Instance
	op       InstanceOperation
}

func (f *concreteFactory) RemoteEnvironment() RemoteEnvironment {
	return f.re
}

func (f *concreteFactory) Instance() Instance {
	return f.instance
}

func (f *concreteFactory) InstanceOperation() InstanceOperation {
	return f.op
}
