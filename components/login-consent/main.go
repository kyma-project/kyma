package main

import (
	"flag"
	"github.com/kyma-project/kyma/components/login-consent/internal/login"
	"net/http"
)

func main() {
	var hydraAddr string
	var hydraPort string

	flag.StringVar(&hydraAddr, "hydra-address", "http://ory-hydra-admin.kyma-system.svc.cluster.local", "Hydra administrative endpoint address")
	flag.StringVar(&hydraPort, "hydra-port", "4445", "Hydra administrative endpoint port")

	flag.Parse()

	hydraCfg := login.NewHydraConfig(hydraAddr, hydraPort)

	m := http.NewServeMux()
	m.HandleFunc("/", hydraCfg.Login)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: m,
	}

	srv.ListenAndServe()
}
