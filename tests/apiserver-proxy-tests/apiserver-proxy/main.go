package main

import (
	"fmt"

	"github.com/kyma-project/kyma/tests/apiserver-proxy-tests/apiserver-proxy/internal"
)

func main() {

	config, err := graphql.LoadConfig(graphql.AdminUser)
	if err != nil {
		fmt.Print(err)
	}

	token, err := graphql.Authenticate(config.IdProviderConfig)
	if err != nil {
		fmt.Print(err)
	}
	fmt.Print(token)
}
