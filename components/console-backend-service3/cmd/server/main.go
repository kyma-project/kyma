package main

import (
	"context"
	"github.com/99designs/gqlgen/graphql"
	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/graph/model"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/graph"
	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/graph/generated"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	config, err := newRestClientConfig("/Users/i355395/.kube/config", 2)
	if err != nil {
		panic(err)
	}

	resolver, err := graph.NewResolver(config, 10 * time.Minute)
	if err != nil {
		panic(err)
	}
	resolver.WaitForCacheSync(make(chan struct{}))
	srvConfig := generated.Config{Resolvers: resolver}
	srvConfig.Directives.HasAccess = func(ctx context.Context, obj interface{}, next graphql.Resolver, attributes model.ResourceAttributes) (res interface{}, err error) {
		return next(ctx)
	}
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(srvConfig))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func newRestClientConfig(kubeconfigPath string, burst int) (*restclient.Config, error) {
	var config *restclient.Config
	var err error
	if kubeconfigPath != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	} else {
		config, err = restclient.InClusterConfig()
	}

	if err != nil {
		return nil, err
	}

	config.Burst = burst
	config.UserAgent = "console-backend-service"
	return config, nil
}
