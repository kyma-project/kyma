package webhook

import (
	"fmt"
	"net/http"

	"github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func StartWebhookServer(log *logrus.Logger, port string, applicationInterface v1alpha1.ApplicationInterface) {
	handler := Handler{
		applicationInterface,
	}

	router := mux.NewRouter()
	router.HandleFunc("/mutate", handler.handle)

	server := http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: router,
	}

	certPath := "/keys/webhook.crt"
	keyPath := "/keys/webhook.key"

	log.Info(server.ListenAndServeTLS(certPath, keyPath))
}
