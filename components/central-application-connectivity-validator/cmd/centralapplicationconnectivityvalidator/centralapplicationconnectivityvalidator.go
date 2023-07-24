package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/kyma-project/kyma/common/logging/logger"
	"github.com/kyma-project/kyma/common/logging/tracing"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apis/applicationconnector/v1alpha1"

	"github.com/kyma-project/kyma/components/central-application-connectivity-validator/internal/controller"
	"github.com/kyma-project/kyma/components/central-application-connectivity-validator/internal/externalapi"
	"github.com/kyma-project/kyma/components/central-application-connectivity-validator/internal/validationproxy"
	"github.com/oklog/run"
	"github.com/patrickmn/go-cache"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	shutdownTimeout = 2 * time.Second
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
}

func main() {
	options, err := parseOptions()
	if err != nil {
		if logErr := logger.LogFatalError("Failed to parse options: %s", err.Error()); logErr != nil {
			fmt.Printf("Failed to initializie default fatal error logger: %s,Failed to parse options: %s", logErr, err)
		}
		os.Exit(1)
	}
	if err = options.validate(); err != nil {
		if logErr := logger.LogFatalError("Failed to validate options: %s", err.Error()); logErr != nil {
			fmt.Printf("Failed to initializie default fatal error logger: %s,Failed to validate options: %s", logErr, err)
		}
		os.Exit(1)
	}
	level, err := logger.MapLevel(options.LogLevel)
	if err != nil {
		if logErr := logger.LogFatalError("Failed to map log level from options: %s", err.Error()); logErr != nil {
			fmt.Printf("Failed to initializie default fatal error logger: %s, Failed to map log level from options: %s", logErr, err)
		}

		os.Exit(2)
	}
	format, err := logger.MapFormat(options.LogFormat)
	if err != nil {
		if logErr := logger.LogFatalError("Failed to map log format from options: %s", err.Error()); logErr != nil {
			fmt.Printf("Failed to initializie default fatal error logger: %s, Failed to map log format from options: %s", logErr, err)
		}
		os.Exit(3)
	}

	log, err := logger.New(format, level)
	if err != nil {
		if logErr := logger.LogFatalError("Failed to initialize logger: %s", err.Error()); logErr != nil {
			fmt.Printf("Failed to initializie default fatal error logger: %s, Failed to initialize logger: %s", logErr, err)
		}
		os.Exit(4)
	}
	if err := logger.InitKlog(log, level); err != nil {
		log.WithContext().Error("While initializing klog logger: %s", err.Error())
		os.Exit(5)
	}

	log.WithContext().With("options", options).Info("Starting Validation Proxy.")

	idCache := cache.New(
		time.Duration(options.cacheExpirationSeconds)*time.Second,
		time.Duration(options.cacheCleanupIntervalSeconds)*time.Second,
	)
	idCache.OnEvicted(func(key string, i interface{}) {
		log.WithContext().
			With("controller", "cache_janitor").
			With("name", key).
			Warnf("Deleted the application from the cache with values %v.", i)
	})

	proxyHandler := validationproxy.NewProxyHandler(
		options.eventingPublisherHost,
		options.eventingDestinationPath,
		idCache,
		log)

	tracingMiddleware := tracing.NewTracingMiddleware(proxyHandler.ProxyAppConnectorRequests)

	proxyServer := http.Server{
		Handler: validationproxy.NewHandler(tracingMiddleware),
		Addr:    fmt.Sprintf(":%d", options.proxyPort),
	}

	externalServer := http.Server{
		Handler: externalapi.NewHandler(),
		Addr:    fmt.Sprintf(":%d", options.externalAPIPort),
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: "0",
		SyncPeriod:         &options.syncPeriod,
		ClientDisableCacheFor: []client.Object{
			&v1alpha1.Application{},
		},
	})
	if err != nil {
		log.WithContext().Error("Unable to start manager: %s", err.Error())
		os.Exit(1)
	}
	if err = controller.NewController(log, mgr.GetClient(), idCache, options.appNamePlaceholder, options.eventingPathPrefixV1, options.eventingPathPrefixV2, options.eventingPathPrefixEvents).SetupWithManager(mgr); err != nil {
		log.WithContext().Error("Unable to create reconciler: %s", err.Error())
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	var g run.Group
	addInterruptSignalToRunGroup(ctx, cancel, log, &g)
	addManagerToRunGroup(ctx, log, &g, mgr)
	addHttpServerToRunGroup(log, "proxy-server", &g, &proxyServer)
	addHttpServerToRunGroup(log, "external-server", &g, &externalServer)

	err = g.Run()
	if err != nil && err != http.ErrServerClosed {
		log.WithContext().Fatal(err)
	}
}

func addHttpServerToRunGroup(log *logger.Logger, name string, g *run.Group, srv *http.Server) {
	log.WithContext().Infof("Starting %s HTTP server on %s", name, srv.Addr)
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		log.WithContext().Fatalf("Unable to start %s HTTP server: '%s'", name, err.Error())
	}
	g.Add(func() error {
		defer log.WithContext().Infof("Server %s finished", name)
		return srv.Serve(ln)
	}, func(error) {
		log.WithContext().Infof("Shutting down %s HTTP server on %s", name, srv.Addr)

		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		err = srv.Shutdown(ctx)
		if err != nil && err != http.ErrServerClosed {
			log.WithContext().Warnf("HTTP server shutdown %s failed: %s", name, err.Error())
		}
	})
}

func addManagerToRunGroup(ctx context.Context, log *logger.Logger, g *run.Group, mgr manager.Manager) {
	g.Add(func() error {
		defer log.WithContext().Infof("Manager finished")
		return mgr.Start(ctx)
	}, func(error) {
	})
}

func addInterruptSignalToRunGroup(ctx context.Context, cancel context.CancelFunc, log *logger.Logger, g *run.Group) {
	g.Add(func() error {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-ctx.Done():
		case sig := <-c:
			log.WithContext().Infof("received signal %s", sig)
		}
		return nil
	}, func(error) {
		cancel()
	})
}
