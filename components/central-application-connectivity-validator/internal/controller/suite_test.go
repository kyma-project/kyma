package controller_test

import (
	"fmt"
	gocache "github.com/patrickmn/go-cache"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"context"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/kyma-project/kyma/common/logging/logger"
	"github.com/kyma-project/kyma/common/logging/tracing"
	"github.com/kyma-project/kyma/components/central-application-connectivity-validator/internal/controller"
	"github.com/kyma-project/kyma/components/central-application-connectivity-validator/internal/validationproxy"
	"github.com/kyma-project/kyma/components/central-application-gateway/pkg/apis/applicationconnector/v1alpha1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	testEnv        *envtest.Environment
	config         *rest.Config
	k8sClient      client.Client
	suiteCtx       context.Context
	cancelSuiteCtx context.CancelFunc

	idCache                 *gocache.Cache
	appNamePlaceholder      = "%%APP_NAME%%"
	eventingPathPrefixV1    = "/%%APP_NAME%%/v1/events"
	eventingPathPrefixV2    = "/%%APP_NAME%%/v2/events"
	eventingPathPrefix      = "/%%APP_NAME%%/events"
	eventingPublisherHost   = "eventing-event-publisher-proxy.kyma-system"
	eventingDestinationPath = "/publish"
	testProxyServerPort     = "8078"
)

type testTransport struct {
}

func (t testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	responseBody := "eventing-event-publisher-proxy.kyma-system: [OK]"
	respReader := io.NopCloser(strings.NewReader(responseBody))
	resp := http.Response{
		StatusCode:    http.StatusOK,
		Body:          respReader,
		ContentLength: int64(len(responseBody)),
		Header: map[string][]string{
			"Content-Type": {"text/plain"},
		},
	}
	return &resp, nil
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	suiteCtx, cancelSuiteCtx = context.WithCancel(context.Background())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "hack")},
		ErrorIfCRDPathMissing: true,
	}

	var err error
	// config is defined in this file globally.
	config, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(config).NotTo(BeNil())

	err = v1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).To(BeNil())

	k8sClient, err = client.New(config, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	k8sManager, err := ctrl.NewManager(config, ctrl.Options{
		Scheme: scheme.Scheme,
		Cache:  cache.Options{SyncPeriod: pointer.Duration(time.Second * 2)},
	})
	Expect(err).ToNot(HaveOccurred())

	idCache = gocache.New(
		gocache.NoExpiration,
		gocache.NoExpiration,
	)

	log, err := logger.New(logger.TEXT, logger.DEBUG)
	Expect(err).Should(BeNil())

	controller := controller.NewController(
		log,
		k8sClient,
		idCache,
		appNamePlaceholder,
		eventingPathPrefixV1,
		eventingPathPrefixV2,
		eventingPathPrefix,
	)
	err = controller.SetupWithManager(k8sManager)
	Expect(err).To(BeNil())

	ceProxyTransport := &testTransport{}

	proxyHandler := validationproxy.NewProxyHandler(
		eventingPublisherHost,
		eventingDestinationPath,
		idCache,
		log,
		validationproxy.WithCEProxyTransport(ceProxyTransport))

	tracingMiddleware := tracing.NewTracingMiddleware(proxyHandler.ProxyAppConnectorRequests)

	go func() {
		defer GinkgoRecover()

		srv := http.Server{
			Handler: validationproxy.NewHandler(tracingMiddleware),
			Addr:    fmt.Sprintf(":%s", testProxyServerPort),
		}

		defer func() {
			if err := srv.Shutdown(suiteCtx); err != nil {
				logf.Log.Error(err, "while shutting down http server")
			}
		}()

		ln, err := net.Listen("tcp", srv.Addr)
		Expect(err).Should(BeNil())

		err = srv.Serve(ln)
		Expect(err).Should(BeNil())
	}()

	go func() {
		err = k8sManager.Start(suiteCtx)
		Expect(err).ToNot(HaveOccurred(), "failed to run manager")
	}()
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	cancelSuiteCtx()

	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
