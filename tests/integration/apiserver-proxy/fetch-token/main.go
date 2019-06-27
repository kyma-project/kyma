package main

import (
	"fmt"
	"log"

	"github.com/kyma-project/kyma/tests/integration/apiserver-proxy/fetch-token/internal"
)

func main() {

	config, err := internal.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	token, err := internal.Authenticate(config.IdProviderConfig)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(token)
}
