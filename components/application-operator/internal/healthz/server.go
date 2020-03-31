package healthz

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func StartHealthCheckServer(log *logrus.Logger, port string) {
	router := mux.NewRouter()
	router.HandleFunc("/healthz", newHTTPHandler(log))

	server := http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: router,
	}

	log.Info(server.ListenAndServe())
}
