package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/kyma-project/kyma/components/busola-migrator/internal/app"

	"github.com/kyma-project/kyma/components/busola-migrator/internal/config"
	"github.com/kyma-project/kyma/components/busola-migrator/internal/router"
)

func main() {
	cfg := config.LoadConfig()

	application := app.New(cfg.BusolaURL)
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
		log.Fatal("Server startup failed")
	}
}
