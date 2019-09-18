package main

import (
	"context"
	"github.com/kyma-project/kyma/components/cms-services/pkg/runtime/signal"
	"github.com/vrischmann/envconfig"

	"github.com/kyma-project/kyma/components/cms-services/pkg/endpoint/asyncapi"
	logpkg "github.com/kyma-project/kyma/components/cms-services/pkg/runtime/log"
	"github.com/kyma-project/kyma/components/cms-services/pkg/runtime/service"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type config struct {
	Verbose bool `envconfig:"default=false"`
	Service service.Config
}

func main() {
	cfg, err := loadConfig("APP")
	if err != nil {
		log.Fatal(errors.Wrap(err, "while loading the configuration"))
	}

	logpkg.Setup(cfg.Verbose)

	stopCh := signal.SetupChannel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	signal.CancelOnInterrupt(ctx, cancel, stopCh)

	srv := service.New(cfg.Service)

	log.Info("Registering endpoints")
	if err := asyncapi.AddToService(srv); err != nil {
		log.Fatal(errors.Wrap(err, "while registering endpoints"))
	}

	if err := srv.Start(ctx); err != nil {
		log.Fatal(errors.Wrap(err, "while starting the service"))
	}
}

func loadConfig(prefix string) (config, error) {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, prefix)
	return cfg, err
}
