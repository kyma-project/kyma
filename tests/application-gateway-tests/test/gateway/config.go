package gateway

import (
	"fmt"
	"log"

	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type TestConfig struct {
	GatewayNamespace string `envconfig:"default=gateway-tests"`

	MockServiceURL  string `envconfig:"default=http://mock:8080"`
	MockServicePort int32  `envconfig:"default=8080"`
}

func (c TestConfig) String() string {
	return fmt.Sprintf("GatewayNamespace=%s, MockServiceURL=%s, MockServerPort=%d",
		c.GatewayNamespace, c.MockServiceURL, c.MockServicePort)
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
