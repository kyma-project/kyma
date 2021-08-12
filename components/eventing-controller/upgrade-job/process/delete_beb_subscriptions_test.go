package process

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/eventing-controller/upgrade-job/processtest"

	"github.com/onsi/gomega"

	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/config"
	bebClientTesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
)

var beb *bebClientTesting.BebMock
var ctrLogger *logger.Logger

// startBebMock starts the beb mock and configures the controller process to use it
func startBebMock(logger *logger.Logger) *bebClientTesting.BebMock {
	bebConfig := &config.Config{}
	beb = bebClientTesting.NewBebMock(bebConfig)
	bebURI := beb.Start()
	logger.Logger.WithContext().Info("beb mock listening at", "address", bebURI)
	tokenURL := fmt.Sprintf("%s%s", bebURI, bebClientTesting.TokenURLPath)
	messagingURL := fmt.Sprintf("%s%s", bebURI, bebClientTesting.MessagingURLPath)
	beb.TokenURL = tokenURL
	beb.MessagingURL = messagingURL
	bebConfig = config.GetDefaultConfig(messagingURL)
	beb.BebConfig = bebConfig
	return beb
}

// TestDeleteBebSubscriptions tests the DeleteBebSubscriptions_DO step
func TestDeleteBebSubscriptions(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	e2eSetup := newE2ESetup()
	cfg := e2eSetup.config

	// Create logger instance
	ctrLogger, err := logger.New(cfg.LogFormat, cfg.LogLevel)
	if err != nil {
		log.Fatalf("failed to initialize logger in testing: %s", err)
	}

	// Setup BEB Mock
	bebMock := startBebMock(ctrLogger)
	bebSecrets, err := processtest.NewBebSecrets(bebMock)
	if err != nil {
		ctrLogger.Logger.WithContext().Error(err)
		os.Exit(1)
	}
	e2eSetup.secrets = bebSecrets

	// Create process
	p := &Process{
		Logger:         ctrLogger.Logger,
		TimeoutPeriod:  60 * time.Second,
		ReleaseName:    cfg.ReleaseName,
		KymaNamespace:  cfg.KymaNamespace,
		ControllerName: cfg.EventingControllerName,
		PublisherName:  cfg.EventingPublisherName,
		State:          State{},
	}

	p.Clients = getProcessClients(e2eSetup, g)

	t.Run("Delete BEB subscriptions", func(t *testing.T) {
		// Now delete the BEB subscriptions
		p.Steps = []Step{
			NewCheckClusterVersion(p),
			NewCheckIsBebEnabled(p),
			NewGetSubscriptions(p),
			NewFilterSubscriptions(p),
			NewDeleteBebSubscriptions(p),
		}
		err := p.Execute()
		g.Expect(err).Should(gomega.BeNil())

		// Check if eventing-publisher-proxy deployment is deleted
		// For each subscription, a single delete request should be received by BEBMock
		for _, sub := range p.State.FilteredSubscriptions.Items {
			g.Expect(countBebDeleteRequests(sub.Name)).Should(gomega.Equal(1))
		}

	})
}

// countBebRequests returns how many delete requests for a given subscription are sent to BEB mock
func countBebDeleteRequests(subscriptionName string) (countDelete int) {
	countDelete = 0
	for req, _ := range beb.Requests {
		switch method := req.Method; method {
		case http.MethodDelete:
			if strings.Contains(req.URL.Path, subscriptionName) {
				countDelete++
			}
		}
	}
	return countDelete
}
