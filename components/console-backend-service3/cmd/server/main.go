package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/graph/model"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/graph"
	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/graph/generated"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	authenticatorpkg "k8s.io/apiserver/pkg/authentication/authenticator"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/authn"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/authz"
)

type config struct {
	Host                 string        `envconfig:"default=127.0.0.1"`
	Port                 int           `envconfig:"default=3000"`
	AllowedOrigins       []string      `envconfig:"optional"`
	Verbose              bool          `envconfig:"default=false"`
	KubeconfigPath       string        `envconfig:"optional"`
	SystemNamespaces     []string      `envconfig:"default=istio-system;knative-eventing;knative-serving;kube-public;kube-system;kyma-installer;kyma-integration;kyma-system;natss;compass-system"`
	InformerResyncPeriod time.Duration `envconfig:"default=10m"`
	ServerTimeout        time.Duration `envconfig:"default=10s"`
	Burst                int           `envconfig:"default=2"`
}

func main() {

	cfg, developmentMode, err := loadConfig("APP")
	if err != nil {
		panic(err)
	}
	parseFlags(cfg)

	restClientConfig, err := newRestClientConfig(cfg.KubeconfigPath, 2)
	if err != nil {
		panic(err)
	}

	resolver, err := graph.NewResolver(restClientConfig, 10*time.Minute)
	if err != nil {
		panic(err)
	}
	resolver.WaitForCacheSync(make(chan struct{}))
	srvConfig := generated.Config{Resolvers: resolver}
	srvConfig.Directives.HasAccess = func(ctx context.Context, obj interface{}, next graphql.Resolver, attributes model.ResourceAttributes) (res interface{}, err error) {
		return next(ctx)
	}

	var authenticator authenticatorpkg.Request
	if !developmentMode {
		authenticator, err = authn.NewOIDCAuthenticator(&cfg.OIDC)
		exitOnError(err, "Error while creating OIDC authenticator")
		sarClient := kubeClient.AuthorizationV1beta1().SubjectAccessReviews()
		authorizer, err := authz.NewAuthorizer(sarClient, cfg.SARCacheConfig)
		exitOnError(err, "Failed to create authorizer")

		gqlCfg.Directives.HasAccess = authz.NewRBACDirective(authorizer, kubeClient.Discovery())
	}
	
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(srvConfig))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	bind := fmt.Sprintf("%s:%v", cfg.Host,cfg.Port)
	log.Printf("connect to http://%s/ for GraphQL playground", bind)
	log.Fatal(http.ListenAndServe(bind, nil))
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

func loadConfig(prefix string) (config, bool, error) {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, prefix)
	if err != nil {
		return cfg, false, err
	}

	developmentMode := cfg.KubeconfigPath != ""

	return cfg, developmentMode, nil
}

func parseFlags(cfg config) {
	if cfg.Verbose {
		err := flag.Set("stderrthreshold", "INFO")
		if err != nil {
			glog.Error(errors.Wrap(err, "while parsing flags"))
		}
	}
	flag.Parse()
}
