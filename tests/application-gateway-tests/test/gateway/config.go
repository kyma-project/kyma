package gateway

import (
	"fmt"
	"log"

	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type TestConfig struct {
	Namespace string `envconfig:"default=default"`

	MockSelectorKey   string `envconfig:"default=app"`
	MockSelectorValue string `envconfig:"default=mock-service"`
	MockServerPort    int32  `envconfig:"default=8080"`
}

func (c TestConfig) String() string {
	return fmt.Sprintf("Namespace=%s, MockSelectorKey=%s, MockSelectorValue=%s, MockServerPort=%d",
		c.Namespace, c.MockSelectorKey, c.MockSelectorValue, c.MockServerPort)
}

func ReadConfig() (TestConfig, error) {
	cfg := TestConfig{}

	err := envconfig.InitWithPrefix(&cfg, "")
	if err != nil {
		return TestConfig{}, errors.Wrap(err, "Error while loading app config")
	}

	log.Printf("Read configuration: %s", cfg.String())
	return cfg, nil
}
