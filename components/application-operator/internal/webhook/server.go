package webhook

import (
	"fmt"
	"github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	"net/http"
	"os"

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

	certPath := "/tmp/webhook.pem"
	keyPath := "/tmp/key.pem"

	certFile, err := os.Create(certPath)
	if err != nil {
		log.Error(err.Error())
		return
	}

	lines, err := certFile.WriteString(cert)
	if err != nil {
		log.Error(err.Error())
		return
	}
	log.Info(lines)

	keyFile, err := os.Create(keyPath)
	if err != nil {
		log.Error(err.Error())
		return
	}

	lines, err = keyFile.WriteString(key)

	if err != nil {
		log.Error(err.Error())
		return
	}
	log.Info(lines)

	log.Info(server.ListenAndServeTLS(certPath, keyPath))
}
