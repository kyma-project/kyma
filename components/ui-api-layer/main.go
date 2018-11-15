package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/signal"
	"github.com/pkg/errors"
	"github.com/rs/cors"
	"github.com/vrischmann/envconfig"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/99designs/gqlgen/handler"
	"github.com/gorilla/websocket"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/remoteenvironment"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/origin"
)

type config struct {
	Host                 string   `envconfig:"default=127.0.0.1"`
	Port                 int      `envconfig:"default=3000"`
	AllowedOrigins       []string `envconfig:"optional"`
	Verbose              bool     `envconfig:"default=false"`
	KubeconfigPath       string   `envconfig:"optional"`
	Content              content.Config
	InformerResyncPeriod time.Duration `envconfig:"default=10m"`
	ServerTimeout        time.Duration `envconfig:"default=10s"`
	RemoteEnvironment    remoteenvironment.Config
}

func main() {
	cfg, err := loadConfig("APP")
	exitOnError(err, "Error while loading app config")
	parseFlags(cfg)

	k8sConfig, err := newRestClientConfig(cfg.KubeconfigPath)
	exitOnError(err, "Error while initializing REST client config")

	resolvers, err := domain.New(k8sConfig, cfg.Content, cfg.RemoteEnvironment, cfg.InformerResyncPeriod)
	exitOnError(err, "Error while creating resolvers")

	stopCh := signal.SetupChannel()
	resolvers.WaitForCacheSync(stopCh)

	executableSchema := gqlschema.NewExecutableSchema(gqlschema.Config{Resolvers: resolvers})
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	runServer(stopCh, addr, cfg.AllowedOrigins, executableSchema)
}

func loadConfig(prefix string) (config, error) {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, prefix)
	return cfg, err
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		glog.Fatal(wrappedError)
	}
}

func parseFlags(cfg config) {
	if cfg.Verbose {
		flag.Set("stderrthreshold", "INFO")
	}
	flag.Parse()
}

func newRestClientConfig(kubeconfigPath string) (*restclient.Config, error) {
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
	return config, nil
}

func runServer(stop <-chan struct{}, addr string, allowedOrigins []string, schema graphql.ExecutableSchema) {
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"*"}
	}

	mux := http.NewServeMux()
	mux.Handle("/", handler.Playground("Dataloader", "/graphql"))
	mux.Handle("/graphql", handler.GraphQL(schema,
		handler.WebsocketUpgrader(websocket.Upgrader{
			CheckOrigin: origin.CheckFn(allowedOrigins),
		}),
	))

	serverHandler := cors.New(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{
			"POST", "GET", "OPTIONS",
		},
		AllowCredentials:   true,
		AllowedHeaders:     []string{"*"},
		OptionsPassthrough: false,
	}).Handler(mux)

	srv := &http.Server{Addr: addr, Handler: serverHandler}

	glog.Infof("Listening on %s", addr)

	go func() {
		<-stop
		// Interrupt signal received - shut down the server
		if err := srv.Shutdown(context.Background()); err != nil {
			glog.Errorf("HTTP server Shutdown: %v", err)
		}
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		glog.Errorf("HTTP server ListenAndServe: %v", err)
	}
}
