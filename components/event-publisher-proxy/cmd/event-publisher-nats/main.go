package main

import (
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func main() {
	logger := logrus.New()
	logger.Info("Event Publisher NATS Started")

	// wait for shutdown signal
	shutdown := make(chan os.Signal)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown
	close(shutdown)

	logger.Info("Event Publisher NATS Shutdown")
}
