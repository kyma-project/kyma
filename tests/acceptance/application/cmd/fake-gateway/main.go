package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Bplotka/go-httplog"
	"github.com/Bplotka/go-httplog/logrus"
	"github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
)

type Config struct {
	// ApplicationServerPort is a port used by this gateway to listen on requests incoming from services.
	ApplicationServerPort int `envconfig:"default=8080"`
}

type httpHandler struct {
	Cfg Config
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Application gateway, request URL: \"%s\" request headers: %v\n", r.URL, r.Header)
}

func main() {
	var cfg Config
	if err := envconfig.Init(&cfg); err != nil {
		panic(err)
	}

	l := logrus.New()
	l.Infof("Starting server, port: %d", cfg.ApplicationServerPort)

	h := &httpHandler{
		Cfg: cfg,
	}

	mux := http.NewServeMux()
	mux.Handle("/", h)

	loggedHandler := httplog.RegisterMiddleware(
		httplogrus.ToHTTPFieldLoggerInfo(l),
		httplog.DefaultReqResConfig(),
	)(http.HandlerFunc(mux.ServeHTTP))

	listenOn := fmt.Sprintf(":%d", cfg.ApplicationServerPort)
	httpServer := http.Server{Addr: listenOn, Handler: loggedHandler}

	go func() {
		log.Fatal(httpServer.ListenAndServe())
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	l.Infoln("Shutdown signal received shutting down gracefully...")

	httpServer.Shutdown(context.Background())
}
