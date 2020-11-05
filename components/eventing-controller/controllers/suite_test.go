package controllers

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	apigatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/ems/api/events/config"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go controllertesting framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

// TODO: make configurable
const (
	useExistingCluster       = false
	attachControlPlaneOutput = false
)

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
var beb *BebMock
var bebConfig *config.Config

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))

	By("bootstrapping test environment")
	useExistingCluster := useExistingCluster
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:        []string{filepath.Join("..", "config", "crd", "bases")},
		AttachControlPlaneOutput: attachControlPlaneOutput,
		UseExistingCluster:       &useExistingCluster,
	}

	var err error

	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	err = eventingv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = apigatewayv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	// +kubebuilder:scaffold:scheme

	bebMock := startBebMock()
	//client, err := client.New()
	// Source: https://book.kubebuilder.io/cronjob-tutorial/writing-tests.html
	syncPeriod := time.Second
	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:     scheme.Scheme,
		SyncPeriod: &syncPeriod,
		//NewClient: ,
	})
	Expect(err).ToNot(HaveOccurred())
	envConf := &env.Config{
		BebApiUrl:                bebMock.MessagingURL,
		ClientID:                 "foo-id",
		ClientSecret:             "foo-secret",
		TokenEndpoint:            bebMock.TokenURL,
		WebhookActivationTimeout: 0,
		WebhookClientID:          "foo-client-id",
		WebhookClientSecret:      "foo-client-secet",
		WebhookTokenEndpoint:     "foo-token-endpoint",
		WebhookAuthType:          "oauth",
		WebhookGrantType:         "client_credentials",
		Domain:                   "domain.com",
	}
	err = NewSubscriptionReconciler(
		k8sManager.GetClient(),
		k8sManager.GetCache(),
		ctrl.Log.WithName("controllers").WithName("Subscription"),
		k8sManager.GetEventRecorderFor("subscription-controller"),
		scheme.Scheme, envConf,
	).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	err = NewAPIRuleReconciler(
		k8sManager.GetClient(),
		k8sManager.GetCache(),
		ctrl.Log.WithName("controllers").WithName("APIRule"),
		k8sManager.GetEventRecorderFor("api-rule-controller"),
		scheme.Scheme,
	).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err = k8sManager.Start(ctrl.SetupSignalHandler())
		Expect(err).ToNot(HaveOccurred())
	}()

	k8sClient = k8sManager.GetClient()
	Expect(k8sClient).ToNot(BeNil())

	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})

// startBebMock starts the beb mock and configures the controller process to use it
func startBebMock() *BebMock {
	By("Preparing BEB Mock")
	bebConfig := &config.Config{}
	beb = NewBebMock(bebConfig)
	bebURI := beb.Start()
	logf.Log.Info("beb mock listening at", "address", bebURI)
	tokenURL := fmt.Sprintf("%s%s", bebURI, TokenURLPath)
	messagingURL := fmt.Sprintf("%s%s", bebURI, MessagingURLPath)
	beb.TokenURL = tokenURL
	beb.MessagingURL = messagingURL
	bebConfig = config.GetDefaultConfig(messagingURL)
	beb.BebConfig = bebConfig
	return beb
}
