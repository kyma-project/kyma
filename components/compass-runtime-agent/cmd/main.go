package main

import (
	"time"

	log "github.com/sirupsen/logrus"
)

func main() {
	for range time.Tick(time.Second * 30) {
		log.Info("Runtime agent works.")
	}
}
