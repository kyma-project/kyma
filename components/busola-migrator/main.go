package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/pkg/errors"

	"github.com/kyma-project/kyma/components/busola-migrator/internal/app"
	"github.com/kyma-project/kyma/components/busola-migrator/internal/busola"
	"github.com/kyma-project/kyma/components/busola-migrator/internal/config"
	"github.com/kyma-project/kyma/components/busola-migrator/internal/kubernetes"
	"github.com/kyma-project/kyma/components/busola-migrator/internal/router"
)

func main() {
	cfg := config.LoadConfig()

	kubeConfig, err := kubernetes.GetKubeConfig()
	if err != nil {
		log.Fatal(errors.Wrap(err, "while getting kubeconfig"))
	}

	busolaURL, err := busola.BuildInitURL(cfg, kubeConfig)
	if err != nil {
		log.Fatal(errors.Wrap(err, "while building Busola init url"))
	}

	application, err := app.New(cfg, busolaURL, kubeConfig)
	if err != nil {
		log.Fatal(errors.Wrap(err, "while creating new application"))
	}
	appRouter := router.New(application)

	log.Printf("Starting server :%d\n", cfg.Port)

	s := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      appRouter,
		ReadTimeout:  cfg.TimeoutRead,
		WriteTimeout: cfg.TimeoutWrite,
		IdleTimeout:  cfg.TimeoutIdle,
	}

	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(errors.Wrap(err, "while starting server"))
	}
}
