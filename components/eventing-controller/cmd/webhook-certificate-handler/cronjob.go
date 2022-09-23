package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/webhook"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func main() {
	logger := zap.New()
	logger.WithName("webhook-cert-cronjob")
	crdName, ok := os.LookupEnv("CRD_NAME")
	if !ok {
		crdName = "subscriptions.eventing.kyma-project.io"
		logger.Info(fmt.Sprintf("CRD_NAME environment variable wasn't set. Using '%s'", crdName))
	}
	secretName, ok := os.LookupEnv("SECRET_NAME")
	if !ok {
		crdNameWithoutGroup, _, _ := strings.Cut(crdName, ".")
		secretName = crdNameWithoutGroup + "-webhook-server-cert"
		logger.Info(fmt.Sprintf("SECRET_NAME environment variable wasn't set. Using '%s'", secretName))
	}

	client, err := ctrlclient.New(ctrl.GetConfigOrDie(), ctrlclient.Options{})
	if err != nil {
		logger.Error(err, "setup")
		return
	}

	webhookCertHandler := webhook.NewWebhookCertificateHandler(context.Background(), client, &logger, crdName, secretName)
	err = webhookCertHandler.SetupCertificates()
	if err != nil {
		panic(err)
	}
}
