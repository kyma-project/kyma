package app

import (
	"context"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/handler"
	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/authn"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/authz"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/origin"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/tracing"
	"github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
	"github.com/pkg/errors"
	"github.com/rs/cors"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net"
	"net/http"
	"strconv"
)

func Run(listener net.Listener, stopCh <-chan struct{}, cfg Config, k8sConfig *rest.Config, authRequest authenticator.Request) error {
	resolvers, err := domain.New(k8sConfig, cfg.Application, cfg.AssetStore, cfg.InformerResyncPeriod, cfg.FeatureToggles)
	if err != nil {
		return errors.Wrap(err, "Error while creating resolvers")
	}

	kubeClient, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return errors.Wrap(err, "Failed to instantiate Kubernetes client")
	}

	gqlCfg := gqlschema.Config{Resolvers: resolvers}
	if authRequest != nil {
		sarClient := kubeClient.AuthorizationV1beta1().SubjectAccessReviews()
		authorizer, err := authz.NewAuthorizer(sarClient, cfg.SARCacheConfig)
		if err != nil {
			return errors.Wrap(err, "Failed to create authorizer")
		}

		gqlCfg.Directives.HasAccess = authz.NewRBACDirective(authorizer, kubeClient.Discovery())
	}

	resolvers.WaitForCacheSync(stopCh)
	executableSchema := gqlschema.NewExecutableSchema(gqlCfg)
	return runServer(listener, stopCh, cfg, executableSchema, authRequest)
}

func runServer(listener net.Listener, stop <-chan struct{}, cfg Config, schema graphql.ExecutableSchema, authRequest authenticator.Request) error {
	err := setupTracing(cfg.Tracing, cfg.Port)
	if err != nil {
		return err
	}
	var allowedOrigins []string
	if len(cfg.AllowedOrigins) == 0 {
		allowedOrigins = []string{"*"}
	} else {
		allowedOrigins = cfg.AllowedOrigins
	}

	router := mux.NewRouter()

	if authRequest != nil {
		router.Use(authn.AuthMiddleware(authRequest))
	}
	router.HandleFunc("/", handler.Playground("Dataloader", "/graphql"))
	graphQLHandler := handler.GraphQL(schema,
		handler.WebsocketUpgrader(websocket.Upgrader{
			CheckOrigin: origin.CheckFn(allowedOrigins),
		}),
		handler.Tracer(tracing.New()),
	)
	router.HandleFunc("/graphql", tracing.NewWithParentSpan(cfg.Tracing.ServiceSpanName, graphQLHandler))
	serverHandler := cors.New(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{
			"POST", "GET", "OPTIONS",
		},
		AllowCredentials:   true,
		AllowedHeaders:     []string{"*"},
		OptionsPassthrough: false,
	}).Handler(router)

	srv := &http.Server{Handler: serverHandler}

	go func() {
		<-stop
		// Interrupt signal received - shut down the server
		if err := srv.Shutdown(context.Background()); err != nil {
			glog.Errorf("HTTP server Shutdown: %v", err)
		}
	}()

	if err := srv.Serve(listener); err != http.ErrServerClosed {
		glog.Errorf("HTTP server ListenAndServe: %v", err)
	}
	return nil
}

func setupTracing(cfg tracing.Config, hostPort int) error {
	collector, err := zipkin.NewHTTPCollector(cfg.CollectorUrl)
	if err != nil {
		return errors.Wrap(err, "Error while initializing zipkin")
	}
	recorder := zipkin.NewRecorder(collector, cfg.Debug, strconv.Itoa(hostPort), cfg.ServiceSpanName)
	tracer, err := zipkin.NewTracer(recorder, zipkin.TraceID128Bit(false))
	if err != nil {
		return errors.Wrap(err, "Error while initializing tracer")
	}
	opentracing.SetGlobalTracer(tracer)
	return nil
}
